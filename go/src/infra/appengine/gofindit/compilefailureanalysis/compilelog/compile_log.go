// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package compilelogs handles downloading logs for compile failures
package compilelog

import (
	"context"
	"encoding/json"
	"fmt"
	"infra/appengine/gofindit/internal/buildbucket"
	"infra/appengine/gofindit/internal/logdog"
	gfim "infra/appengine/gofindit/model"
	"infra/appengine/gofindit/util"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// GetCompileLogs gets the compile log for a build bucket build
// Returns the ninja log and stdout log
func GetCompileLogs(c context.Context, bbid int64) (*gfim.CompileLogs, error) {
	build, err := buildbucket.GetBuild(c, bbid, &buildbucketpb.BuildMask{
		Fields: &fieldmaskpb.FieldMask{
			Paths: []string{"steps"},
		},
	})
	if err != nil {
		return nil, err
	}
	ninjaUrl := ""
	stdoutUrl := ""
	for _, step := range build.Steps {
		if util.IsCompileStep(step) {
			for _, log := range step.Logs {
				if log.Name == "json.output[ninja_info]" {
					ninjaUrl = log.ViewUrl
				}
				if log.Name == "stdout" {
					stdoutUrl = log.ViewUrl
				}
			}
			break
		}
	}

	ninjaLog := &gfim.NinjaLog{}
	stdoutLog := ""

	// TODO(crbug.com/1295566): Parallelize downloading ninja & stdout logs
	if ninjaUrl != "" {
		log, err := logdog.GetLogFromViewUrl(c, ninjaUrl)
		if err != nil {
			logging.Errorf(c, "Failed to get ninja log: %v", err)
		}
		if err = json.Unmarshal([]byte(log), ninjaLog); err != nil {
			return nil, fmt.Errorf("Failed to unmarshal ninja log %w. Log: %s", err, log)
		}
	}

	if stdoutUrl != "" {
		stdoutLog, err = logdog.GetLogFromViewUrl(c, stdoutUrl)
		if err != nil {
			logging.Errorf(c, "Failed to get stdout log: %v", err)
		}
	}

	if len(ninjaLog.Failures) > 0 || stdoutLog != "" {
		return &gfim.CompileLogs{
			NinjaLog:  ninjaLog,
			StdOutLog: stdoutLog,
		}, nil
	}

	return nil, fmt.Errorf("Could not get compile log from build %d", bbid)
}

func GetFailedTargets(compileLogs *gfim.CompileLogs) []string {
	if compileLogs.NinjaLog == nil {
		return []string{}
	}
	results := []string{}
	for _, failure := range compileLogs.NinjaLog.Failures {
		results = append(results, failure.OutputNodes...)
	}
	return results
}
