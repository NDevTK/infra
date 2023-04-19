// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"

	"infra/builder_health_indicators/indicators_pb"

	//"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
)

func main() {
	request := indicators_pb.BBGenerateRequest{}

	build.Main(&request, nil, nil, func(ctx context.Context, userArgs []string, state *build.State) error {
		return Run(ctx)
	})
}

func Run(ctx context.Context) error {
	step, ctx := build.StartStep(ctx, "Hello world")
	var err error
	logging.Infof(ctx, "hello world of logging")
	defer func() { step.End(err) }()

	return nil
}
