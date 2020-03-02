// Copyright 2098 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"

	"io/ioutil"
	"os"
	"path/filepath"

	"infra/cmd/cros_test_platform/internal/autotest/artifacts"
	"infra/cmd/cros_test_platform/internal/autotest/testspec"
	"infra/cmd/cros_test_platform/internal/enumeration"
	"infra/cmd/cros_test_platform/internal/site"

	"github.com/kr/pretty"
	"github.com/maruel/subcommands"
	"go.chromium.org/chromiumos/infra/proto/go/chromite/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/common/logging"
)

// Enumerate is the `enumerate` subcommand implementation.
var Enumerate = &subcommands.Command{
	UsageLine: "enumerate -input_json /path/to/input.json -output_json /path/to/output.json",
	ShortDesc: "Enumerate tasks to execute for given requests.",
	LongDesc: `Enumerate tasks to execute for given requests.

Step input and output is JSON encoded protobuf defined at
https://chromium.googlesource.com/chromiumos/infra/proto/+/master/src/test_platform/steps/enumeration.proto`,
	CommandRun: func() subcommands.CommandRun {
		c := &enumerateRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.Flags.StringVar(&c.inputPath, "input_json", "", "Path that contains JSON encoded test_platform.steps.EnumerationRequests")
		c.Flags.StringVar(&c.outputPath, "output_json", "", "Path where JSON encoded test_platform.steps.EnumerationResponses should be written.")
		c.Flags.BoolVar(&c.debug, "debug", false, "Print debugging information to stderr.")
		return c
	},
}

type enumerateRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags

	inputPath  string
	outputPath string
	debug      bool
}

func (c *enumerateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	err := c.innerRun(a, args, env)
	if err != nil {
		fmt.Fprintf(a.GetErr(), "%s\n", err)
	}
	return exitCode(err)
}

func (c *enumerateRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.processCLIArgs(args); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ctx = setupLogging(ctx)

	taggedRequests, err := c.readRequests()
	if err != nil {
		return err
	}
	if len(taggedRequests) == 0 {
		return errors.Reason("zero requests").Err()
	}

	workspace, err := ioutil.TempDir("", "enumerate")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(workspace)
	}()

	// TODO(crbug.com/1012863) Properly handle recoverable error in some
	// requests. Currently a catastrophic error in any request immediately
	// aborts all requests.
	tms := make(map[string]*api.TestMetadataResponse)
	merr := errors.NewMultiError()
	for t, r := range taggedRequests {
		m := r.GetMetadata().GetTestMetadataUrl()
		if m == "" {
			return errors.Reason("empty request.metadata.test_metadata_url in %s", r).Err()
		}
		gsPath := gs.Path(m)

		w, err := ioutil.TempDir(workspace, "request")
		if err != nil {
			return err
		}

		lp, err := c.downloadArtifacts(ctx, gsPath, w)
		if err != nil {
			return err
		}

		tm, writableErr := computeMetadata(lp, w)
		if writableErr != nil && tm == nil {
			// Catastrophic error. There is no reasonable response to write.
			return writableErr
		}
		tms[t] = tm
		merr = append(merr, writableErr)
	}
	var writableErr error
	if merr.First() != nil {
		writableErr = merr
	}

	resps := make(map[string]*steps.EnumerationResponse)
	merr = errors.NewMultiError()
	for t, r := range taggedRequests {
		if ts, imerr := c.enumerate(tms[t], r); imerr != nil {
			merr = append(merr, annotateEach(imerr, "enumerate %s", t)...)
		} else {
			resps[t] = &steps.EnumerationResponse{AutotestInvocations: ts}
		}
	}

	if c.debug {
		c.debugDump(ctx, taggedRequests, tms, resps, merr)
	}
	if merr.First() != nil {
		return merr
	}
	return c.writeResponsesWithError(resps, writableErr)
}

func annotateEach(imerr errors.MultiError, fmt string, args ...interface{}) errors.MultiError {
	var merr errors.MultiError
	for _, err := range imerr {
		merr = append(merr, errors.Annotate(err, fmt, args...).Err())
	}
	return merr
}

func (c *enumerateRun) debugDump(ctx context.Context, reqs map[string]*steps.EnumerationRequest, tms map[string]*api.TestMetadataResponse, resps map[string]*steps.EnumerationResponse, merr errors.MultiError) {
	logging.Infof(ctx, "## Begin debug dump")

	logging.Infof(ctx, "Errors encountered...")
	if merr.First() != nil {
		for _, e := range merr {
			if e != nil {
				logging.Warningf(ctx, "%s", e)
			}
		}
	}
	logging.Infof(ctx, "###")
	logging.Infof(ctx, "")

	for t := range reqs {
		logging.Infof(ctx, "Tag: %s", t)
		logging.Infof(ctx, "Request: %s", pretty.Sprint(reqs[t]))
		if r, ok := resps[t]; ok {
			logging.Infof(ctx, "Response: %s", pretty.Sprint(r))
		} else {
			logging.Warningf(ctx, "No response for %s", t)
		}
		if m, ok := tms[t]; ok {
			logging.Infof(ctx, "Test Metadata: %s", pretty.Sprint(m))
		} else {
			logging.Warningf(ctx, "No metadata for %s", t)
		}
		logging.Infof(ctx, "")
	}
	logging.Infof(ctx, "## End debug dump")
}

