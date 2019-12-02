// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/hex"
	"time"

	"go.chromium.org/luci/auth"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/data/rand/cryptorand"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/gcloud/googleoauth"
	logdog_types "go.chromium.org/luci/logdog/common/types"
)

// SwarmingHostname retrieves the Swarming hostname from the JobDefinition.
//
// This value occurs in different places, depending on if the JobDefinition is
// in buildbucket or swarming mode.
func (jd *JobDefinition) SwarmingHostname() string {
	if jd.GetBuildbucket() != nil {
		return jd.GetBuildbucket().GetBbagentArgs().GetBuild().GetInfra().GetSwarming().GetHostname()
	}
	return jd.GetSwarming().GetHostname()
}

// TaskName retrieves the human-readable Swarming task name from the
// JobDefinition.
//
// This value occurs in different places, depending on if the JobDefinition is
// in buildbucket or swarming mode.
func (jd *JobDefinition) TaskName() string {
	if jd.GetBuildbucket() != nil {
		return jd.GetBuildbucket().GetName()
	}
	return jd.GetSwarming().GetTask().GetName()
}

func (jd *JobDefinition) addLedProperties(ctx context.Context, uid string) (logdogPrefix string, err error) {
	// Set the "$recipe_engine/led" recipe properties.
	buf := make([]byte, 32)
	if _, err := cryptorand.Read(ctx, buf); err != nil {
		return "", errors.Annotate(err, "generating random token").Err()
	}
	streamName, err := logdog_types.MakeStreamName("", "led", uid, hex.EncodeToString(buf))
	if err != nil {
		return "", errors.Annotate(err, "generating logdog token").Err()
	}
	logdogPrefix = string(streamName)

	err = jd.EditBuildbucket(func(ejd *EditBBJobDefinition) {
		ejd.tweak(func() error {
			// Pass the CIPD package or isolate containing the recipes code into
			// the led recipe module. This gives the build the information it needs
			// to launch child builds using the same version of the recipes code.
			ledProperties := map[string]interface{}{
				// The logdog prefix is unique to each led job, so it can be used as an
				// ID for the job.
				"led_run_id": string(logdogPrefix),
			}
			if dgst := jd.GetUserPayload().GetDigest(); dgst != "" {
				ledProperties["isolated_input"] = map[string]interface{}{
					// TODO(iannucci): Set server and namespace too.
					"hash": dgst,
				}
			} else if pkg := ejd.bb.GetBbagentArgs().GetBuild().GetExe(); pkg != nil {
				ledProperties["cipd_input"] = map[string]interface{}{
					"package": pkg.GetCipdPackage(),
					"version": pkg.GetCipdVersion(),
				}
			}

			ejd.bb.WriteProperties(map[string]interface{}{
				"$recipe_engine/led": ledProperties,
			})

			return nil
		})
		return
	})
	return
}

func getUIDFromAuth(ctx context.Context, authOpts auth.Options) (string, error) {
	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts)
	tok, err := authenticator.GetAccessToken(time.Minute)
	if err != nil {
		return "", errors.Annotate(err, "making authenticator").Err()
	}
	info, err := googleoauth.GetTokenInfo(ctx, googleoauth.TokenInfoParams{
		AccessToken: tok.AccessToken,
	})
	if err != nil {
		return "", errors.Annotate(err, "extracting token info").Err()
	}
	uid := info.Email
	if uid == "" {
		uid = "uid:" + info.Sub
	}
	return uid, nil
}

// FlattenToSwarming modifies this JobDefinition to populate the Swarming field
// from the Buildbucket field.
//
// After flattening, buildbucket-only edit functionality will no longer work.
func (jd *JobDefinition) FlattenToSwarming(ctx context.Context, authOpts auth.Options) error {
	if jd.GetSwarming() != nil {
		return nil
	}

	// Adjust bbargs to use "$recipeCheckoutDir/luciexe" as ExecutablePath.

	// TODO(iannucci): content
	// setOutputResultPath tells the task to write its result to a JSON file in the ISOLATED_OUTDIR.
	// This is an interface for getting structured output from a task launched by led.
	//func setOutputResultPath(s *Systemland) error {
	//	ka := s.KitchenArgs
	//	if ka == nil {
	//		// TODO(iannucci): Support LUCI runner.
	//		// Intentionally not fatal. led supports jobs which don't use kitchen or LUCI runner,
	//		// and we don't want to block that usage.
	//	return nil
	//	}
	//	ka.OutputResultJSONPath = "${ISOLATED_OUTDIR}/build.proto.json"
	//	return nil
	//}

	uid, err := getUIDFromAuth(ctx, authOpts)
	if err != nil {
		return errors.Annotate(err, "getting user ID").Err()
	}

	logdogPrefix, err := jd.addLedProperties(ctx, uid)
	if err != nil {
		return errors.Annotate(err, "adding led properties").Err()
	}

	// generate "log_location:logdog://" and "allow_milo:1" tags
	// generate "recipe_package:" and "recipe_name:" tags

	panic("implement")
}

// ToSwarmingNewTask renders a (swarming) JobDefinition to
// a SwarmingRpcsNewTaskRequest.
//
// If you call this on something other than a swarming JobDefinition, it will
// panic.
func (jd *JobDefinition) ToSwarmingNewTask(ctx context.Context, authOpts auth.Options) (*swarming.SwarmingRpcsNewTaskRequest, error) {
	if err := jd.FlattenToSwarming(ctx, authOpts); err != nil {
		return nil, errors.Annotate(err, "flattening to swarming JobDefinition").Err()
	}

	// TODO(iannucci): set "User" top level property to uid

	panic("implement")
}

// GetCurrentIsolated returns the current isolated contents for the
// JobDefinition.
//
// Supports:
//   Buildbucket JobDefinitions with UserPayload set.
//   Swarming JobDefinitions with UserPayload set.
//   Swarming JobDefinitions where all the slices have the same CasInput.
//
// Returns error for unuspported JobDefinition.
func (jd *JobDefinition) GetCurrentIsolated() (string, error) {
	isolatedOptions := stringset.New(1)
	isolatedOptions.Add(jd.GetUserPayload().GetDigest())

	if sw := jd.GetSwarming(); sw != nil {
		for _, slc := range sw.GetTask().GetTaskSlices() {
			input := slc.GetProperties().GetCasInputs()
			if input != nil {
				isolatedOptions.Add(input.Digest)
			}
		}
	}
	isolatedOptions.Del("") // don't care about empty string
	if isolatedOptions.Len() > 1 {
		return "", errors.Reason(
			"JobDefinition contains multiple isolateds: %v", isolatedOptions.ToSlice()).Err()
	}
	ret, _ := isolatedOptions.Pop()
	return ret, nil
}

// ClearCurrentIsolated removes all isolateds from this JobDefinition.
func (jd *JobDefinition) ClearCurrentIsolated() {
	jd.UserPayload = nil
	for _, slc := range jd.GetSwarming().GetTask().GetTaskSlices() {
		if input := slc.GetProperties().GetCasInputs(); input != nil {
			input.Digest = ""
		}
	}
}
