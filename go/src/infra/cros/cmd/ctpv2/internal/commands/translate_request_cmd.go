// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/ctpv2/internal/data"

	"github.com/gogo/protobuf/jsonpb"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"
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
	}

	internalStruct.SuiteInfo = &testapi.SuiteInfo{
		SuiteMetadata: suitemd,
		SuiteRequest:  cmd.CtpV2Req.GetSuiteRequest(),
	}

	cmd.InternalTestPlan = internalStruct

	return err
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
