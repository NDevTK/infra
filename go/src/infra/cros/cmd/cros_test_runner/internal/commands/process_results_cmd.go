// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	"go.chromium.org/chromiumos/config/go/test/api"
	commonpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/common"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
)

// ProcessResultsCmd represents process results command.
type ProcessResultsCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps (all are optional)
	CftTestRequest *skylab_test_runner.CFTTestRequest
	GcsUrl         string
	StainlessUrl   string
	TesthausUrl    string
	ProvisionResp  *api.InstallResponse
	TestResponses  *api.CrosTestResponse

	// Updates
	SkylabResult *skylab_test_runner.Result
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *ProcessResultsCmd) ExtractDependencies(ctx context.Context, ski interfaces.StateKeeperInterface) error {
	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, sk)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *ProcessResultsCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.updateHwTestStateKeeper(ctx, sk)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// Execute executes the command.
func (cmd *ProcessResultsCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Results")
	defer func() { step.End(err) }()

	common.AddLinksToStepSummaryMarkdown(step, cmd.TesthausUrl, cmd.StainlessUrl, common.GetGcsClickableLink(cmd.GcsUrl))

	// Default values
	prejobVerdict := skylab_test_runner.Result_Prejob_Step_VERDICT_UNDEFINED
	prejobReason := ""
	isIncomplete := true
	logData := getLogData(cmd.TesthausUrl, cmd.StainlessUrl, cmd.GcsUrl)

	// Parse provision info
	if cmd.ProvisionResp != nil {
		if cmd.ProvisionResp.GetStatus() == api.InstallResponse_STATUS_OK {
			prejobVerdict = skylab_test_runner.Result_Prejob_Step_VERDICT_PASS
		} else {
			prejobVerdict = skylab_test_runner.Result_Prejob_Step_VERDICT_FAIL
		}
		prejobReason = cmd.ProvisionResp.GetStatus().String()
		_ = common.CreateStepWithStatus(ctx, "Provision", cmd.ProvisionResp.GetStatus().String(), cmd.ProvisionResp.GetStatus() != api.InstallResponse_STATUS_SUCCESS, false)
	}

	// Parse test results
	autotestTestCases := []*skylab_test_runner.Result_Autotest_TestCase{}
	if cmd.TestResponses != nil && len(cmd.TestResponses.GetTestCaseResults()) > 0 {
		isIncomplete = false
		for _, testResult := range cmd.TestResponses.GetTestCaseResults() {
			testVerdict, isTestFailure := getTestVerdict(ctx, testResult)
			testResultReason := testResult.GetReason()
			autotestTestCase := &skylab_test_runner.Result_Autotest_TestCase{
				Name:                 testResult.GetTestCaseId().GetValue(),
				Verdict:              testVerdict,
				HumanReadableSummary: testResultReason,
			}
			autotestTestCases = append(autotestTestCases, autotestTestCase)

			// Set test steps
			_ = common.CreateStepWithStatus(ctx, testResult.GetTestCaseId().GetValue(), testResultReason, isTestFailure, false)
		}
	}

	// If no test results, add default results from input.
	if len(autotestTestCases) == 0 {
		autotestTestCases = getDefaultAutotestTestCasesResult(ctx, cmd.CftTestRequest)
	}

	autotestResult := &skylab_test_runner.Result_Autotest{
		TestCases:  autotestTestCases,
		Incomplete: isIncomplete,
	}
	skylabResult := &skylab_test_runner.Result{
		Prejob: &skylab_test_runner.Result_Prejob{
			Step: []*skylab_test_runner.Result_Prejob_Step{
				{
					Name:                 "provision",
					Verdict:              prejobVerdict,
					HumanReadableSummary: prejobReason,
				},
			},
		},
		AutotestResults: map[string]*skylab_test_runner.Result_Autotest{
			"original_test": autotestResult,
		},
		StateUpdate: &skylab_test_runner.Result_StateUpdate{
			DutState: "ready",
		},
		LogData: logData,
	}

	cmd.SkylabResult = skylabResult
	common.WriteProtoToStepLog(ctx, step, skylabResult, "skylab_result")

	return nil
}

