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
	"strings"

	"infra/cros/internal/cmd"

	"github.com/maruel/subcommands"
	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	recipeSteps = map[string][]pb.RetryStep{
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
)

func getCmdRetry() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "retry [flags]",
		ShortDesc: "(Experimental) Rerun the specified (release) build.",
		CommandRun: func() subcommands.CommandRun {
			c := &retryRun{}
			c.tryRunBase.cmdRunner = cmd.RealCommandRunner{}
			c.addDryrunFlag()
			c.Flags.BoolVar(&c.paygenRetry, "paygen", false, "If set, only retries paygen.")
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
	paygenRetry  bool
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

// getSigningSummary gets the signing_summary from the specified build.
func (r *retryRun) getSigningSummary(ctx context.Context, bbid string, outputProps *structpb.Struct, allowEmpty bool) (map[string]string, error) {
	v, ok := outputProps.AsMap()["signing_summary"]
	if !ok {
		if allowEmpty {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("Could not get `signing_summary` property from %s", bbid)
	}
	signingSummary := reflect.ValueOf(v)
	summary := map[string]string{}
	for _, k := range signingSummary.MapKeys() {
		summary[k.Interface().(string)] = signingSummary.MapIndex(k).Interface().(string)
	}
	return summary, nil
}

type buildInfo struct {
	bbid           string
	status         bbpb.Status
	retrySummary   map[pb.RetryStep]string
	signingSummary map[string]string
}

// getChildBuildInfo gets information about child builders. The map keys are the
// builder name.
func (r *retryRun) getChildBuildInfo(ctx context.Context, parentBuildOutputProps *structpb.Struct) (map[string]buildInfo, error) {
	childBuildBBIDs, ok := parentBuildOutputProps.AsMap()["child_builds"]
	if !ok {
		return nil, fmt.Errorf("Could not get `child_builds` property from %s", r.originalBBID)
	}

	childBuildInfo := map[string]buildInfo{}
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
		signingSummary, err := r.getSigningSummary(ctx, bbid, originalBuildOutputProps, true)
		if err != nil {
			return nil, err
		}
		childBuildInfo[buildData.GetBuilder().GetBuilder()] = buildInfo{
			bbid:           bbid,
			status:         buildData.GetStatus(),
			retrySummary:   retrySummary,
			signingSummary: signingSummary,
		}

	}
	return childBuildInfo, nil
}

func hasFailedChild(childData map[string]buildInfo) bool {
	for _, data := range childData {
		if data.status != bbpb.Status_SUCCESS {
			return true
		}
	}
	return false
}

// getExecStep looks at the retry summary and decides what step we need to pick
// up at during the retry run.
func getExecStep(recipe string, buildData buildInfo) (pb.RetryStep, error) {
	steps, ok := recipeSteps[recipe]
	if !ok {
		return pb.RetryStep_UNDEFINED, fmt.Errorf("unsupported recipe \"%s\"", recipe)
	}

	// Make sure that the build we're trying to retry doesn't violate the suffix
	// constraint.
	missingStep := false
	foundFailedStep := pb.RetryStep_UNDEFINED
	for _, step := range steps {
		status, stepRan := buildData.retrySummary[step]
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

	// If there are signing failures, start at PUSH_IMAGES to rekick signing.
	// The suffix constraint steps above will prevent us from skipping earlier
	// steps that didn't pass in the previous build, and we know that if
	// signing_summary is set then we at least got to COLLECT_SIGNING in the
	// previous build. Everything between PUSH_IMAGES and COLLECT_SIGNING
	// (currently just DEBUG_SYMBOLS) can be rerun without consequence /
	// clobbering (if this changes we'll need to tweak this approach).
	if recipe == "build_release" && len(buildData.signingSummary) > 0 {
		for _, status := range buildData.signingSummary {
			if status == "FAILED" || status == "TIMED_OUT" {
				return pb.RetryStep_PUSH_IMAGES, nil
			}
		}
	}

	// Return the earliest failed step, or the first one that didn't run.
	for _, step := range steps {
		if status, stepRan := buildData.retrySummary[step]; !stepRan || status != "SUCCESS" {
			return step, nil
		}
	}
	// If all the steps succeeded, there's nothing to retry.
	// If the build is successful, this isn't a problem (this function shouldn't
	// really be getting called anyways). If the build failed, we should fail
	// too.
	if buildData.status != bbpb.Status_SUCCESS {
		return pb.RetryStep_UNDEFINED, fmt.Errorf("build %v was not successful but all retry steps passed, not sure what to retry", buildData.bbid)
	}
	return pb.RetryStep_UNDEFINED, nil
}

// Process a standard retry.
func (r *retryRun) processRetry(ctx context.Context, buildData *bbpb.Build, propsStruct *structpb.Struct) int {
	originalBuildProps := buildData.GetOutput().GetProperties()

	retrySummary, err := r.getRetrySummary(ctx, r.originalBBID, originalBuildProps, false)
	if err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	recipe := propsStruct.AsMap()["recipe"].(string)
	if recipe != "orchestrator" && recipe != "build_release" {
		r.LogErr(fmt.Errorf("unsupported recipe `%s`", recipe).Error())
		return CmdError
	}

	checkpointProps := map[string]interface{}{
		"retry":               true,
		"original_build_bbid": r.originalBBID,
	}

	// Set exec_steps.
	execStep, err := getExecStep(recipe, buildInfo{
		bbid:         r.originalBBID,
		status:       buildData.GetStatus(),
		retrySummary: retrySummary,
	})
	if err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	var childInfo map[string]buildInfo
	if recipe == "orchestrator" {
		childInfo, err = r.getChildBuildInfo(ctx, originalBuildProps)
		if err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
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
	if err := setProperty(propsStruct, "$chromeos/signing.ignore_already_exists_errors", true); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}
	// If we're retrying an orchestrator, try and set builder_exec_steps.
	if recipe == "orchestrator" && execStep == pb.RetryStep_RUN_FAILED_CHILDREN {
		for builder, info := range childInfo {
			if info.status == bbpb.Status_SUCCESS {
				continue
			}
			execStep, err := getExecStep("build_release", info)
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
	return Success
}

// Process a paygen retry.
func (r *retryRun) processPaygenRetry(ctx context.Context, buildData *bbpb.Build, propsStruct *structpb.Struct) int {
	recipe := propsStruct.AsMap()["recipe"].(string)
	if recipe != "build_release" {
		r.LogErr("build is not a `build_release` build, can't retry with --paygen")
		return CmdError
	}

	retrySummary, err := r.getRetrySummary(ctx, r.originalBBID, buildData.GetOutput().GetProperties(), true)
	if err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	buildStatus := buildData.GetStatus()
	if len(retrySummary) == 0 {
		// If we don't have a retry summary, we can only retry a successful build.
		if buildStatus != bbpb.Status_SUCCESS {
			r.LogErr("no `retry_summary` and build was unsuccessful, can't retry paygen")
			return CmdError
		}
	} else {
		// If we do have a retry summary, everything before PAYGEN must be successful.
		for _, step := range recipeSteps["build_release"] {
			if step == pb.RetryStep_PAYGEN {
				break
			}
			status, stepRan := retrySummary[step]
			if !stepRan {
				r.LogErr("build did not run step %v, can't retry paygen", step)
				return CmdError
			} else if status != "SUCCESS" {
				r.LogErr("step %v failed, can't retry paygen", step)
				return CmdError
			}
		}
	}

	checkpointProps := map[string]interface{}{
		"retry":               true,
		"original_build_bbid": r.originalBBID,
		"exec_steps": map[string]interface{}{
			"steps": []interface{}{int32(pb.RetryStep_PAYGEN.Number())},
		},
	}
	if err := setProperty(propsStruct, "$chromeos/checkpoint", checkpointProps); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}
	return Success
}

// Run provides the logic for a `try retry` command run.
func (r *retryRun) Run(_ subcommands.Application, _ []string, _ subcommands.Env) int {
	r.stdoutLog = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	r.stderrLog = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	if err := r.validate(); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}
	// Allow the "b" suffix on bbids.
	r.originalBBID = strings.TrimPrefix(r.originalBBID, "b")

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
	propsStruct := buildData.GetInput().GetProperties()

	if r.paygenRetry {
		ret := r.processPaygenRetry(ctx, buildData, propsStruct)
		if ret != Success {
			return ret
		}
	} else {
		ret := r.processRetry(ctx, buildData, propsStruct)
		if ret != Success {
			return ret
		}
	}

	// Write props to file and launch builder.
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
	if err := writeStructToFile(propsStruct, propsFile); err != nil {
		r.LogErr(errors.Annotate(err, "writing input properties to tempfile").Err().Error())
		return UnspecifiedError
	}
	if r.propsFile == nil {
		defer os.Remove(propsFile.Name())
	}
	r.bbAddArgs = append(r.bbAddArgs, "-p", fmt.Sprintf("@%s", propsFile.Name()))

	builder := buildData.GetBuilder()
	builderName := fmt.Sprintf("%s/%s/%s", builder.GetProject(), builder.GetBucket(), builder.GetBuilder())
	if err := r.BBAdd(ctx, append([]string{builderName}, r.bbAddArgs...)...); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	return Success
}
