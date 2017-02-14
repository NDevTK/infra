// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package cipd

import (
	"bytes"
	"fmt"
	"io"

	"infra/tools/vpython/api/env"
	"infra/tools/vpython/venv"

	"github.com/luci/luci-go/cipd/client/cipd"
	"github.com/luci/luci-go/cipd/client/cipd/common"
	"github.com/luci/luci-go/cipd/client/cipd/ensure"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/logging"

	"golang.org/x/net/context"
)

// PackageLoader is an implementation of venv.PackageLoader that uses the
// CIPD service to fetch packages.
//
// Packages that use the CIPD loader use the CIPD package name as their Path
// and a CIPD version/tag/ref as their Version.
type PackageLoader struct {
	// Options are additional client options to use when generating CIPD clients.
	Options cipd.ClientOptions
}

var _ venv.PackageLoader = (*PackageLoader)(nil)

// Resolve implements venv.PackageLoader.
//
// The resulting packages slice will be updated in-place with the resolves
// package name and instance ID.
func (pl *PackageLoader) Resolve(c context.Context, root string, packages []*env.Spec_Package) error {
	if len(packages) == 0 {
		return nil
	}

	var ensureFile bytes.Buffer
	if err := writeEnsureFile(&ensureFile, packages); err != nil {
		return errors.Annotate(err).Reason("failed to generate manifest").Err()
	}

	// Generate a CIPD client. Use the supplied root.
	opts := pl.Options
	opts.Root = root
	client, err := cipd.NewClient(opts)
	if err != nil {
		return errors.Annotate(err).Reason("failed to generate CIPD client").Err()
	}

	// Start a CIPD client batch.
	client.BeginBatch(c)
	defer client.EndBatch(c)

	// Parse and resolve the CIPD ensure file.
	logging.Debugf(c, "Resolving CIPD manifest:\n%s", ensureFile.Bytes())
	ef, err := ensure.ParseFile(&ensureFile)
	if err != nil {
		return errors.Annotate(err).Reason("failed to process ensure file").Err()
	}

	resolved, err := ef.Resolve(func(pkg, vers string) (common.Pin, error) {
		return client.ResolveVersion(c, pkg, vers)
	})
	if err != nil {
		return errors.Annotate(err).Reason("failed to resolve ensure file").Err()
	}

	// Write the results to "packages". All of them should have been installed
	// into the root subdir.
	for i, pkg := range resolved.PackagesBySubdir[""] {
		packages[i].Path = pkg.PackageName
		packages[i].Version = pkg.InstanceID
	}
	return nil
}

// Ensure implement venv.PackageLoader.
//
// The packages must be valid (PackageIsComplete). If they aren't, Ensure will
// panic.
//
// The CIPD client that is used for the operation is generated from the supplied
// options, opts.
func (pl *PackageLoader) Ensure(c context.Context, root string, packages []*env.Spec_Package) error {
	pins, err := packagesToPins(packages)
	if err != nil {
		return errors.Annotate(err).Reason("failed to convert packages to CIPD pins").Err()
	}
	pinSlice := common.PinSliceBySubdir{
		"": pins,
	}

	// Generate a CIPD client. Use the supplied root.
	opts := pl.Options
	opts.Root = root
	client, err := cipd.NewClient(opts)
	if err != nil {
		return errors.Annotate(err).Reason("failed to generate CIPD client").Err()
	}

	// Start a CIPD client batch.
	client.BeginBatch(c)
	defer client.EndBatch(c)

	actions, err := client.EnsurePackages(c, pinSlice, false)
	if err != nil {
		return errors.Annotate(err).Reason("failed to install CIPD packages").Err()
	}
	if len(actions.Errors) > 0 {
		for _, err := range actions.Errors {
			logging.Errorf(c, "CIPD package [%s] error: %s", err.Pin, err.Action)
		}
		return errors.New("CIPD package installation encountered errors")
	}
	return nil
}

func writeEnsureFile(out io.Writer, packages []*env.Spec_Package) error {
	for _, pkg := range packages {
		if err := validatePackage(pkg); err != nil {
			panic(errors.Annotate(err).Reason("invalid CIPD package").Err())
		}
		if _, err := fmt.Fprintf(out, "%s %s\n", pkg.Path, pkg.Version); err != nil {
			return errors.Annotate(err).Reason("failed to write manifest line").Err()
		}
	}
	return nil
}

// validatePackage returns an error if the package does not have all of the
// required fields to describe a CIPD package.
func validatePackage(pkg *env.Spec_Package) error {
	switch {
	case pkg.Path == "":
		return errors.New("package must have a path")
	case pkg.Version == "":
		return errors.New("package must have a version")
	default:
		return nil
	}
}

func packagesToPins(packages []*env.Spec_Package) ([]common.Pin, error) {
	pins := make([]common.Pin, len(packages))
	for i, pkg := range packages {
		pins[i] = common.Pin{
			PackageName: pkg.Path,
			InstanceID:  pkg.Version,
		}
	}
	return pins, nil
}
