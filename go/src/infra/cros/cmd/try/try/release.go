// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package try

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"infra/cros/internal/cmd"
	"infra/cros/internal/gerrit"
	bb "infra/cros/lib/buildbucket"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/errors"
)

func GetCmdRelease(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "release [flags]",
		ShortDesc: "Run a release builder.",
		CommandRun: func() subcommands.CommandRun {
			c := &releaseRun{}
			c.tryRunBase.cmdRunner = cmd.RealCommandRunner{}
			c.tryRunBase.authOpts = authOpts
			c.addDryrunFlag()
			c.addBranchFlag("main")
			c.addProductionFlag()
			c.addPatchesFlag()
			c.addBuildTargetsFlag()
			c.Flags.BoolVar(&c.useProdTests, "prod_tests", false, "Run (production) HW tests even if in staging. "+
				"By default, HW tests are disabled in staging.")
			c.Flags.BoolVar(&c.skipPaygen, "skip_paygen", false, "Skip payload generation. Only supported for staging builds.")
			if flag.NArg() > 1 && flag.Args()[1] == "help" {
				fmt.Printf("Run `cros try help` or `cros try help ${subcomand}` for help.")
				os.Exit(0)
			}
			return c
		},
	}
}

/*
In includeAllAncestors, patchInfo allows for two things

1. Marking a visited index in an array representing a RelatedChanges chain
2. Keep a record of the changeNumber at a given index in the RelatedChanes chain

consider a chain [A, B, C, D] and a patches list [D, B].
after evaluating required changes for D,
once we get to B and realize that it was visited as part of the alignment for D,
we stop and go no further.
*/
type patchInfo struct {
	changeNumber int
	visited      bool
}

// includeAllAncestors includes all ancestors of a patch so that the cherry-pick steps don't fail due to merge conflict.
func includeAllAncestors(ctx context.Context, client gerrit.Client, patches []string) ([]string, error) {
	hostMap := map[string]string{
		"c": "https://chromium-review.googlesource.com",
		"i": "https://chrome-internal-review.googlesource.com",
	}
	patchSpec := regexp.MustCompile(PatchRegexpPattern)
	var patchesWithAncestors []string
	// the `^crrev\.com\/([ci])\/(\d{7,8})` component that indicates whether to use gerrit internal or external.
	// Example: c for the chromium instance for i for internal.
	// rootChangeMap keeps a map of each gerritInstance to a map of parent changeNumbers to their RelatedChanges.
	rootChangeMap := make(map[string]map[int][]patchInfo)
	for gerritInstance := range hostMap {
		rootChangeMap[gerritInstance] = make(map[int][]patchInfo)
	}
	for _, patch := range patches {
		// r.validate() already ensures all patches match the expected regex pattern.
		regexMatch := patchSpec.FindStringSubmatch(patch)
		gerritInstance := regexMatch[1]
		host := hostMap[gerritInstance]
		changeNumber, _ := strconv.Atoi(regexMatch[2])
		// Get the list of relatedChanges for this given patch.
		// Example for crrev.com/c/4279215: [4279218, 4279217, 4279216, 4279215, 4279214, 4279213, 4279212, 4279211, 4279210]}}.
		relatedChanges, err := client.GetRelatedChanges(ctx, host, changeNumber)
		if err != nil {
			return []string{}, errors.Annotate(err, "GetRelatedChanges(ctx, %s, %d):", host, changeNumber).Err()
		}
		numberOfRelatedChanges := len(relatedChanges)
		if numberOfRelatedChanges == 0 {
			if _, ok := rootChangeMap[gerritInstance][changeNumber]; !ok {
				// For a singleton patch with no related changes, the patch alone must be returned.
				rootChangeMap[gerritInstance][changeNumber] = []patchInfo{{changeNumber, true}}
				patchesWithAncestors = append(patchesWithAncestors, formatPatchURL(gerritInstance, changeNumber))
			}
		} else {
			// The rootChange is the oldest in the list of relatedChanges and is at the last position in the slice.
			rootChange := relatedChanges[numberOfRelatedChanges-1].ChangeNumber
			if _, ok := rootChangeMap[gerritInstance][rootChange]; !ok {
				rootChangeMap[gerritInstance][rootChange] = make([]patchInfo, numberOfRelatedChanges)
				// Reverse the list of relatedChanges from oldest to newest when mapped to the root (oldest).
				// Example: {'c': {4279210:[{4279210}, {4279211}, {4279212}, {4279213}, {4279214}, {4279215}, {4279216}, {4279217}, {4279218}]}}.
				for i := numberOfRelatedChanges - 1; i >= 0; i-- {
					rootChangeMap[gerritInstance][rootChange][numberOfRelatedChanges-i-1] = patchInfo{relatedChanges[i].ChangeNumber, false}
				}
			}
			for i, change := range rootChangeMap[gerritInstance][rootChange] {
				if !change.visited {
					patchesWithAncestors = append(patchesWithAncestors, formatPatchURL(gerritInstance, change.changeNumber))
					rootChangeMap[gerritInstance][rootChange][i] = patchInfo{change.changeNumber, true}
				}
				if change.changeNumber == changeNumber {
					break
				}
			}
		}
	}
	return patchesWithAncestors, nil
}

