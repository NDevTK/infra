// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package core

import (
	"errors"
)

// PackageManager represents the management interface for packages. Generator
// relies on storage to provide a place to store and retrieve packages.
type PackageManager interface {
	// Get(id) returns the handler for the package.
	Get(id string) PackageHandler
}

var (
	ErrPackageNotExist = errors.New("package does not exist")
)

// PackageHandler is the interface for a handler in the storage. The content of
// a package can be built by calling Build(func(Package) error) error, which
// will make package available if successful and can be referenced by other
// packages. Package shouldn't be modified after build.
type PackageHandler interface {
	// OutputDirectory() returns the output directory.
	OutputDirectory() string

	// LoggingDirectory() returns the logging directory.
	LoggingDirectory() string

	// Build(buildFunc) makes packages available in the storage.
	// It's responsible for:
	// - Hold exclusive lock to the package during the build.
	// - Ensure build only happens once unless the package is removed.
	// - Check the remote cache server (if possible).
	// - Set up the build environment (e.g. create output directory).
	// - Mark package available if build function successfully returns.
	// Calling build function is expected to trigger the actual build.
	Build(builder func() error) error

	// TryRemove(), IncRef(), DecRef() are the interface for removable packages.
	// If removing package is not supported by PackageManager, TryRemove() will
	// always return false.
	// IncRef() and DecRef() references/dereferences the package to prevent
	// package from being removed, thus they can be no-op if the removal never
	// happens.

	// TryRemove() may remove the package only if there is no reference to it.
	TryRemove() (ok bool, err error)

	// Reference the package to prevent it from being removed while in use.
	// IncRef() updates the last access time of the package as well.
	// IncRef() only succeeds if the package is available.
	// Otherwise ErrPackageNotExist will be returned.
	IncRef() error
	DecRef() error
}
