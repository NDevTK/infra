// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"fmt"

	bbpb "go.chromium.org/luci/buildbucket/proto"
)

type BuildInfo struct {
	BBID   string `json:"bbid"`
	Status string `json:"status"`
	Retry  bool   `json:"retry"`
}

type BuilderRun struct {
	Builds     []*BuildInfo `json:"builds"`
	RetryCount int          `json:"retry_count"`
	LastStatus string       `json:"last_status"`
}

type CollectReport struct {
	// Maps builder name to information about the series of retries.
	// Retries are grouped by the original build they're retrying.
	BuilderInfo map[string][]*BuilderRun `json:"builders"`
	RetryCount  int                      `json:"retry_count"`
	ReportError bool                     `json:"report_error"`
}

// recordRetry records a build result.
func (r *CollectReport) recordBuild(build *bbpb.Build, originalBBID string, isRetry bool) {
	// Perform necessary initialization.
	if r.BuilderInfo == nil {
		r.BuilderInfo = map[string][]*BuilderRun{}
	}
	builderName := build.GetBuilder().GetBuilder()
	if _, ok := r.BuilderInfo[builderName]; !ok {
		r.BuilderInfo[builderName] = []*BuilderRun{}
	}

	status := build.GetStatus().String()
	buildInfo := &BuildInfo{
		BBID:   fmt.Sprintf("%d", build.GetId()),
		Status: status,
		Retry:  isRetry,
	}

	var builderRun *BuilderRun
	// Look for existing builder run.
	if len(originalBBID) > 0 {
		for i := range r.BuilderInfo[builderName] {
			builds := r.BuilderInfo[builderName][i].Builds
			if len(builds) > 0 && builds[0].BBID == originalBBID {
				builderRun = r.BuilderInfo[builderName][i]
			}
		}
	}
	if builderRun == nil {
		// Start new builder run.
		builderRun = &BuilderRun{
			Builds: []*BuildInfo{},
		}
		r.BuilderInfo[builderName] = append(r.BuilderInfo[builderName], builderRun)
	}

	builderRun.Builds = append(builderRun.Builds, buildInfo)
	builderRun.LastStatus = status
	if isRetry {
		builderRun.RetryCount += 1
		r.RetryCount += 1
	}
}
