// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/subcommands"
	"google.golang.org/grpc/metadata"

	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/system/exitcode"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
)

func cmdGtest() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `gtest [flags] TEST_CMD [TEST_ARG]...`,
		ShortDesc: "Batch upload results of the gtest execution to ResultSink",
		LongDesc: text.Doc(`
			Reads the gtest test results, converts them to ResultSink native format
      and uploads them to ResultDB via ResultSink.
		`),
		CommandRun: func() subcommands.CommandRun {
			r := &gtestRun{}
			r.baseRun.RegisterGlobalFlags()
			return r
		},
	}
}

type gtestRun struct {
	baseRun
}

func (r *gtestRun) Run(a subcommands.Application, args []string, env subcommands.Env) (ret int) {
	ctx := cli.GetContext(a, r, env)

	if err := r.validate(); err != nil {
		return r.done(err)
	}

	if err := r.initSinkClient(ctx); err != nil {
		return r.done(err)
	}

	err := r.runTestCmd(ctx, args)
	ec, ok := exitcode.Get(err)
	if !ok {
		return r.done(errors.Annotate(err, "result_adapter: test command failed").Err())
	}

	var trs []*sinkpb.TestResult
	trs, err = r.generateTestResults()
	if err != nil {
		return r.done(err)
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", "ResultSink "+r.sinkCtx.AuthToken)
	// TODO(crbug.com/1108016): confirm the command works when reports a large amound of results.
	if _, err := r.sinkC.ReportTestResults(ctx, &sinkpb.ReportTestResultsRequest{TestResults: trs}); err != nil {
		return r.done(err)
	}
	return ec
}

// generateTestResults converts test results from results file to sinkpb.TestResult.
// TODO(crbug.com/1108016): Implement.
func (r *gtestRun) generateTestResults() ([]*sinkpb.TestResult, error) {
	return nil, errors.New("not implemented yet")
}
