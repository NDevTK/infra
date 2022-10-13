package testplan

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"infra/cros/internal/gerrit"
	"infra/cros/internal/shared"
	"infra/tools/dirmd"

	planpb "go.chromium.org/chromiumos/config/go/test/plan"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// ValidateMapping validates ChromeOS test config in mapping.
func ValidateMapping(ctx context.Context, authedClient gerrit.Client, mapping *dirmd.Mapping) error {
	validationFns := []func(context.Context, gerrit.Client, string, *planpb.SourceTestPlan) error{
		validateAtLeastOneTestPlanStarlarkFile,
		validatePathRegexps,
		validateStarlarkFileExists,
	}

	for dir, metadata := range mapping.Dirs {
		logging.Infof(ctx, "validating dir %q", dir)
		for _, sourceTestPlan := range metadata.GetChromeos().GetCq().GetSourceTestPlans() {
			for _, fn := range validationFns {
				if err := fn(ctx, authedClient, dir, sourceTestPlan); err != nil {
					return errors.Annotate(err, "validation failed for %s", dir).Err()
				}
			}
		}
	}

	return nil
}

func validateAtLeastOneTestPlanStarlarkFile(_ context.Context, _ gerrit.Client, _ string, plan *planpb.SourceTestPlan) error {
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

func validatePathRegexps(_ context.Context, _ gerrit.Client, dir string, plan *planpb.SourceTestPlan) error {
	for _, re := range append(plan.PathRegexps, plan.PathRegexpExcludes...) {
		if _, err := regexp.Compile(re); err != nil {
			return errors.Annotate(err, "failed to compile path regexp %q", re).Err()
		}

		if dir != "." && !strings.HasPrefix(re, dir) {
			return fmt.Errorf(
				"path_regexp(_exclude)s defined in a directory that is not "+
					"the root of the repo must have the sub-directory as a prefix. "+
					"Invalid regexp %q in directory %q",
				re, dir,
			)
		}
	}

	return nil
}

func validateStarlarkFileExists(ctx context.Context, client gerrit.Client, _ string, plan *planpb.SourceTestPlan) error {
	for _, file := range plan.GetTestPlanStarlarkFiles() {
		_, err := client.DownloadFileFromGitiles(ctx, file.GetHost(), file.GetProject(), "HEAD", file.GetPath(), shared.LongerOpts)
		if err != nil {
			return fmt.Errorf("failed downloading file %q", file)
		}
	}

	return nil
}
