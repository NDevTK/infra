// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/system/exitcode"
	"go.chromium.org/luci/common/system/signals"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/lucictx"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
	"google.golang.org/grpc/metadata"
)

// ExitCodeCommandFailure indicates that a given command failed due to internal errors
// or invalid input parameters.
const ExitCodeCommandFailure = 123

// baseRun provides common command run functionality.
type baseRun struct {
	subcommands.CommandRunBase

	// Flags.
	artifactDir            string
	resultFile             string
	testMetadataFile       string
	trimArtifactPrefix     string
	invocationLinkArtifact string

	sinkCtx       *lucictx.ResultSink
	sinkC         sinkpb.SinkClient
	captureOutput bool
}

type converter func(ctx context.Context, data []byte) ([]*sinkpb.TestResult, error)

func (r *baseRun) RegisterGlobalFlags() {
	r.Flags.StringVar(&r.artifactDir, "artifact-directory", "", text.Doc(`
				Directory of the artifacts. Required.
			`))
	r.Flags.StringVar(&r.resultFile, "result-file", "", text.Doc(`
				Path to the result output file. Required.
			`))
	r.Flags.StringVar(&r.testMetadataFile, "test-metadata-file", "", text.Doc(`
				Path to the CFT test metadata output file. Optional.
			`))
	r.Flags.StringVar(&r.trimArtifactPrefix, "trim-artifact-prefix", "", text.Doc(`
				Artifact Path to be trimmed, e.g. absolute path inside a container. Optional.
			`))
	r.Flags.StringVar(&r.invocationLinkArtifact, "invocation-link-artifacts", "", text.Doc(`
				Name and URL pairs to insert as invocation level artifacts, e.g. stainless_logs=https://stainless/logs/1234,testhaus_logs=https://testhaus/logs/1234
				URLs must be url encoded, so '=', ',' and ' ' are not valid characters. Optional.
	`))
}

// validate validates the command has required flags.
func (r *baseRun) validate() (err error) {
	if r.resultFile == "" {
		return errors.Reason("-result-file is required").Err()
	}
	return nil
}

// initSinkClient initializes the result sink client.
func (r *baseRun) initSinkClient(ctx context.Context) (err error) {
	r.sinkCtx = lucictx.GetResultSink(ctx)
	if r.sinkCtx == nil {
		return errors.Reason("no result sink info found in $LUCI_CONTEXT").Err()
	}

	r.sinkC = sinkpb.NewSinkPRPCClient(&prpc.Client{
		Host:    r.sinkCtx.Address,
		Options: &prpc.Options{Insecure: true},
	})

	return nil
}

// runTestCmd waits for test cmd to complete.
func (r *baseRun) runTestCmd(ctx context.Context, args []string) (output []byte, err error) {

	// Subprocess exiting will unblock result_uploader and will stop soon.
	cmdCtx, cancelCmd := context.WithCancel(ctx)
	defer cancelCmd()
	defer signals.HandleInterrupt(func() {
		logging.Warningf(ctx, "result_uploader: interrupt signal received; killing the subprocess")
		cancelCmd()
	})()

	cmd := exec.CommandContext(cmdCtx, args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	if !r.captureOutput {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	// Launch the command w/o the result_sink section in lucictx, in case the test
	// framework has SinkAPI integrated. If a test binary was launched by result_adapter,
	// the test binary shouldn't be able to talk to the local SinkServer directly.
	exported, err := lucictx.Export(lucictx.SetResultSink(cmdCtx, nil))
	if err != nil {
		return nil, errors.Annotate(err, "failed to export a luci-context w/o result-sink").Err()
	}
	defer exported.Close()
	exported.SetInCmd(cmd)

	if r.captureOutput {
		return cmd.Output()
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.Annotate(err, "cmd.start").Err()
	}

	return nil, cmd.Wait()
}

func (r *baseRun) done(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "result_adapter: %s\n", err)
		return ExitCodeCommandFailure
	}
	return 0
}

func (r *baseRun) run(ctx context.Context, args []string, f converter) (ret int) {
	if err := r.initSinkClient(ctx); err != nil {
		return r.done(err)
	}

	out, err := r.runTestCmd(ctx, args)
	ec, ok := exitcode.Get(err)
	if !ok {
		return r.done(errors.Annotate(err, "test command failed").Err())
	}

	trs, err := f(ctx, out)
	switch {
	case err != nil:
		return r.done(err)
	case len(trs) == 0:
		return ec
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", "ResultSink "+r.sinkCtx.AuthToken)

	// Try to upload invocation link artifacts.
	// Upload before test results so that the links are present even if something goes wrong in the test result upload.
	// We do not abort on error here, as we still want to upload the test results even if we can't upload the links.
	linkArtifacts := map[string]*sinkpb.Artifact{}
	for _, pair := range strings.Split(r.invocationLinkArtifact, ",") {
		name, url, found := strings.Cut(pair, "=")
		if !found {
			logging.Warningf(ctx, "Warning: no '=' in invocation-link-artifacts pair: %q, ignoring", pair)
			continue
		}
		linkArtifacts[name] = &sinkpb.Artifact{
			Body:        &sinkpb.Artifact_Contents{Contents: []byte(url)},
			ContentType: "text/x-uri",
		}
	}
	if _, err := r.sinkC.ReportInvocationLevelArtifacts(ctx, &sinkpb.ReportInvocationLevelArtifactsRequest{Artifacts: linkArtifacts}); err != nil {
		logging.Warningf(ctx, "Warning: failed to upload invocation level link artifacts: %q, err: %v", r.invocationLinkArtifact, err)
	}

	if _, err := r.sinkC.ReportTestResults(ctx, &sinkpb.ReportTestResultsRequest{TestResults: trs}); err != nil {
		return r.done(err)
	}
	return ec
}
