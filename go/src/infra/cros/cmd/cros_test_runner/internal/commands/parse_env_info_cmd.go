// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/libs/skylab/common/heuristics"
	"os"
	"strings"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe/build"
)

// ParseEnvInfoCmd represents parse env info cmd.
type ParseEnvInfoCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Updates
	HostName            string
	CurrentInvocationId string
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *ParseEnvInfoCmd) UpdateStateKeeper(
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

func (cmd *ParseEnvInfoCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Parse environment information")
	defer func() { step.End(err) }()

	summary := []string{}

	botID := os.Getenv("SWARMING_BOT_ID")
	hostName := heuristics.NormalizeBotNameToDeviceName(botID)
	summary = append(summary, fmt.Sprintf("hostname: %s", hostName))
	cmd.HostName = hostName

	// Write all env vars to log for debugging
	envVarsLog := step.Log("environment vars")
	for _, e := range os.Environ() {
		envVarsLog.Write([]byte(fmt.Sprintf("%s\n", e)))
	}

	// Get current invocation name
	rdb := lucictx.GetResultDB(ctx)
	cmd.CurrentInvocationId = rdb.GetCurrentInvocation().GetName()
	summary = append(summary, fmt.Sprintf("currentInvocationId: %s", cmd.CurrentInvocationId))

	step.SetSummaryMarkdown(strings.Join(summary, "\n"))

	return nil
}

func (cmd *ParseEnvInfoCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.HostName != "" {
		sk.HostName = cmd.HostName
	}

	if cmd.CurrentInvocationId != "" {
		sk.CurrentInvocationId = cmd.CurrentInvocationId
	}

	return nil
}

func NewParseEnvInfoCmd() *ParseEnvInfoCmd {
	abstractCmd := interfaces.NewAbstractCmd(ParseEnvInfoCmdType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &ParseEnvInfoCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}
