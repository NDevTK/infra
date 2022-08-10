// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Binary fakebuild is a luciexe binary that pretends to do some work.
//
// To be used for Swarming and Buildbucket load testing.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/data/rand/mathrand"
	"go.chromium.org/luci/luciexe/build"

	"infra/experimental/swarming/fakebuild/fakebuildpb"
)

func main() {
	mathrand.SeedRandomly()

	inputs := &fakebuildpb.Inputs{}

	build.Main(inputs, nil, nil, func(ctx context.Context, args []string, st *build.State) error {
		for i := 0; i < int(inputs.Steps); i++ {
			sleepStep(ctx, inputs, i)
		}
		return nil
	})
}

func sleepStep(ctx context.Context, inputs *fakebuildpb.Inputs, idx int) {
	var secs int64
	if dt := inputs.SleepMaxSec - inputs.SleepMinSec; dt > 0 {
		secs = inputs.SleepMinSec + rand.Int63n(dt)
	} else {
		secs = inputs.SleepMinSec
	}

	step, ctx := build.StartStep(ctx, fmt.Sprintf("Step %d: sleep %d", idx+1, secs))
	defer func() { step.End(nil) }()

	clock.Sleep(ctx, time.Duration(secs)*time.Second)
}