// releaseRun tracks relevant info for a given `try release` run.
type releaseRun struct {
	tryRunBase
	useProdTests bool
	skipPaygen   bool
	// Used for testing purposes. If set, props will be written to this file
	// rather than a temporary one.
	propsFile *os.File
}

// validate validates release-specific args for the command.
func (r *releaseRun) validate() error {
	if r.production {
		if r.skipPaygen {
			return fmt.Errorf("--skip_paygen is not supported for production builds")
		}
	}

	if err := r.tryRunBase.validate(); err != nil {
		return err
	}
	return nil
}

// checkChildrenExist checks that any explicitly requested build targets
// have a builder for the relevant branch.
func (r *releaseRun) checkChildrenExist(ctx context.Context) error {
	if len(r.buildTargets) > 0 {
		builderNames := r.getReleaseBuilderNames()
		bucket := "staging-try"
		if r.production {
			bucket = "release"
		}
		for i, builderName := range builderNames {
			fullBuilderName := fmt.Sprintf("chromeos/%s/%s", bucket, builderName)
			_, err := r.bbClient.GetBuilderInputProps(ctx, fullBuilderName)
			if err != nil && strings.Contains(err.Error(), "not found") {
				return fmt.Errorf(
					"%s is not a valid build target for %s. (If you just "+
						"created the branch, you may need to wait 10-15 min "+
						"for LUCI to pick up the new builders.)",
					r.buildTargets[i], r.branch)
			}
		}
	}
	return nil
}

