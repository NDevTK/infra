// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"

	"infra/cros/internal/cmd"

	"github.com/maruel/subcommands"
	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

func getCmdRetry() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "retry [flags]",
		ShortDesc: "Rerun the specified (release) build.",
		CommandRun: func() subcommands.CommandRun {
			c := &retryRun{}
			c.tryRunBase.cmdRunner = cmd.RealCommandRunner{}
			c.addDryrunFlag()
			c.Flags.StringVar(&c.originalBBID, "bbid", "", "Buildbucket ID of the builder to retry.")
			if flag.NArg() > 1 && flag.Args()[1] == "help" {
				fmt.Printf("Run `cros try help` or `cros try help ${subcomand}` for help.")
				os.Exit(0)
			}
			return c
		},
	}
}

// retryRun tracks relevant info for a given `try retry` run.
type retryRun struct {
	tryRunBase
	originalBBID string
	// Used for testing purposes. If set, props will be written to this file
	// rather than a temporary one.
	propsFile *os.File
}

// validate validates retry-specific args for the command.
func (r *retryRun) validate() error {
	return nil
}

// getRetrySummary gets the retry_summary from the specified build.
func (r *retryRun) getRetrySummary(ctx context.Context, bbid string, outputProps *structpb.Struct, allowEmpty bool) (map[pb.RetryStep]string, error) {
	v, ok := outputProps.AsMap()["retry_summary"]
	if !ok {
		if allowEmpty {
			return map[pb.RetryStep]string{}, nil
		}
		return nil, fmt.Errorf("Could not get `retry_summary` property from %s", bbid)
	}
	retrySummary := reflect.ValueOf(v)
	summary := map[pb.RetryStep]string{}
	for _, k := range retrySummary.MapKeys() {
		step := pb.RetryStep(pb.RetryStep_value[k.Interface().(string)])
		summary[step] = retrySummary.MapIndex(k).Interface().(string)
	}
	return summary, nil
}

type childBuild struct {
	bbid         string
	status       bbpb.Status
	retrySummary map[pb.RetryStep]string
}

// getChildBuildInfo gets information about child builders. The map keys are the
// builder name.
func (r *retryRun) getChildBuildInfo(ctx context.Context, parentBuildOutputProps *structpb.Struct) (map[string]childBuild, error) {
	childBuildBBIDs, ok := parentBuildOutputProps.AsMap()["child_builds"]
	if !ok {
		return nil, fmt.Errorf("Could not get `child_builds` property from %s", r.originalBBID)
	}

	childBuildInfo := map[string]childBuild{}
	for _, v := range childBuildBBIDs.([]interface{}) {
		bbid := v.(string)

		buildData, err := r.GetBuild(ctx, bbid)
		if err != nil {
			return nil, errors.Annotate(err, "Could not get output props for %s", bbid).Err()
		}
		originalBuildInputProps := buildData.GetInput().GetProperties()
		if originalBuildInputProps.AsMap()["recipe"] != "build_release" {
			continue
		}
		originalBuildOutputProps := buildData.GetOutput().GetProperties()

		retrySummary, err := r.getRetrySummary(ctx, bbid, originalBuildOutputProps, true)
		if err != nil {
			return nil, err
		}
		childBuildInfo[buildData.GetBuilder().GetBuilder()] = childBuild{
			bbid:         bbid,
			status:       buildData.GetStatus(),
			retrySummary: retrySummary,
		}

	}
	return childBuildInfo, nil
}

func hasFailedChild(childData map[string]childBuild) bool {
	for _, data := range childData {
		if data.status != bbpb.Status_SUCCESS {
			return true
		}
	}
	return false
}

