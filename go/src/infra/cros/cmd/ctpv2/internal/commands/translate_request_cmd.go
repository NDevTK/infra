// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"strings"

	// "infra/libs/skylab/inventory"
	// "infra/libs/skylab/request"
	// "infra/libs/skylab/worker"
	"net/http"

	"github.com/gogo/protobuf/jsonpb"

	// "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"

	// "go.chromium.org/chromiumos/infra/proto/go/test_platform"
	// "go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/auth"
	// "go.chromium.org/luci/buildbucket"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"

	// "go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/prpc"
	// "go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/ctpv2/data"
)

// FilterExecutionCmd represents test execution cmd.
type TranslateRequestCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	CtpV2Req *testapi.CTPv2Request

	// Updates
	InternalTestPlan *testapi.InternalTestplan
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *TranslateRequestCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.FilterStateKeeper:
		err = cmd.extractDepsFromFilterStateKeepr(ctx, sk)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *TranslateRequestCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.FilterStateKeeper:
		err = cmd.updateLocalTestStateKeeper(ctx, sk)
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

func (cmd *TranslateRequestCmd) extractDepsFromFilterStateKeepr(
	ctx context.Context,
	sk *data.FilterStateKeeper) error {

	if sk.CtpV2Req == nil {
		return fmt.Errorf("Cmd %q missing dependency: CtpV2Req", cmd.GetCommandType())
	}

	cmd.CtpV2Req = sk.CtpV2Req
	return nil
}

func (cmd *TranslateRequestCmd) updateLocalTestStateKeeper(
	ctx context.Context,
	sk *data.FilterStateKeeper) error {

	if cmd.InternalTestPlan != nil {
		sk.InitialInternalTestPlan = cmd.InternalTestPlan
	}

	return nil
}

