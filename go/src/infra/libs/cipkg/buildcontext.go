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
	Platform Platform
	Storage  Storage
	context.Context
}

func (ctx *BuildContext) WithPlatform(plat Platform) *BuildContext {
	return &BuildContext{
		Platform: plat,
		Storage:  ctx.Storage,
		Context:  ctx,
	}
}

// Cross-compile platform tuple.
type Platform struct {
	Build  string
	Host   string
	Target string
}

// Storage represents the management interface for packages. Generator relies on
// storage to provide a place to store and retrieve packages.
type Storage interface {
	// Get(id) returns the handler for the package. It won't try to make package
	// available in the storage and not promising the returned package is valid.
	// Get a package added by Add(drv) with valid derivation will always return a
	// valid package.
	Get(id string) Package

	// Add(drv) returns a valid package if derivation is valid. Whether the
	// derivation will be persisted when it's added depends on storage's
	// implementation.
	Add(drv Derivation) Package

	// Prune(ctx, ttl, max) removes at most ${max} packages from storage which
	// haven't been used in the past ${ttl}.
	Prune(ctx context.Context, ttl time.Duration, max int)
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
	// return the last time it's used. There It only checks the completion stamp in
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
