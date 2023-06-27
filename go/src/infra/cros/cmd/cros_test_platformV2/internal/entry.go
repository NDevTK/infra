// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package internal

import (
	"context"

	parsers "infra/cros/cmd/cros_test_platformV2/tools"

	"fmt"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func Execute(inputPath string, cloud bool) (*api.TestSuite, error) {
	request, err := parsers.ReadInput(inputPath)
	if err != nil {
		fmt.Printf("Unable to parse: %s", err)
		// return nil, fmt.Errorf("unable to parse request: %s", err)
	}
	ctx := context.Background()

	// Build Exectors. Right now do CTRv2 + Both filters.
	// Note: the filter impls are currently nil.
	executors, _ := buildExecutors(ctx, request, cloud)

	InternalStruct := translateRequest(request)

	// Run the same commands for each
	for _, executor := range executors {
		// For CTR, init = start Server async. For services it will be pull/prep container/launch
		err := executor.Execute(ctx, "init", InternalStruct)
		fmt.Println("Executing Executor Init.")
		if err != nil {
			fmt.Println("Error")
		}
	}
	// Run the same commands for each
	for _, executor := range executors {
		// Gcloud auth for CTR (kinda odd...). For services, it will be `call the service.`
		err = executor.Execute(ctx, "run", InternalStruct)
		if err != nil {
			fmt.Println("Error")
		}
	}

	// After all execs are run, stop them all. TODO, this probably needs to be a bit smarter defered.
	for _, executor := range executors {
		err := executor.Execute(ctx, "stop", InternalStruct)
		if err != nil {
			fmt.Println("Error")
		}
	}
	return nil, nil

}
