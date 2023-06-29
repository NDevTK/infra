// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package base

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"infra/libs/cipkg_new/base/actions"
	"infra/libs/cipkg_new/base/generators"
	"infra/libs/cipkg_new/core"

	"go.chromium.org/luci/common/logging"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type PreExecuteHook func(ctx context.Context, pkg actions.Package) error

// ExecutionConfig includes all configs for Executor.
type ExecutionConfig struct {
	OutputDir  string
	WorkingDir string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Executor func(ctx context.Context, cfg *ExecutionConfig, drv *core.Derivation) error

// Execute is the default Executor which runs the command presented by the
// derivation.
func Execute(ctx context.Context, cfg *ExecutionConfig, drv *core.Derivation) error {
	cmd := exec.CommandContext(ctx, drv.Args[0], drv.Args[1:]...)
	cmd.Path = drv.Args[0]
	cmd.Dir = cfg.WorkingDir
	cmd.Stdin = cfg.Stdin
	cmd.Stdout = cfg.Stdout
	cmd.Stderr = cfg.Stderr
	cmd.Env = append([]string{
		"out=" + cfg.OutputDir,
	}, drv.Env...)
	return cmd.Run()
}

// ExecutionPlan is the plan for packages to be built by running executor.
// It will ensure all packages' dependencies added into the plan and will be
// available when the package is built.
type ExecutionPlan struct {
	newPkgs    []actions.Package // New packages planned to be built.
	availables []actions.Package // All available packages will be referenced.
	added      map[string]struct{}
}

// NewExecutionPlan generates an execution plan for building packages to make
// all of them and their dependencies available. A preExecFn can be provided to
// e.g. fetch package from local or remote cache before the package being added
// to the plan.
func NewExecutionPlan(ctx context.Context, pkgs []actions.Package, preExecFn PreExecuteHook) (*ExecutionPlan, error) {
	p := &ExecutionPlan{
		added: make(map[string]struct{}),
	}
	for _, pkg := range pkgs {
		if err := p.add(ctx, pkg, preExecFn); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p *ExecutionPlan) add(ctx context.Context, pkg actions.Package, preExecFn PreExecuteHook) error {
	if _, ok := p.added[pkg.PackageID]; ok {
		return nil
	}

	switch err := pkg.Handler.IncRef(); {
	case errors.Is(err, core.ErrPackageNotExist):
		if preExecFn != nil {
			if err := preExecFn(ctx, pkg); err != nil {
				return fmt.Errorf("failed to run preExecute hook for the package: %s: %w", pkg.PackageID, err)
			}
			return p.add(ctx, pkg, nil)
		}

		p.added[pkg.PackageID] = struct{}{}
		for _, d := range pkg.Dependencies {
			if err := p.add(ctx, d.Package, preExecFn); err != nil {
				return err
			}
		}
		p.newPkgs = append(p.newPkgs, pkg)
		return nil
	case err == nil:
		p.added[pkg.PackageID] = struct{}{}
		for _, d := range pkg.Dependencies {
			if d.Runtime {
				if err := p.add(ctx, d.Package, preExecFn); err != nil {
					return err
				}
			}
		}
		p.availables = append(p.availables, pkg)
		return nil
	default:
		return err
	}
}

// Execute executes packages' derivations added to the plan and all their
// dependencies. All packages will be dereferenced after the build. Leave
// it to the user to decide those of which packages will be used at the
// runtime. There may be a chance that a package is removed during the short
// amount of time. But since IncRef will update the last accessed timestam,
// this is highly unlikely. And even if it's happened, we can retry the process.
func (p *ExecutionPlan) Execute(ctx context.Context, tempDir string, execFn Executor) (err error) {
	for _, pkg := range p.newPkgs {
		if err := pkg.Handler.Build(func() error {
			if err := dumpProto(pkg.Action, pkg.Handler.LoggingDirectory(), "action.pb"); err != nil {
				return err
			}
			if err := dumpProto(pkg.Derivation, pkg.Handler.LoggingDirectory(), "derivation.pb"); err != nil {
				return err
			}

			logging.Infof(ctx, "build package %s", pkg.PackageID)
			d, err := os.MkdirTemp(tempDir, fmt.Sprintf("%s-", pkg.PackageID))
			if err != nil {
				return err
			}

			var out strings.Builder
			cfg := &ExecutionConfig{
				OutputDir:  pkg.Handler.OutputDirectory(),
				WorkingDir: d,
				Stdout:     &out,
				Stderr:     &out,
			}
			if err := execFn(ctx, cfg, pkg.Derivation); err != nil {
				logging.Errorf(ctx, "\n%s\n", out.String())
				return err
			}
			logging.Debugf(ctx, "\n%s", out.String())
			return nil
		}); err != nil {
			return fmt.Errorf("failed to build package: %s: %w", pkg.PackageID, err)
		}
		if err := pkg.Handler.IncRef(); err != nil {
			return fmt.Errorf("failed to reference the package: %s: %w", pkg.PackageID, err)
		}
		p.availables = append(p.availables, pkg)
	}

	return
}

func dumpProto(m protoreflect.ProtoMessage, dir, name string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := protojson.MarshalOptions{Multiline: true}.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		return err
	}
	return nil
}

type Builder struct {
	platforms generators.Platforms
	packages  core.PackageManager
	processor *actions.ActionProcessor
}

func NewBuilder(plats generators.Platforms, pm core.PackageManager) *Builder {
	return &Builder{
		platforms: plats,
		packages:  pm,
		processor: actions.NewActionProcessor(plats.Build.String(), pm),
	}
}

// Generate converts generators to actions.
func (b *Builder) Generate(ctx context.Context, gs []generators.Generator) ([]*core.Action, error) {
	var acts []*core.Action
	for _, g := range gs {
		a, err := g.Generate(ctx, b.platforms)
		if err != nil {
			return nil, err
		}

		acts = append(acts, a)
	}

	return acts, nil
}

// Process converts actions to packages with derivation.
func (b *Builder) Process(ctx context.Context, acts []*core.Action) ([]actions.Package, error) {
	var pkgs []actions.Package
	for _, a := range acts {
		pkg, err := b.processor.Process(a)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}
