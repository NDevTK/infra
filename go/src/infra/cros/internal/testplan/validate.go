// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package testplan

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"infra/cros/internal/gerrit"
	"infra/cros/internal/shared"
	"infra/tools/dirmd"

	planpb "go.chromium.org/chromiumos/config/go/test/plan"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// ValidateMapping validates ChromeOS test config in mapping.
func ValidateMapping(
	ctx context.Context,
	authedClient gerrit.Client,
	mapping *dirmd.Mapping,
	repoRoot string,
) error {
	validationFns := []func(context.Context, gerrit.Client, string, string, *planpb.SourceTestPlan) error{
		validateAtLeastOneTestPlanStarlarkFile,
		validatePathRegexps,
		validateStarlarkFileExists,
		validateTemplateParameters,
	}

	// Iterate the mappings in lexicographical order.
	dirs := make([]string, 0, len(mapping.Dirs))
	for dir := range mapping.Dirs {
		dirs = append(dirs, dir)
	}

	sort.Strings(dirs)

	multiError := errors.MultiError{}

	for _, dir := range dirs {
		metadata := mapping.Dirs[dir]
		logging.Infof(ctx, "validating dir %q", dir)
		for _, sourceTestPlan := range metadata.GetChromeos().GetCq().GetSourceTestPlans() {
			for _, fn := range validationFns {
				if err := fn(ctx, authedClient, dir, repoRoot, sourceTestPlan); err != nil {
					multiError = append(multiError, errors.Annotate(err, "validation failed for %s", dir).Err())
				}
			}
		}
	}

	return multiError.AsError()
}

func validateAtLeastOneTestPlanStarlarkFile(_ context.Context, _ gerrit.Client, _, _ string, plan *planpb.SourceTestPlan) error {
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

func validatePathRegexps(ctx context.Context, _ gerrit.Client, dir, repoRoot string, plan *planpb.SourceTestPlan) error {
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

func validateStarlarkFileExists(ctx context.Context, client gerrit.Client, _, _ string, plan *planpb.SourceTestPlan) error {
	for _, file := range plan.GetTestPlanStarlarkFiles() {
		_, err := client.DownloadFileFromGitiles(ctx, file.GetHost(), file.GetProject(), "HEAD", file.GetPath(), shared.LongerOpts)
		if err != nil {
			return fmt.Errorf("failed downloading file %q", file)
		}
	}

	return nil
}

func validateTemplateParameters(ctx context.Context, client gerrit.Client, _, _ string, plan *planpb.SourceTestPlan) error {
	// Get the FieldDescriptor for template_parameters to check whether
	// TemplateParameters has been set for a given TestPlanStarlarkFile.
	templateParametersDesc := (&planpb.SourceTestPlan_TestPlanStarlarkFile{}).
		ProtoReflect().Descriptor().Fields().ByName("template_parameters")
	if templateParametersDesc == nil {
		panic("failed to find template_parameters descriptor")
	}

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

		starlarkContent, err := client.DownloadFileFromGitiles(ctx, file.GetHost(), file.GetProject(), "HEAD", file.GetPath(), shared.LongerOpts)
		if err != nil {
			return fmt.Errorf("failed downloading file %q", file)
		}

		if !(strings.Contains(starlarkContent, "testplan.get_suite_name()") ||
			strings.Contains(starlarkContent, "testplan.get_tag_criteria()")) {
			return fmt.Errorf("file %q is not templated, setting TemplateParameters has no effect", file)
		}
	}

	return nil
}