// extractDepsFromHwTestStateKeeper extracts cmd deps from hw test state keeper.
func (cmd *ProcessResultsCmd) extractDepsFromHwTestStateKeeper(ctx context.Context, sk *data.HwTestStateKeeper) error {
	if sk.CftTestRequest == nil {
		logging.Infof(ctx, "Warning: cmd %q missing non-critical dependency: CftTestRequest", cmd.GetCommandType())
	}
	if sk.ProvisionResp == nil {
		logging.Infof(ctx, "Warning: cmd %q missing non-critical dependency: ProvisionResp", cmd.GetCommandType())
	}
	if sk.TestResponses == nil {
		logging.Infof(ctx, "Warning: cmd %q missing non-critical dependency: TestResponses", cmd.GetCommandType())
	}
	if sk.GcsUrl == "" {
		logging.Infof(ctx, "Warning: cmd %q missing non-critical dependency: GcsUrl", cmd.GetCommandType())
	}
	if sk.StainlessUrl == "" {
		logging.Infof(ctx, "Warning: cmd %q missing non-critical dependency: StainlessUrl", cmd.GetCommandType())
	}
	if sk.TesthausUrl == "" {
		logging.Infof(ctx, "Warning: cmd %q missing non-critical dependency: TesthausUrl", cmd.GetCommandType())
	}

	cmd.CftTestRequest = sk.CftTestRequest
	cmd.ProvisionResp = sk.ProvisionResp
	cmd.TestResponses = sk.TestResponses
	cmd.GcsUrl = sk.GcsUrl
	cmd.StainlessUrl = sk.StainlessUrl
	cmd.TesthausUrl = sk.TesthausUrl

	return nil
}

func (cmd *ProcessResultsCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.SkylabResult != nil {
		sk.SkylabResult = cmd.SkylabResult
	}

	return nil
}

// getTestVerdict converts testcase result to testcase verdict.
func getTestVerdict(ctx context.Context, testResult *api.TestCaseResult) (skylab_test_runner.Result_Autotest_TestCase_Verdict, bool) {
	// Default values
	isTestFailure := true
	var testVerdict skylab_test_runner.Result_Autotest_TestCase_Verdict

	// Convert testcase result to testcase verdict
	switch testResult.Verdict.(type) {
	case *api.TestCaseResult_Pass_:
		isTestFailure = false
		testVerdict = skylab_test_runner.Result_Autotest_TestCase_VERDICT_PASS
	case *api.TestCaseResult_Fail_:
		testVerdict = skylab_test_runner.Result_Autotest_TestCase_VERDICT_FAIL
	case *api.TestCaseResult_Abort_:
		testVerdict = skylab_test_runner.Result_Autotest_TestCase_VERDICT_ABORT
	case *api.TestCaseResult_Crash_:
		testVerdict = skylab_test_runner.Result_Autotest_TestCase_VERDICT_ERROR
	case *api.TestCaseResult_Skip_:
		testVerdict = skylab_test_runner.Result_Autotest_TestCase_VERDICT_NO_VERDICT
	case *api.TestCaseResult_NotRun_:
		testVerdict = skylab_test_runner.Result_Autotest_TestCase_VERDICT_NO_VERDICT
	default:
		logging.Infof(ctx, "No valid test case result status found for %s.", testResult.GetTestCaseId().GetValue())
		testVerdict = skylab_test_runner.Result_Autotest_TestCase_VERDICT_NO_VERDICT
	}

	return testVerdict, isTestFailure
}

// getDefaultAutotestTestCasesResult constructs default result from input.
func getDefaultAutotestTestCasesResult(ctx context.Context, req *skylab_test_runner.CFTTestRequest) []*skylab_test_runner.Result_Autotest_TestCase {
	autotestTestCases := []*skylab_test_runner.Result_Autotest_TestCase{}
	for _, testSuite := range req.GetTestSuites() {
		for _, testCaseId := range testSuite.GetTestCaseIds().GetTestCaseIds() {
			autotestTestCase := &skylab_test_runner.Result_Autotest_TestCase{
				Name:    testCaseId.GetValue(),
				Verdict: skylab_test_runner.Result_Autotest_TestCase_VERDICT_NO_VERDICT,
			}
			autotestTestCases = append(autotestTestCases, autotestTestCase)

			_ = common.CreateStepWithStatus(ctx, testCaseId.GetValue(), common.TestDidNotRunErr, true, false)
		}
	}

	return autotestTestCases
}

// getLogData constructs tasklogdata from provided links.
func getLogData(testhausUrl string, stainlessUrl string, gcsUrl string) *commonpb.TaskLogData {
	logData := &commonpb.TaskLogData{}
	if testhausUrl != "" {
		logData.TesthausUrl = testhausUrl
	}
	if stainlessUrl != "" {
		logData.StainlessUrl = testhausUrl
	}
	if gcsUrl != "" {
		logData.GsUrl = gcsUrl
	}

	return logData
}

func NewProcessResultsCmd() *ProcessResultsCmd {
	abstractCmd := interfaces.NewAbstractCmd(ProcessResultsCmdType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &ProcessResultsCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}
