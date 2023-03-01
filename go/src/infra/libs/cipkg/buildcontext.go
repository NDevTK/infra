// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cipkg

import (
	"context"
	"time"
)

// BuildContext provides context for the build environment.
// Platform: The cross-compile tuple for the build.
// Storage: The package storage for add and get packages.
// TODO: Maybe change to GenerationContext since it's not used in build?
type BuildContext struct {
	Platforms Platforms
	Packages  PackageManager
	context.Context
}

// WithPlatform returns a BuildContext with Platforms replaced by the argument.
// This is useful for building dependencies based on their dependency types.
func (ctx *BuildContext) WithPlatform(plats Platforms) *BuildContext {
	return &BuildContext{
		Platforms: plats,
		Packages:  ctx.Packages,
		Context:   ctx,
	}
}

// Platforms is the cross-compile platform tuple.
type Platforms struct {
	Build  Platform
	Host   Platform
	Target Platform
}

// Platform includes key-value pairs that represent the platform.
// The minimal platform should at least include os and arch.
type Platform interface {
	OS() string
	Arch() string
	Get(key string) string
	String() string
}

// PackageManager represents the management interface for packages. Generator relies on
// storage to provide a place to store and retrieve packages.
type PackageManager interface {
	// Get(id) returns the handler for the package. It won't try to make package
	// available in the storage and not promising the returned package is valid.
	// Get a package added by Add(...) with valid derivation will always return a
	// valid package.
	Get(id string) Package

	// Add(drv, metadata) returns a valid package if derivation is valid. Whether
	// the derivation will be persisted when it's added depends on storage's
	// implementation.
	Add(drv Derivation, metadata PackageMetadata) Package
}

// PackageMetadata includes all the information for managing packages.
type PackageMetadata struct {
	// Version of the package
	Version string

	// Runtime dependencies for the package.
	Dependencies []string

	// Key for caching the package. It depends on storage to interpret the key.
	CacheKey string
}

type PackageStatus struct {
	Available bool
	LastUsed  time.Time
}

// Package is the interface for a package in the storage. The content of a package
// can be built by calling Build(func(Package) error) error, which will make
// package available if successful and can be referenced by other packages.
// Package shouldn't be modified after build. The implementation should ensure
// TryRemove failed if there is any reference to the package.
type Package interface {
	// Derivation() returns the derivation of the Package. For ill-formed
	// packages (package is retrieved without adding the derivation), calling
	// Derivation() will PANIC. We either need to know how to build the
	// package (require derivation) or we have the ID and the package is
	// available in someplace where we can retrieve it by ID. In any cases,
	// trying to get Derivation() from an ill-formed package is a fatal error.
	Derivation() Derivation

	// Metadata() returns the metadata of the Package. For ill-formed
	// packages (package is retrieved without adding the derivation), calling
	// Metadata() will PANIC. Storage may modify the added metadata based on its
	// implementation.
	Metadata() PackageMetadata

	// Directory() returns the output directory.
	Directory() string

	// Build(buildFunc) makes packages available in the storage.
	// It's responsible for:
	// - Ensure build only happens once unless the package is removed.
	// - Check the remote cache server (if possible).
	// - Set up the build environment (e.g. create output directory).
	// - Mark package available if build function successfully returns.
	// Calling build function is expected to trigger the actual build.
	Build(func(Package) error) error

	// Status() returns the status for the package.
	Status() PackageStatus

	// TryRemove(), IncRef(), DecRef() are the interface for removable packages.
	// If removing package is not supported by PackageManager, TryRemove() will
	// always return false.
	// IncRef() and DecRef() references/dereferences the package to prevent
	// package from being removed, thus they can be no-op if the removal never
	// happens.

	// TryRemove() may remove the package if there is no reference to it.
	TryRemove() (ok bool, err error)

	// Reference the package to prevent it from being removed while in use.
	// IncRef() updates the last access time of the package as well.
	// IncRef() only succeeds if the package is available.
	IncRef() error
	DecRef() error
}