// Run provides the logic for a `try release` command run.
func (r *releaseRun) innerRun(_ subcommands.Application, _ []string, _ subcommands.Env) int {
	if err := r.validate(); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	if r.production && !r.skipProductionPrompt {
		if yes, err := r.promptYes(); err != nil {
			r.LogErr(err.Error())
			return CmdError
		} else if !yes {
			r.LogOut("Exiting.")
			return Success
		}
	}

	ctx := context.Background()
	if ret, err := r.run(ctx); err != nil {
		r.LogErr(err.Error())
		return ret
	}

	if err := r.checkChildrenExist(ctx); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	propsStruct, err := r.bbClient.GetBuilderInputProps(ctx, r.getReleaseOrchestratorName())
	if err != nil {
		r.LogErr(err.Error())
		if strings.Contains(err.Error(), "not found") {
			if strings.HasPrefix(r.branch, "stabilize-") {
				r.LogErr(fmt.Sprintf("Builder not found, is '%s' defined in stabilize_builders.textpb?", r.branch))
			}
		}
		return CmdError
	}

	// TODO(b/266850767): Remove in 2024.
	// crrev.com/c/4205799 updated `cros try` to track a CIPD ref instead of a
	// speific CIPD version, allowing us to push updates to users. We want to
	// invalidate try builds that (roughly) predated this change.
	// This can be removed after it has baked for a sufficiently long period of
	// time (several quarters).
	if err := bb.SetProperty(propsStruct, "$chromeos/cros_try.supported_build", true); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	if len(r.patches) > 0 {
		// If gerrit patches are set, we include the ancestors in order to avoid conflicts.
		patchListBBArgs, err := includeAllAncestors(ctx, r.gerritClient, r.patches)
		if err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
		// If gerrit patches are set, the orchestrator by default will try to do
		// build planning, which is meaningless for release builds and drops
		// all children. This property skips pruning.
		if err := bb.SetProperty(propsStruct, "$chromeos/build_plan.disable_build_plan_pruning", true); err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
		r.bbAddArgs = append(r.bbAddArgs, patchListToBBAddArgs(patchListBBArgs)...)
	}

	if r.useProdTests {
		if err := bb.SetProperty(propsStruct, "$chromeos/cros_test_plan.use_prod_config", true); err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
	}

	if err := bb.SetProperty(propsStruct, "$chromeos/orch_menu.schedule_public_build", false); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	if r.skipPaygen {
		if err := bb.SetProperty(propsStruct, "$chromeos/orch_menu.skip_paygen", true); err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
	}

	if len(r.buildTargets) > 0 {
		if err := bb.SetProperty(propsStruct, "$chromeos/orch_menu.child_builds", r.getReleaseBuilderNames()); err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
	}

	var propsFile *os.File
	if r.propsFile != nil {
		propsFile = r.propsFile
	} else {
		propsFile, err = os.CreateTemp("", "input_props")
		if err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
	}
	if err := bb.WriteStructToFile(propsStruct, propsFile); err != nil {
		r.LogErr(errors.Annotate(err, "writing input properties to tempfile").Err().Error())
		return UnspecifiedError
	}
	if r.propsFile == nil {
		defer os.Remove(propsFile.Name())
	}
	r.bbAddArgs = append(r.bbAddArgs, "-p", fmt.Sprintf("@%s", propsFile.Name()))

	if err := r.runReleaseOrchestrator(ctx); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	return Success
}

func (r *releaseRun) Run(_ subcommands.Application, _ []string, _ subcommands.Env) int {
	r.stdoutLog = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	r.stderrLog = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	// Do not create a gerritClient for test structs with a mockClient.
	if r.gerritClient == nil {
		if err := r.createGerritClient(r.authOpts); err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
	}

	return r.innerRun(nil, nil, nil)
}

// getReleaseOrchestratorName finds the full name of the release orchestrator matching the try CLI flags.
func (r *releaseRun) getReleaseOrchestratorName() string {
	const project = "chromeos"
	var bucket, builder, stagingPrefix string
	if r.production {
		bucket = "release"
	} else {
		bucket = "staging-try"
		stagingPrefix = "staging-"
	}
	if strings.HasPrefix(r.branch, "release-") {
		builder = fmt.Sprintf("%s%s-orchestrator", stagingPrefix, r.branch)
	} else {
		builder = fmt.Sprintf("%srelease-%s-orchestrator", stagingPrefix, r.branch)
	}
	return fmt.Sprintf("%s/%s/%s", project, bucket, builder)
}

func (r *releaseRun) getReleaseBuilderNames() []string {
	const project = "chromeos"
	var builder, stagingPrefix string
	if !r.production {
		stagingPrefix = "staging-"
	}
	builderNames := make([]string, len(r.buildTargets))
	for i, buildTarget := range r.buildTargets {
		if strings.HasPrefix(r.branch, "release-") {
			builder = fmt.Sprintf("%s%s-%s", stagingPrefix, buildTarget, r.branch)
		} else {
			builder = fmt.Sprintf("%s%s-release-%s", stagingPrefix, buildTarget, r.branch)
		}
		builderNames[i] = builder
	}

	return builderNames
}

// runReleaseOrchestrator creates a release orchestrator build via `bb add`, and reports it to the user.
func (r *releaseRun) runReleaseOrchestrator(ctx context.Context) error {
	orchName := r.getReleaseOrchestratorName()
	_, err := r.bbClient.BBAdd(ctx, r.dryrun, append([]string{orchName}, r.bbAddArgs...)...)
	return err
}
