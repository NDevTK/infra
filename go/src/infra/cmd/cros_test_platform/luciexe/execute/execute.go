// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package execute houses the top-level logic for the execute step.
package execute

import (
	"context"
	"fmt"

	"go.chromium.org/luci/luciexe/exe"

	"github.com/golang/protobuf/proto"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"infra/cmd/cros_test_platform/internal/execution"
	trservice "infra/cmd/cros_test_platform/internal/execution/testrunner/service"
	"infra/cmd/cros_test_platform/luciexe/common"
)

// Args contains all the arguments necessary to Run() an execute step.
type Args struct {
	InputPath      string
	OutputPath     string
	SwarmingTaskID string

	Build *bbpb.Build
	Send  exe.BuildSender
}

// Run is the entry point for an execute step.
func Run(ctx context.Context, args Args) error {
	request := &steps.ExecuteRequests{}
	if err := common.ReadRequest(args.InputPath, request); err != nil {
		return err
	}

	if err := validateRequests(request.TaggedRequests); err != nil {
		return err
	}

	cfg := extractOneConfig(request.TaggedRequests)

	ea := execution.Args{
		Request:      request,
		WorkerConfig: cfg.SkylabWorker,
		ParentTaskID: args.SwarmingTaskID,
	}
	// crbug.com/1112514 These arguments optional during the transition to
	// luciexe.
	if args.Build != nil {
		ea.Build = args.Build
		ea.Send = args.Send
	} else {
		ea.Build = &bbpb.Build{}
		ea.Send = func() {}
	}

	skylab, err := trservice.NewClient(ctx, cfg, ea.Build.GetInfra().GetResultdb().GetHostname())
	if err != nil {
		return err
	}

	var resps map[string]*steps.ExecuteResponse
	resps, err = execution.Run(ctx, skylab, ea)

	// Timeout while waiting for tasks is not considered an Test Platform
	// infrastructure error because root cause is mostly related to fleet
	// capacity or long test runtimes.
	if execution.IsGlobalTimeoutError(ctx, err) {
		logging.Warningf(ctx, "Exited wait dut to timeout: %s", err)
		logging.Warningf(ctx, "Execution responses will contain test failures as a consequence of the timeout.")
	}
	if err != nil {
		return err
	}

	updateWithEnumerationErrors(ctx, resps, request.TaggedRequests)
	return common.WriteResponse(
		args.OutputPath,
		&steps.ExecuteResponses{
			TaggedResponses: resps,
		},
	)
}

func extractOneConfig(trs map[string]*steps.ExecuteRequest) *config.Config {
	for _, r := range trs {
		return r.Config
	}
	return nil
}

func validateRequests(trs map[string]*steps.ExecuteRequest) error {
	if len(trs) == 0 {
		return errors.Reason("zero requests").Err()
	}

	for t, r := range trs {
		if err := validateRequest(r); err != nil {
			return errors.Annotate(err, "validate request %s", t).Err()
		}
	}

	var sTag string
	var sReq *steps.ExecuteRequest
	for t, r := range trs {
		sTag = t
		sReq = r
		break
	}

	sCfg := sReq.Config
	if err := validateRequestConfig(sCfg); err != nil {
		return errors.Annotate(err, "validate request %s", sTag).Err()
	}
	for t, r := range trs {
		o := r.Config
		if !proto.Equal(sCfg, o) {
			return errors.Reason("validate request: mistmatched config: %s[%#v] vs %s[%#v]", sTag, sCfg, t, o).Err()
		}
	}

	sTimeout := sReq.RequestParams.Time.MaximumDuration
	for t, r := range trs {
		o := r.RequestParams.Time.MaximumDuration
		if !proto.Equal(sTimeout, o) {
			return errors.Reason("validate request: per-request timeout support unimplemented: %s[%s] vs %s[%s]", sTag, sTimeout, t, o).Err()
		}
	}
	return nil
}

func validateRequest(request *steps.ExecuteRequest) error {
	if request == nil {
		return fmt.Errorf("nil request")
	}
	if request.Config == nil {
		return fmt.Errorf("nil request.config")
	}
	return nil
}

func validateRequestConfig(cfg *config.Config) error {
	if cfg.SkylabSwarming == nil {
		return fmt.Errorf("nil request.config.skylab_swarming")
	}
	if cfg.SkylabWorker == nil {
		return fmt.Errorf("nil request.config.skylab_worker")
	}
	return nil
}

func updateWithEnumerationErrors(ctx context.Context, resps map[string]*steps.ExecuteResponse, reqs map[string]*steps.ExecuteRequest) {
	for t, resp := range resps {
		req, ok := reqs[t]
		if !ok {
			panic(fmt.Sprintf("request for non-existent request for %s", t))
		}
		if es := req.GetEnumeration().GetErrorSummary(); es != "" {
			if resp.State == nil {
				resp.State = &test_platform.TaskState{}
			}
			resp.State.Verdict = test_platform.TaskState_VERDICT_FAILED
			logging.Infof(ctx, "Set request %s to VERDICT_FAILED because of enumeration error: %s", t, es)
		}
	}
}