// Execute executes the command.
func (cmd *TranslateRequestCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Translate request")
	defer func() { step.End(err) }()

	req := step.Log("request received")
	marsh := jsonpb.Marshaler{Indent: "  "}
	if err = marsh.Marshal(req, cmd.CtpV2Req); err != nil {
		err = errors.Annotate(err, "failed to marshal proto").Err()
	}

	internalStruct := &testapi.InternalTestplan{}
	targs := targetRequirements(cmd.CtpV2Req)
	suitemd := &testapi.SuiteMetadata{
		TargetRequirements: targs,
		Pool:               cmd.CtpV2Req.GetPool(),
		ExecutionMetadata:  executionMetadata(cmd.CtpV2Req),
	}

	internalStruct.SuiteInfo = &testapi.SuiteInfo{
		SuiteMetadata: suitemd,
		SuiteRequest:  cmd.CtpV2Req.GetSuiteRequest(),
	}

	translated_req := step.Log("translated request")
	if err = marsh.Marshal(translated_req, internalStruct); err != nil {
		err = errors.Annotate(err, "failed to marshal proto").Err()
	}

	cmd.InternalTestPlan = internalStruct

	// // -- Test out TRv2 Scheduling --
	// deadline := timestamppb.New(time.Now().Add(2 * time.Hour))
	// parentRequestUID := "TestPlanRuns/8771381664436235537/drallion-cq.hw.bvt-tast-cq"
	// var parentBuildId int64 = 8771381664436235537

	// buildTarget := "zork"
	// modelName := "morphius"
	// dutModel := &labapi.DutModel{
	// 	BuildTarget: buildTarget,
	// 	ModelName:   modelName,
	// }
	// companionDuts := []*skylab_test_runner.CFTTestRequest_Device{}
	// containerMetadata := &api.ContainerMetadata{}
	// testSuites := []*testapi.TestSuite{}
	// kv := map[string]string{}

	// cftTestRequest := &skylab_test_runner.CFTTestRequest{
	// 	Deadline:         deadline,
	// 	ParentRequestUid: parentRequestUID,
	// 	ParentBuildId:    parentBuildId,
	// 	PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
	// 		DutModel:             dutModel,
	// 		ProvisionState:       nil,
	// 		ContainerMetadataKey: buildTarget,
	// 	},
	// 	CompanionDuts:                companionDuts,
	// 	ContainerMetadata:            containerMetadata,
	// 	TestSuites:                   testSuites,
	// 	DefaultTestExecutionBehavior: test_platform.Request_Params_NON_CRITICAL,
	// 	AutotestKeyvals:              kv,
	// 	RunViaTrv2:                   true,
	// 	StepsConfig:                  nil,
	// }

	// // -- Create request
	// cmd1 := &worker.Command{
	// 	ClientTest:      false,
	// 	Deadline:        time.Now().Add(2 * time.Hour),
	// 	Keyvals:         kv,
	// 	OutputToIsolate: true,
	// 	TaskName:        "foo",
	// 	TestArgs:        "bar",
	// }
	// labels := &inventory.SchedulableLabels{}
	// dims := []string{"label-board:zork", "label-model:morphius", "dut_state:ready"}
	// args := request.Args{
	// 	Cmd:                              *cmd1,
	// 	SchedulableLabels:                labels,
	// 	SecondaryDevicesLabels:           []*inventory.SchedulableLabels{},
	// 	Dimensions:                       dims,
	// 	ParentTaskID:                     "645cad975ff6cf11",
	// 	ParentRequestUID:                 parentRequestUID,
	// 	Priority:                         10,
	// 	ProvisionableDimensions:          []string{},
	// 	ProvisionableDimensionExpiration: 2 * time.Hour,
	// 	StatusTopic:                      "",
	// 	SwarmingTags:                     dims,
	// 	SwarmingPool:                     "DUT_POOL_QUOTA",
	// 	TestRunnerRequest:                nil,
	// 	CFTTestRunnerRequest:             cftTestRequest,
	// 	CFTIsEnabled:                     true,
	// 	Timeout:                          2 * time.Hour,
	// 	Experiments:                      nil,
	// 	GerritChanges:                    nil,
	// }

	// builderId := &buildbucketpb.BuilderID{
	// 	Project: "chromeos",
	// 	Bucket:  "test_runner",
	// 	Builder: "test_runner",
	// }

	// req1, err := args.NewBBRequest(builderId)
	// if err != nil {
	// 	return err
	// }

	// // Check if there's a parent build for the task to be launched.
	// // If a ScheduleBuildToken can be found in the Buildbucket section of LUCI_CONTEXT,
	// // it will be the token for the parent build.
	// // Attaching the token to the ScheduleBuild request will enable Buildbucket to
	// // track the parent/child build relationship between the build with the token
	// // and this new build.
	// bbCtx := lucictx.GetBuildbucket(ctx)
	// // Do not attach the buildbucket token if it's empty or the build is a led build.
	// // Because led builds are not real Buildbucket builds and they don't have
	// // real buildbucket tokens, so we cannot make them  any builds's parent,
	// // even for the builds they scheduled.
	// if bbCtx != nil && bbCtx.GetScheduleBuildToken() != "" && bbCtx.GetScheduleBuildToken() != buildbucket.DummyBuildbucketToken {
	// 	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(buildbucket.BuildbucketTokenHeader, bbCtx.ScheduleBuildToken))

	// 	// Decide if the child can outlive its parent or not.
	// 	// For details see https://source.chromium.org/chromium/infra/infra/+/main:go/src/go.chromium.org/luci/buildbucket/proto/builds_service.proto;l=458;drc=79232ce0a0af1f7ab9ae78efa9ab3931a520d2bc.
	// 	if req1.GetCanOutliveParent() == buildbucketpb.Trinary_UNSET {
	// 		// We do not want test_runner runs to outrun parent CTP.
	// 		req1.CanOutliveParent = buildbucketpb.Trinary_YES
	// 		if req1.GetSwarming().GetParentRunId() != "" {
	// 			req1.CanOutliveParent = buildbucketpb.Trinary_YES
	// 		}
	// 	}
	// }

	// bbClient, err := newBBClient(ctx)
	// if err != nil {
	// 	return err
	// }

	// resp, err := bbClient.ScheduleBuild(ctx, req1)
	// if err != nil {
	// 	return err
	// }

	// logging.Infof(ctx, "buildbucket id: %d", resp.Id)

	// time.Sleep(5 * time.Minute)

	// // --------  End testing -------

	return err
}

