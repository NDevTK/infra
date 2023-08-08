// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package testplan

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"infra/cros/internal/cmd"
	"infra/cros/internal/docker"
	"infra/cros/internal/gerrit"
	"infra/cros/internal/shared"
	"infra/tools/dirmd"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"go.chromium.org/chromiumos/config/go/build/api"
	testpb "go.chromium.org/chromiumos/config/go/test/api"
	planpb "go.chromium.org/chromiumos/config/go/test/plan"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/sync/parallel"
	"google.golang.org/protobuf/encoding/protojson"
)

type validator struct {
	gerritClient gerrit.Client
	bbClient     bbpb.BuildsClient
	cmdRunner    cmd.CommandRunner
	tmpdirFn     func(string, string) (string, error)
}

// NewValidator returns a validator with default configuration, which can be
// used to validate ChromeOS test configs in dirmd.Mappings.
func NewValidator(gerritClient gerrit.Client, bbClient bbpb.BuildsClient, cmdRunner cmd.CommandRunner) *validator {
	return &validator{
		gerritClient: gerritClient,
		bbClient:     bbClient,
		cmdRunner:    cmdRunner,
		tmpdirFn:     os.MkdirTemp,
	}
}

// WithTmpdirFn overrides the function the validator uses to create temp dirs;
// for example when the validator calls Docker containers it binds a temp dir
// on the host to the image for I/O.
//
// Modifies the validator and also returns a pointer to it.
//
// By default the tempdir function is os.MkdirTemp, and the new tmpdir fn must
// have the same signature. The most common reason to override the tempdir
// function is for injecting test data.
func (v *validator) WithTmpdirFn(f func(string, string) (string, error)) *validator {
	v.tmpdirFn = f
	return v
}

// ValidateMapping validates ChromeOS test config in mapping.
func (v *validator) ValidateMapping(
	ctx context.Context,
	mapping *dirmd.Mapping,
	repoRoot string,
) error {
	validationFns := []func(context.Context, string, string, *planpb.SourceTestPlan) error{
		v.validateAtLeastOneTestPlanStarlarkFile,
		v.validatePathRegexps,
		v.validateStarlarkFileExists,
		v.validateTemplateParameters,
	}

	return parallel.WorkPool(0, func(c chan<- func() error) {
		for dir, metadata := range mapping.Dirs {
			dir := dir
			metadata := metadata
			logging.Infof(ctx, "validating dir %q", dir)

			for _, sourceTestPlan := range metadata.GetChromeos().GetCq().GetSourceTestPlans() {
				for _, fn := range validationFns {
					sourceTestPlan := sourceTestPlan
					fn := fn
					c <- func() error {
						return fn(ctx, dir, repoRoot, sourceTestPlan)
					}
				}
			}
		}
	})
}

func (v *validator) validateAtLeastOneTestPlanStarlarkFile(_ context.Context, _, _ string, plan *planpb.SourceTestPlan) error {
	if len(plan.GetTestPlanStarlarkFiles()) == 0 {
		return fmt.Errorf("at least one TestPlanStarlarkFile must be specified")
	}

	for _, file := range plan.GetTestPlanStarlarkFiles() {
		if !strings.HasSuffix(file.GetPath(), ".star") {
			return fmt.Errorf("all TestPlanStarlarkFile must specify \".star\" files, got %q", file.GetPath())
		}
	}

	return nil
}

func (v *validator) validatePathRegexps(ctx context.Context, dir, repoRoot string, plan *planpb.SourceTestPlan) error {
	for _, pattern := range append(plan.PathRegexps, plan.PathRegexpExcludes...) {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return errors.Annotate(err, "failed to compile path regexp %q", pattern).Err()
		}

		if dir != "." && !strings.HasPrefix(pattern, dir) {
			return fmt.Errorf(
				"path_regexp(_exclude)s defined in a directory that is not "+
					"the root of the repo must have the sub-directory as a prefix. "+
					"Invalid regexp %q in directory %q",
				pattern, dir,
			)
		}

		matchedPath := false
		if err := filepath.WalkDir(filepath.Join(repoRoot, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if re.Match([]byte(path)) {
				logging.Debugf(ctx, "found match for pattern %q: %q", pattern, path)
				matchedPath = true
				return fs.SkipAll
			}

			return nil
		}); err != nil {
			return err
		}

		if !matchedPath {
			logging.Warningf(ctx, "pattern %q doesn't match any files in directory %q", pattern, dir)
		}
	}

	return nil
}

