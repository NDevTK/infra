// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	"errors"
	"fmt"
	"sync"

	"infra/libs/cipkg_new/core"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// PackageDependency is the dependency of a package.
type PackageDependency struct {
	Package Package
	Runtime bool
}

// Package represents a high-level package including:
// - A unique package id.
// - A storage handler for the package.
// - The transformed derivation and its unique id.
// - All its dependencies.
// - The action responsible for the package's content.
type Package struct {
	PackageID string
	Handler   core.PackageHandler

	DerivationID string
	Derivation   *core.Derivation

	Dependencies []PackageDependency

	Action *core.Action
}

var (
	ErrUnsupportedAction = errors.New("unsupported action spec")
)

// Transformer is the function transforms an action specification to a
// derivation.
type Transformer[M proto.Message] func(M, []PackageDependency) (*core.Derivation, error)

// ActionProcessor processes and transforms actions into packages.
type ActionProcessor struct {
	buildPlatform string
	packages      core.PackageManager
	transformers  map[protoreflect.FullName]Transformer[proto.Message]

	mu     sync.Mutex
	sealed bool
}

func NewActionProcessor(buildPlat string, pm core.PackageManager) *ActionProcessor {
	ap := &ActionProcessor{
		buildPlatform: buildPlat,
		packages:      pm,
		transformers:  make(map[protoreflect.FullName]Transformer[proto.Message]),
	}
	MustSetTransformer[*core.ActionCommand](ap, ActionCommandTransformer)
	MustSetTransformer[*core.ActionReexec](ap, ActionReexecTransformer)
	MustSetTransformer[*core.ActionURLFetch](ap, ActionURLFetchTransformer)
	MustSetTransformer[*core.ActionFilesCopy](ap, ActionFilesCopyTransformer)
	MustSetTransformer[*core.ActionCIPDExport](ap, ActionCIPDExportTransformer)
	return ap
}

var (
	ErrTransformerExisted    = errors.New("transformer for the message type already existed")
	ErrActionProcessorSealed = errors.New("transformer can't be set after processor being used")
)

// SetTransformer set the transformer for the action specification M.
func SetTransformer[M proto.Message](ap *ActionProcessor, tf Transformer[M]) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	if ap.sealed {
		return ErrActionProcessorSealed
	}

	var m M
	name := proto.MessageName(m)
	if _, ok := ap.transformers[name]; ok {
		return ErrTransformerExisted
	}
	ap.transformers[name] = func(m proto.Message, deps []PackageDependency) (*core.Derivation, error) {
		return tf(m.(M), deps)
	}
	return nil
}

// MustSetTransformer set the transformer for the action specification M and
// panic if any error happened.
func MustSetTransformer[M proto.Message](ap *ActionProcessor, tf Transformer[M]) {
	if err := SetTransformer[M](ap, tf); err != nil {
		panic(err)
	}
}

func (ap *ActionProcessor) Process(a *core.Action) (Package, error) {
	ap.mu.Lock()
	if !ap.sealed {
		ap.sealed = true
	}
	ap.mu.Unlock()

	if a.Metadata == nil {
		return Package{}, fmt.Errorf("action name can't be empty: %s", a.Spec)
	}

	pkg := Package{
		Action: a,
	}

	// Recursivedly process all dependencies
	for _, d := range a.Deps {
		dpkg, err := ap.Process(d.Action)
		if err != nil {
			return Package{}, err
		}

		pkg.Dependencies = append(pkg.Dependencies, PackageDependency{
			Package: dpkg,
			Runtime: d.Runtime,
		})
	}

	// Parse spec message
	var (
		spec proto.Message
		err  error
	)

	if s, ok := a.Spec.(*core.Action_Extension); ok {
		if spec, err = s.Extension.UnmarshalNew(); err != nil {
			return Package{}, fmt.Errorf("%s: %w", a.Spec, err)
		}
	} else {
		rm := a.ProtoReflect()
		fds := rm.Descriptor().Oneofs().ByName("spec")
		fd := rm.WhichOneof(fds)
		if fd == nil {
			return Package{}, fmt.Errorf("%s: no action spec available: %w", a.Spec, ErrUnsupportedAction)
		}
		spec = rm.Get(fd).Message().Interface()
	}

	// Transform spec message
	tr := ap.transformers[proto.MessageName(spec)]
	if tr == nil {
		return Package{}, ErrUnsupportedAction
	}
	drv, err := tr(spec, pkg.Dependencies)
	if err != nil {
		return Package{}, fmt.Errorf("%s: %w", a.Spec, err)
	}

	// Update common derivation fields
	drv.Platform = ap.buildPlatform
	for _, d := range pkg.Dependencies {
		drv.Inputs = append(drv.Inputs, d.Package.DerivationID)
	}
	pkg.Derivation = drv

	// Calculate DerivationID
	drvID, err := core.GetDerivationID(pkg.Derivation)
	if err != nil {
		return Package{}, err
	}
	pkg.DerivationID = drvID

	// Calculate PackageID
	pkgID := fmt.Sprintf("%s-%s", a.Metadata.Name, drvID)
	pkg.PackageID = pkgID
	pkg.Handler = ap.packages.Get(pkgID)

	return pkg, nil
}
