// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	"github.com/gogo/protobuf/jsonpb"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"
)

// BuildInputValidationCmd represents build input validation command.
type BuildInputValidationCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	CftTestRequest *skylab_test_runner.CFTTestRequest
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *BuildInputValidationCmd) ExtractDependencies(ctx context.Context, ski interfaces.StateKeeperInterface) error {
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

// Execute executes the command.
func (cmd *BuildInputValidationCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Inputs validation")
	defer func() { step.End(err) }()

	if cmd.CftTestRequest.GetParentBuildId() != 0 {
		step.SetSummaryMarkdown(fmt.Sprintf("* [parent CTP](https://cr-buildbucket.appspot.com/build/%d)", cmd.CftTestRequest.GetParentBuildId()))
	}

	req := step.Log("request")
	marsh := jsonpb.Marshaler{Indent: "  "}
	if err = marsh.Marshal(req, cmd.CftTestRequest); err != nil {
		err = errors.Annotate(err, "failed to marshal proto").Err()
	}

	// TODO (azrahman): add inputs validations steps here.

	return err
}

func (cmd *BuildInputValidationCmd) extractDepsFromHwTestStateKeeper(ctx context.Context, sk *data.HwTestStateKeeper) error {
	if sk.CftTestRequest == nil {
		return fmt.Errorf("Cmd %q missing dependency: CftTestRequest", cmd.GetCommandType())
	}
	cmd.CftTestRequest = sk.CftTestRequest

	return nil
}

func NewBuildInputValidationCmd() *BuildInputValidationCmd {
	abstractCmd := interfaces.NewAbstractCmd(BuildInputValidationCmdType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &BuildInputValidationCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}