func (v *validator) validateStarlarkFileExists(ctx context.Context, _, _ string, plan *planpb.SourceTestPlan) error {
	for _, file := range plan.GetTestPlanStarlarkFiles() {
		_, err := v.gerritClient.DownloadFileFromGitiles(ctx, file.GetHost(), file.GetProject(), "HEAD", file.GetPath(), shared.LongerOpts)
		if err != nil {
			return fmt.Errorf("failed downloading file %q", file)
		}
	}

	return nil
}

// getMostRecentCTFImage returns <name:tag> for the most recent cros-test-finder
// image produced by amd64-generic-postsubmit.
func (v *validator) getMostRecentCTFImage(ctx context.Context) (string, error) {
	bbResp, err := v.bbClient.SearchBuilds(ctx, &bbpb.SearchBuildsRequest{
		Predicate: &bbpb.BuildPredicate{
			Builder: &bbpb.BuilderID{
				Project: "chromeos",
				Bucket:  "postsubmit",
				Builder: "amd64-generic-postsubmit",
			},
			Status: bbpb.Status_SUCCESS,
			Tags:   []*bbpb.StringPair{{Key: "relevance", Value: "relevant"}},
		},
		PageSize: 1,
	})

	if err != nil {
		return "", fmt.Errorf("SearchBuilds failed: %w", err)
	}

	if len(bbResp.Builds) != 1 {
		return "", fmt.Errorf("expected exactly one build from SearchBuilds, got %q", bbResp)
	}

	return fmt.Sprintf("us-docker.pkg.dev/cros-registry/test-services/cros-test-finder:%d", bbResp.Builds[0].Id), nil
}

// callCrosTestFinder runs Docker on ctfImage with req as input and returns the
// response.
func (v *validator) callCrosTestFinder(
	ctx context.Context,
	ctfImage string,
	req *testpb.CrosTestFinderRequest,
) (*testpb.CrosTestFinderResponse, error) {
	tmpDir, err := v.tmpdirFn("", "ctf*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	// Write the request to a file and create a file for the response.
	reqJson, err := protojson.Marshal(req)
	if err != nil {
		return nil, err
	}

	reqFilePath := filepath.Join(tmpDir, "request.json")
	if err = os.WriteFile(reqFilePath, reqJson, os.ModePerm); err != nil {
		return nil, err
	}

	respFilePath := filepath.Join(tmpDir, "result.json")

	var stderrBuf bytes.Buffer
	logging.Debugf(ctx, "running image %q", ctfImage)
	if err := docker.RunContainer(
		ctx, v.cmdRunner,
		&container.Config{
			Image: ctfImage,
			Cmd:   strslice.StrSlice{"cros-test-finder"},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: tmpDir,
					Target: "/tmp/test/cros-test-finder",
				},
			},
		},
		&api.ContainerImageInfo{
			Repository: &api.GcrRepository{
				Hostname: "us-docker.pkg.dev",
				Project:  "cros-registry/test-services",
			},
			Name: "cros-test-finder",
		},
		&docker.RuntimeOptions{
			UseConfigureDocker: true,
			NoSudo:             true,
			StdoutBuf:          io.Discard,
			StderrBuf:          &stderrBuf,
		},
	); err != nil {
		logging.Errorf(ctx, "cros-test-finder failed, stderr: %v", stderrBuf.String())
		return nil, err
	}

	respJson, err := os.ReadFile(respFilePath)
	if err != nil {
		return nil, err
	}

	resp := &testpb.CrosTestFinderResponse{}
	if err := protojson.Unmarshal(respJson, resp); err != nil {
		return nil, fmt.Errorf("error unmarshalling proto read from %q: %w", respFilePath, err)
	}
	return resp, nil
}

