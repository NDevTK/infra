// Copyright (c) 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/logging"
)

var Logger logging.Logger

var (
	logFailFast bool
	todoCount   int32

	leakyUnique   = stringset.New(0)
	leakyUniqueMu sync.Mutex
)

// PIN indicates a pinning violation where this resolution tool is using potentially
// different logic from the pinned tool/script to make a resolution procedure.
//
// This means that the pin for this resolution logic is not recorded in the
// derived Manifest.
func PIN(format string, args ...any) {
	Logger.LogCall(logging.Warning, 1, "PIN - "+format, args)
}

// LEAKY indicates a leaky abstraction or an assumption that crderiveinputs is
// making.
func LEAKY(format string, args ...any) {
	msg := fmt.Sprintf("LEAKY - "+format, args...)
	leakyUniqueMu.Lock()
	defer leakyUniqueMu.Unlock()
	if leakyUnique.Has(msg) {
		return
	}
	leakyUnique.Add(msg)
	Logger.LogCall(logging.Warning, 1, "%s", []any{msg})
}

// SBOM indicates an area where we may be pulling in more sources than are
// necessary for the actual build.
func SBOM(format string, args ...any) {
	Logger.LogCall(logging.Warning, 1, "SBOM - "+format, args)
}

// IMPROVE indicates an area where the underlying program needs improvement, but
// isn't something which (currently) affects correctness (though it may be
// a source of fragility).
func IMPROVE(format string, args ...any) {
	Logger.LogCall(logging.Warning, 1, "IMPROVE - "+format, args)
}

// TODO indicates that there is some critical dependecy which is being skipped
// because it is not yet implemented.
func TODO(format string, args ...any) {
	Logger.LogCall(logging.Error, 1, "TODO - "+format, args)
	atomic.AddInt32(&todoCount, 1)
	if logFailFast {
		Logger.Infof("-fail-fast specified, exiting.")
		os.Exit(1)
	}
}