func newBBClient(ctx context.Context) (buildbucketpb.BuildsClient, error) {
	hClient, err := httpClient(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "create buildbucket client").Err()
	}
	pClient := &prpc.Client{
		C:    hClient,
		Host: "cr-buildbucket.appspot.com",
	}
	return buildbucketpb.NewBuildsPRPCClient(pClient), nil
}

func httpClient(ctx context.Context) (*http.Client, error) {
	a := auth.NewAuthenticator(ctx, auth.SilentLogin, auth.Options{
		Scopes: []string{auth.OAuthScopeEmail},
	})
	h, err := a.Client()
	if err != nil {
		return nil, errors.Annotate(err, "create http client").Err()
	}
	return h, nil
}

func targetRequirements(req *testapi.CTPv2Request) (targs []*testapi.TargetRequirements) {
	for _, targ := range req.Targets {
		switch hw := targ.HwTarget.Target.(type) {
		case *testapi.HWTarget_LegacyHw:

			// There will only be one set by the translation; but other filters might
			// expand this as they see fit.
			var hwDefs []*testapi.SwarmingDefinition
			hwDefs = append(hwDefs, buildcrosDut(hw.LegacyHw))

			legacysw := legacyswpoper(targ.SwTargets)

			builtTarget := &testapi.TargetRequirements{
				HwRequirements: &testapi.HWRequirements{
					HwDefinition: hwDefs,
				},

				SwRequirements: legacysw,
			}
			targs = append(targs, builtTarget)
		}
	}
	return targs
}

func legacyswpoper(sws []*testapi.SWTarget) []*testapi.LegacySW {
	var legsws []*testapi.LegacySW
	for _, swTarg := range sws {
		switch sw := swTarg.SwTarget.(type) {
		case *testapi.SWTarget_LegacySw:
			legsws = append(legsws, sw.LegacySw)
		}
	}
	return legsws
}

func buildcrosDut(hw *testapi.LegacyHW) *testapi.SwarmingDefinition {
	dut := &labapi.Dut{}

	Cros := &labapi.Dut_ChromeOS{DutModel: &labapi.DutModel{
		BuildTarget: hw.Board,
		ModelName:   hw.Model,
	}}
	dut.DutType = &labapi.Dut_Chromeos{Chromeos: Cros}

	return &testapi.SwarmingDefinition{DutInfo: dut}
}

func NewTranslateRequestCmd() *TranslateRequestCmd {
	abstractCmd := interfaces.NewAbstractCmd(TranslateRequestType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &TranslateRequestCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}

func executionMetadata(req *api.CTPv2Request) *api.ExecutionMetadata {
	ta := req.GetSuiteRequest().GetTestArgs()
	args := &testapi.ExecutionMetadata{}
	things := []*testapi.Arg{}

	// ta will often be a comma deliminated string such as:
	// "foo=bar,zoo=mar"
	// Split on the comma, then split again on the =
	// Lazy parsing the `=`; as KV support is weak at best.
	// Users are responsible for clean args.
	for _, kv := range strings.Split(ta, " ") {
		k := ""
		v := ""
		for _, innerkv := range strings.Split(kv, "=") {
			if k == "resultdb_settings" {
				continue
			} else if k == "" {
				k = innerkv
			} else if v == "" {
				v = innerkv
			} else {
				fmt.Println("too many values to unpack, skipping ", innerkv)
				k = ""
				v = ""
			}
		}
		kvproto := &testapi.Arg{
			Flag:  k,
			Value: v,
		}
		things = append(things, kvproto)
	}

	args.Args = things
	return args
}
