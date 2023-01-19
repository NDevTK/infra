// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import bbpb "go.chromium.org/luci/buildbucket/proto"

type BuildInfo struct {
	BBID   int64  `json:"bbid"`
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
	BuilderInfo map[string]*BuilderRun `json:"builders"`
	RetryCount  int                    `json:"retry_count"`
}

// recordRetry records a build result.
func (r *CollectReport) recordBuild(build *bbpb.Build, isRetry bool) {
	// Perform necessary initialization.
	if r.BuilderInfo == nil {
		r.BuilderInfo = map[string]*BuilderRun{}
	}
	builderName := build.GetBuilder().GetBuilder()
	if _, ok := r.BuilderInfo[builderName]; !ok {
		r.BuilderInfo[builderName] = &BuilderRun{
			Builds: []*BuildInfo{},
		}
	}

	status := build.GetStatus().String()
	r.BuilderInfo[builderName].Builds = append(r.BuilderInfo[builderName].Builds, &BuildInfo{
		BBID:   build.GetId(),
		Status: status,
		Retry:  isRetry,
	})
	r.BuilderInfo[builderName].LastStatus = status
	if isRetry {
		r.BuilderInfo[builderName].RetryCount += 1
		r.RetryCount += 1
	}
}