func (c *enumerateRun) processCLIArgs(args []string) error {
	if len(args) > 0 {
		return errors.Reason("have %d positional args, want 0", len(args)).Err()
	}
	if c.inputPath == "" {
		return errors.Reason("-input_json not specified").Err()
	}
	if c.outputPath == "" {
		return errors.Reason("-output_json not specified").Err()
	}
	return nil
}

func (c *enumerateRun) readRequests() (map[string]*steps.EnumerationRequest, error) {
	var rs steps.EnumerationRequests
	if err := readRequest(c.inputPath, &rs); err != nil {
		return nil, err
	}
	return rs.TaggedRequests, nil
}

func (c *enumerateRun) writeResponsesWithError(resps map[string]*steps.EnumerationResponse, err error) error {
	r := &steps.EnumerationResponses{
		TaggedResponses: resps,
	}
	return writeResponseWithError(c.outputPath, r, err)
}

func (c *enumerateRun) gsPath(requests []*steps.EnumerationRequest) (gs.Path, error) {
	if len(requests) == 0 {
		panic("zero requests")
	}

	m := requests[0].GetMetadata().GetTestMetadataUrl()
	if m == "" {
		return "", errors.Reason("empty request.metadata.test_metadata_url in %s", requests[0]).Err()
	}
	for _, r := range requests[1:] {
		o := r.GetMetadata().GetTestMetadataUrl()
		if o != m {
			return "", errors.Reason("mismatched test metadata URLs: %s vs %s", m, o).Err()
		}
	}
	return gs.Path(m), nil
}

func (c *enumerateRun) downloadArtifacts(ctx context.Context, gsDir gs.Path, workspace string) (artifacts.LocalPaths, error) {
	outDir := filepath.Join(workspace, "artifacts")
	if err := os.Mkdir(outDir, 0750); err != nil {
		return artifacts.LocalPaths{}, errors.Annotate(err, "download artifacts").Err()
	}
	client, err := c.newGSClient(ctx)
	if err != nil {
		return artifacts.LocalPaths{}, errors.Annotate(err, "download artifacts").Err()
	}
	lp, err := artifacts.DownloadFromGoogleStorage(ctx, client, gsDir, outDir)
	if err != nil {
		return artifacts.LocalPaths{}, errors.Annotate(err, "download artifacts").Err()
	}
	return lp, err
}

func (c *enumerateRun) newGSClient(ctx context.Context) (gs.Client, error) {
	t, err := newAuthenticatedTransport(ctx, &c.authFlags)
	if err != nil {
		return nil, errors.Annotate(err, "create GS client").Err()
	}
	return gs.NewProdClient(ctx, t)
}

func (c *enumerateRun) enumerate(tm *api.TestMetadataResponse, request *steps.EnumerationRequest) ([]*steps.EnumerationResponse_AutotestInvocation, errors.MultiError) {
	var ts []*steps.EnumerationResponse_AutotestInvocation

	g, err := enumeration.GetForTests(tm.Autotest, request.TestPlan.Test)
	if err != nil {
		return nil, errors.NewMultiError(err)
	}
	ts = append(ts, g...)

	ts = append(ts, enumeration.GetForSuites(tm.Autotest, request.TestPlan.Suite)...)
	ts = append(ts, enumeration.GetForEnumeration(request.TestPlan.GetEnumeration())...)

	if merr := validateEnumeration(ts); merr != nil {
		return nil, merr
	}
	return ts, nil
}

func validateEnumeration(ts []*steps.EnumerationResponse_AutotestInvocation) errors.MultiError {
	if len(ts) == 0 {
		return errors.NewMultiError(errors.Reason("empty enumeration").Err())
	}

	var merr errors.MultiError
	for _, t := range ts {
		if err := validateInvocation(t); err != nil {
			merr = append(merr, errors.Annotate(err, "validate %s", t).Err())
		}
	}
	return errorsOrNil(merr)
}

func errorsOrNil(merr errors.MultiError) errors.MultiError {
	if merr.First() != nil {
		return merr
	}
	return nil
}

func validateInvocation(t *steps.EnumerationResponse_AutotestInvocation) error {
	if t.GetTest().GetName() == "" {
		return errors.Reason("empty name").Err()
	}
	if t.GetTest().GetExecutionEnvironment() == api.AutotestTest_EXECUTION_ENVIRONMENT_UNSPECIFIED {
		return errors.Reason("unspecified execution environment").Err()
	}
	return nil
}

func computeMetadata(localPaths artifacts.LocalPaths, workspace string) (*api.TestMetadataResponse, error) {
	extracted := filepath.Join(workspace, "extracted")
	if err := os.Mkdir(extracted, 0750); err != nil {
		return nil, errors.Annotate(err, "compute metadata").Err()
	}
	if err := artifacts.ExtractControlFiles(localPaths, extracted); err != nil {
		return nil, errors.Annotate(err, "compute metadata").Err()
	}
	return testspec.Get(extracted)
}
