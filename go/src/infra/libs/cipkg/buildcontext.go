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
	Storage   Storage
	context.Context
}

func (ctx *BuildContext) WithPlatform(plats Platforms) *BuildContext {
	return &BuildContext{
		Platforms: plats,
		Storage:   ctx.Storage,
		Context:   ctx,
	}
}

// Cross-compile platform tuple.
type Platforms struct {
	Build  Platform
	Host   Platform
	Target Platform
}

type Platform interface {
	OS() string
	Arch() string
	Get(key string) string
	String() string
}

// Storage represents the management interface for packages. Generator relies on
// storage to provide a place to store and retrieve packages.
type Storage interface {
	// Get(id) returns the handler for the package. It won't try to make package
	// available in the storage and not promising the returned package is valid.
	// Get a package added by Add(...) with valid derivation will always return a
	// valid package.
	Get(id string) Package

	// Add(drv, metadata) returns a valid package if derivation is valid. Whether
	// the derivation will be persisted when it's added depends on storage's
	// implementation.
	Add(drv Derivation, metadata PackageMetadata) Package

	// Prune(ctx, ttl, max) removes at most ${max} packages from storage which
	// haven't been used in the past ${ttl}.
	Prune(ctx context.Context, ttl time.Duration, max int)
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

// Package is the interface for a package in the storage. The content of a package
// can be built by calling Build(func(Package) error) error, which will make
// package available if successful and can be referenced by other packages.
// Package shouldn't be modified after build. Only read lock is provided in the
// interface, but the implementation should ensure it's exclusively locked during
// the build.
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
	// - Hold the exclusive lock of the package during the build.
	// - Check remote cache server (if possible).
	// - Set up the build environment (e.g. create output directory).
	// - Mark package available if build function successfully returns.
	// Calling build function is expected to trigger the actual build.
	Build(func(Package) error) error

	// TryRemove() removes the package if it's not locked.
	TryRemove() (ok bool, err error)

	// Available() checks whether the Package is available in the storage and
	// return the last time it's used. It only checks the completion stamp in
	// most cases.
	Available() (ok bool, mtime time.Time)

	// Lock the package to prevent it from being pruned while in use. RLock()
	// updates the last access time of the package as well.
	// RLock only succeeds if package is available.
	// The package should not be modified after build, so write locking is not
	// provided.
	RLock() error
	RUnlock() error
}
