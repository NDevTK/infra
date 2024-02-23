// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/common_builders"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/ctpv2/data"
)

// TranslateV1ToV2Cmd represents v1 to v2 translation cmd.
type TranslateV1ToV2Cmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	CtpV1Requests map[string]*test_platform.Request
	CtpV2Request  *api.CTPv2Request // This will be updated if isn't set by deps
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *TranslateV1ToV2Cmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.PrePostFilterStateKeeper:
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
func (cmd *TranslateV1ToV2Cmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.PrePostFilterStateKeeper:
		err = cmd.updateLocalTestStateKeeper(ctx, sk)
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

func (cmd *TranslateV1ToV2Cmd) extractDepsFromFilterStateKeepr(
	ctx context.Context,
	sk *data.PrePostFilterStateKeeper) error {

	if sk.CtpV2Request == nil || len(sk.CtpV2Request.GetRequests()) == 0 {
		if sk.CtpV1Requests == nil || len(sk.CtpV1Requests) == 0 {
			return fmt.Errorf("Cmd %q missing dependency: Either v1 or v2 request is required!", cmd.GetCommandType())
		} else {
			cmd.CtpV1Requests = sk.CtpV1Requests
		}
	} else {
		cmd.CtpV2Request = sk.CtpV2Request
	}

	return nil
}

func (cmd *TranslateV1ToV2Cmd) updateLocalTestStateKeeper(
	ctx context.Context,
	sk *data.PrePostFilterStateKeeper) error {

	if cmd.CtpV2Request != nil {
		sk.CtpV2Request = cmd.CtpV2Request
	}

	return nil
}

// Execute executes the command.
func (cmd *TranslateV1ToV2Cmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Translate v1 to v2 request (if required)")
	defer func() { step.End(err) }()

	// exit if v2 is already set
	if cmd.CtpV2Request != nil && len(cmd.CtpV2Request.GetRequests()) > 0 {
		step.SetSummaryMarkdown("V2 request already set so skipping translation...")
		return nil
	}
	common.WriteAnyObjectToStepLog(ctx, step, cmd.CtpV1Requests, "Received CtpV1 Request")
	cmd.CtpV2Request = common_builders.NewCTPV2FromV1(ctx, cmd.CtpV1Requests).BuildRequest()
	step.SetSummaryMarkdown("Translation succeeded")
	common.WriteProtoToStepLog(ctx, step, cmd.CtpV2Request, "Translated CtpV2 Request")

	return err
}

// NewTranslateV1toV2Cmd returns a new TranslateV1ToV2Cmd
func NewTranslateV1toV2Cmd() *TranslateV1ToV2Cmd {
	abstractCmd := interfaces.NewAbstractCmd(TranslateV1toV2RequestType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &TranslateV1ToV2Cmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}
