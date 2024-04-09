// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/jsonpb"

	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/auth"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/ctpv2/data"
)

// FilterExecutionCmd represents test execution cmd.
type TranslateRequestCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	CtpReq *testapi.CTPRequest

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

	if sk.CtpReq == nil {
		return fmt.Errorf("Cmd %q missing dependency: CtpReq", cmd.GetCommandType())
	}

	cmd.CtpReq = sk.CtpReq
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
	if err = marsh.Marshal(req, cmd.CtpReq); err != nil {
		err = errors.Annotate(err, "failed to marshal proto").Err()
	}

	internalStruct := &testapi.InternalTestplan{}
	suitemd := &testapi.SuiteMetadata{
		Pool:              cmd.CtpReq.GetPool(),
		ExecutionMetadata: executionMetadata(cmd.CtpReq),
		DynamicUpdates:    []*api.UserDefinedDynamicUpdate{},
	}

	// Check if new field exists
	exists, err := common.CheckIfSchedulingUnitFieldExistsInSuiteMD()
	if err != nil {
		logging.Infof(ctx, "err while checking scheduling unit field check: %s", err)
	}

	// TODO (azrahman): remove this when all the default containers are up to date
	exists = false

	if exists {
		// new field exists that supports multi-dut
		suitemd.SchedulingUnits = getSchedulingUnits(cmd.CtpReq)
	} else {
		// non-multi-dut legacy flow to support backwards compatibility
		suitemd.TargetRequirements = targetRequirements(cmd.CtpReq)
	}

	suitemd.SchedulerInfo = generateSchedulerInfo(cmd.CtpReq)

	internalStruct.SuiteInfo = &testapi.SuiteInfo{
		SuiteMetadata: suitemd,
		SuiteRequest:  cmd.CtpReq.GetSuiteRequest(),
	}

	translated_req := step.Log("translated request")
	if err = marsh.Marshal(translated_req, internalStruct); err != nil {
		err = errors.Annotate(err, "failed to marshal proto").Err()
	}

	cmd.InternalTestPlan = internalStruct

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

func targetRequirements(req *testapi.CTPRequest) []*testapi.TargetRequirements {
	targs := []*testapi.TargetRequirements{}
	for _, scheduleTarget := range req.GetScheduleTargets() {
		// TODO (azrahman): 0 indexing now for single dut. Add multi-dut support.
		targ := scheduleTarget.GetTargets()[0]
		switch hw := targ.HwTarget.Target.(type) {
		case *testapi.HWTarget_LegacyHw:

			// There will only be one set by the translation; but other filters might
			// expand this as they see fit.
			var hwDefs []*testapi.SwarmingDefinition
			hwDefs = append(hwDefs, buildHwDef(hw.LegacyHw))

			legacysw := legacyswpoper(targ.SwTarget)

			builtTarget := &testapi.TargetRequirements{
				HwRequirements: &testapi.HWRequirements{
					HwDefinition: hwDefs,
				},

				SwRequirement: legacysw,
			}
			targs = append(targs, builtTarget)
		}
	}
	return targs
}

func getSchedulingUnits(req *testapi.CTPRequest) []*testapi.SchedulingUnit {
	schedUnits := []*testapi.SchedulingUnit{}
	for _, scheduleTarget := range req.GetScheduleTargets() {
		newSchedUnit := &api.SchedulingUnit{CompanionTargets: []*api.Target{}}
		for i, targ := range scheduleTarget.GetTargets() {
			newTarget := TargetsToNewTarget(targ)
			if i == 0 {
				// primary target
				newSchedUnit.PrimaryTarget = newTarget
			} else {
				// secondary target
				newSchedUnit.CompanionTargets = append(newSchedUnit.CompanionTargets, newTarget)
			}
		}
		schedUnits = append(schedUnits, newSchedUnit)
	}
	return schedUnits
}

func TargetsToNewTarget(targ *testapi.Targets) *api.Target {
	switch hw := targ.HwTarget.Target.(type) {
	case *testapi.HWTarget_LegacyHw:
		// There will only be one set by the translation; but other filters might
		// expand this as they see fit.
		swDef := buildHwDef(hw.LegacyHw)
		legacysw := legacyswpoper(targ.SwTarget)

		return &api.Target{SwarmingDef: swDef, SwReq: legacysw}
	}
	return nil
}

func legacyswpoper(sws *testapi.SWTarget) *testapi.LegacySW {
	switch sw := sws.SwTarget.(type) {
	case *testapi.SWTarget_LegacySw:
		return sw.LegacySw
	}
	return nil
}

func buildHwDef(hw *testapi.LegacyHW) *testapi.SwarmingDefinition {
	dut := &labapi.Dut{}

	Cros := &labapi.Dut_ChromeOS{DutModel: &labapi.DutModel{
		BuildTarget: hw.Board,
		ModelName:   hw.Model,
	}}
	dut.DutType = &labapi.Dut_Chromeos{Chromeos: Cros}

	return &testapi.SwarmingDefinition{DutInfo: dut, Variant: hw.GetVariant(),
		SwarmingLabels: hw.GetSwarmingDimensions()}
}

func NewTranslateRequestCmd() *TranslateRequestCmd {
	abstractCmd := interfaces.NewAbstractCmd(TranslateRequestType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &TranslateRequestCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}

func generateSchedulerInfo(req *api.CTPRequest) *api.SchedulerInfo {
	return req.GetSchedulerInfo()
}

func executionMetadata(req *api.CTPRequest) *api.ExecutionMetadata {
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