// getExecStep looks at the retry summary and decides what step we need to pick
// up at during the retry run.
func getExecStep(recipe string, retrySummary map[pb.RetryStep]string) (pb.RetryStep, error) {
	recipeSteps := map[string][]pb.RetryStep{
		"orchestrator": {
			pb.RetryStep_CREATE_BUILDSPEC,
			pb.RetryStep_RUN_CHILDREN,
			pb.RetryStep_LAUNCH_TESTS,
		},
		"build_release": {
			pb.RetryStep_STAGE_ARTIFACTS,
			pb.RetryStep_PUSH_IMAGES,
			pb.RetryStep_DEBUG_SYMBOLS,
			pb.RetryStep_COLLECT_SIGNING,
			pb.RetryStep_PAYGEN,
		},
	}

	steps, ok := recipeSteps[recipe]
	if !ok {
		return pb.RetryStep_UNDEFINED, fmt.Errorf("unsupported recipe \"%s\"", recipe)
	}

	// Make sure that the build we're trying to retry doesn't violate the suffix
	// constraint.
	missingStep := false
	foundFailedStep := pb.RetryStep_UNDEFINED
	for _, step := range steps {
		status, stepRan := retrySummary[step]
		if !stepRan {
		} else if missingStep {
			return pb.RetryStep_UNDEFINED, fmt.Errorf("retry summary is missing step %v but has later ones. Can't retry.", step)
		}
		if status != "SUCCESS" {
			foundFailedStep = step
		} else if foundFailedStep != pb.RetryStep_UNDEFINED {
			return pb.RetryStep_UNDEFINED, fmt.Errorf("step %v failed but a subsequent step (%v) succeeded. Can't retry.", foundFailedStep, step)
		}
	}

	// Return the earliest failed step, or the first one that didn't run.
	for _, step := range steps {
		if status, stepRan := retrySummary[step]; !stepRan || status != "SUCCESS" {
			return step, nil
		}
	}
	// If all the steps succeeded, there's nothing to retry.
	return pb.RetryStep_UNDEFINED, nil
}

// Run provides the logic for a `try retry` command run.
func (r *retryRun) Run(_ subcommands.Application, _ []string, _ subcommands.Env) int {
	r.stdoutLog = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	r.stderrLog = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	if err := r.validate(); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	ctx := context.Background()
	if ret, err := r.run(ctx); err != nil {
		r.LogErr(err.Error())
		return ret
	}

	buildData, err := r.GetBuild(ctx, r.originalBBID)
	if err != nil {
		r.LogErr(err.Error())
		return CmdError
	}
	builder := buildData.GetBuilder()
	originalBuildProps := buildData.GetOutput().GetProperties()

	retrySummary, err := r.getRetrySummary(ctx, r.originalBBID, originalBuildProps, false)
	if err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	propsStruct := buildData.GetInput().GetProperties()
	recipe := propsStruct.AsMap()["recipe"].(string)
	// TODO(b/262388770): Support build_release.
	if recipe != "orchestrator" {
		r.LogErr(fmt.Errorf("unsupported recipe `%s`", recipe).Error())
		return CmdError
	}

	childInfo, err := r.getChildBuildInfo(ctx, originalBuildProps)
	if err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	// Set up propsFile.
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

	checkpointProps := map[string]interface{}{
		"retry":               true,
		"original_build_bbid": r.originalBBID,
	}

	// Set exec_steps.
	execStep, err := getExecStep(recipe, retrySummary)
	if err != nil {
		r.LogErr(err.Error())
		return CmdError
	}
	if recipe == "orchestrator" && hasFailedChild(childInfo) {
		execStep = pb.RetryStep_RUN_FAILED_CHILDREN
	}
	checkpointProps["exec_steps"] = map[string]interface{}{
		"steps": []interface{}{int32(execStep.Number())},
	}
	if err := setProperty(propsStruct, "$chromeos/checkpoint", checkpointProps); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}
	// If we're retrying an orchestrator, try and set builder_exec_steps.
	if recipe == "orchestrator" && execStep == pb.RetryStep_RUN_FAILED_CHILDREN {
		for builder, info := range childInfo {
			if info.status == bbpb.Status_SUCCESS {
				continue
			}
			execStep, err := getExecStep("build_release", info.retrySummary)
			if err != nil {
				r.LogErr(err.Error())
				return CmdError
			}
			steps := map[string]interface{}{
				"steps": []interface{}{int32(execStep.Number())},
			}
			subproperty := fmt.Sprintf("$chromeos/checkpoint.builder_exec_steps.%s", builder)
			if err := setProperty(propsStruct, subproperty, steps); err != nil {
				r.LogErr(err.Error())
				return CmdError
			}
		}
	}

	// Write props to file and launch builder.
	if err := writeStructToFile(propsStruct, propsFile); err != nil {
		r.LogErr(errors.Annotate(err, "writing input properties to tempfile").Err().Error())
		return UnspecifiedError
	}
	if r.propsFile == nil {
		defer os.Remove(propsFile.Name())
	}
	r.bbAddArgs = append(r.bbAddArgs, "-p", fmt.Sprintf("@%s", propsFile.Name()))

	builderName := fmt.Sprintf("%s/%s/%s", builder.GetProject(), builder.GetBucket(), builder.GetBuilder())
	if err := r.BBAdd(ctx, append([]string{builderName}, r.bbAddArgs...)...); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	return Success
}