// checkTagCriteriaNonEmpty uses ctfImage to check that the TestCaseTagCriteria
// in templateParameters match at least one test.
func (v *validator) checkTagCriteriaNonEmpty(
	ctx context.Context,
	ctfImage string,
	templateParameters *planpb.SourceTestPlan_TestPlanStarlarkFile_TemplateParameters,
) error {
	suiteName := templateParameters.GetSuiteName()
	logging.Infof(ctx, "checking tag criteria for suite %q are non-empty", suiteName)
	if strings.HasPrefix(suiteName, "tast_gce") || strings.HasPrefix(suiteName, "tast_vm") {
		logging.Warningf(ctx, "cros-test-finder doesn't currently work with VM suites (%s), skipping validation", suiteName)
		return nil
	}

	ctfReq := &testpb.CrosTestFinderRequest{
		TestSuites: []*testpb.TestSuite{
			{
				Name: suiteName,
				Spec: &testpb.TestSuite_TestCaseTagCriteria_{
					TestCaseTagCriteria: templateParameters.GetTagCriteria(),
				},
			},
		},
	}

	ctfResp, err := v.callCrosTestFinder(
		ctx, ctfImage, ctfReq,
	)
	if err != nil {
		return fmt.Errorf("error calling cros-test-finder: %w", err)
	}

	for _, testSuite := range ctfResp.GetTestSuites() {
		if len(testSuite.GetTestCases().GetTestCases()) == 0 {
			return fmt.Errorf("no test cases found for test suite %q", testSuite)
		}
	}

	logging.Infof(ctx, "found test cases for suite %q", suiteName)

	return nil
}

func (v *validator) validateTemplateParameters(ctx context.Context, _, _ string, plan *planpb.SourceTestPlan) error {
	// Get the FieldDescriptor for template_parameters to check whether
	// TemplateParameters has been set for a given TestPlanStarlarkFile.
	templateParametersDesc := (&planpb.SourceTestPlan_TestPlanStarlarkFile{}).
		ProtoReflect().Descriptor().Fields().ByName("template_parameters")
	if templateParametersDesc == nil {
		panic("failed to find template_parameters descriptor")
	}

	// Only get the cros-test-finder image if it is actually needed (i.e. if a
	// file uses TemplateParameters). This prevents unnecessary calls to
	// Buildbucket.
	var ctfImage string

	for _, file := range plan.GetTestPlanStarlarkFiles() {
		if !file.ProtoReflect().Has(templateParametersDesc) {
			continue
		}

		templateParameters := file.GetTemplateParameters()
		if templateParameters.GetSuiteName() == "" {
			return errors.New("suite_name must not be empty")
		}

		tagExcludes := templateParameters.GetTagCriteria().GetTagExcludes()
		if !stringset.NewFromSlice(tagExcludes...).Has("informational") {
			return fmt.Errorf(`tag_excludes must exclude "informational", got %q`, tagExcludes)
		}

		starlarkContent, err := v.gerritClient.DownloadFileFromGitiles(ctx, file.GetHost(), file.GetProject(), "HEAD", file.GetPath(), shared.LongerOpts)
		if err != nil {
			return fmt.Errorf("failed downloading file %q", file)
		}

		if !(strings.Contains(starlarkContent, "testplan.get_suite_name()") ||
			strings.Contains(starlarkContent, "testplan.get_tag_criteria()")) {
			return fmt.Errorf("file %q is not templated, setting TemplateParameters has no effect", file)
		}

		if ctfImage == "" {
			ctfImage, err = v.getMostRecentCTFImage(ctx)
			if err != nil {
				return fmt.Errorf("failed to find recent cros-test-finder image: %w", err)
			}
		}

		if err = v.checkTagCriteriaNonEmpty(ctx, ctfImage, templateParameters); err != nil {
			return err
		}
	}

	return nil
}
