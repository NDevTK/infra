// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package testsupport contains a test fixture that is used by most Karte
// tests.
package testsupport

import (
	"context"
	"time"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"

	"infra/cros/karte/internal/identifiers"
)

// Fixture is a test fixture for Karte.
type Fixture struct {
	Ctx context.Context
}

// NewFixture creates a new fixture for Karte tetsing.
func NewFixture(ctx context.Context) *Fixture {
	ctx = memory.Use(ctx)
	ctx = identifiers.Use(ctx, identifiers.NewNaive())
	testClock := testclock.New(time.Unix(10, 0).UTC())
	ctx = clock.Set(ctx, testClock)
	datastore.GetTestable(ctx).Consistent(true)
	return &Fixture{
		Ctx: ctx,
	}
}
