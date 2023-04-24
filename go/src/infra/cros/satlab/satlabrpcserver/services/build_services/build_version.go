// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package build_services

import (
	moblabapipb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"

	pb "infra/cros/satlab/satlabrpcserver/proto"
)

type BuildVersion struct {
	Version string
	Status  BuildStatus
}

type BuildStatus int

const (
	AVAILABLE BuildStatus = iota
	FAILED
	RUNNING
	ABORTED
)

var FromGCSBucketBuildStatusMap = map[moblabapipb.Build_BuildStatus]BuildStatus{
	moblabapipb.Build_PASS:    AVAILABLE,
	moblabapipb.Build_FAIL:    FAILED,
	moblabapipb.Build_RUNNING: RUNNING,
	moblabapipb.Build_ABORTED: ABORTED,
}

var ToResponseBuildStatusMap = map[BuildStatus]pb.BuildItem_BuildStatus{
	AVAILABLE: pb.BuildItem_BUILD_STATUS_PASS,
	FAILED:    pb.BuildItem_BUILD_STATUS_FAIL,
	RUNNING:   pb.BuildItem_BUILD_STATUS_RUNNING,
	ABORTED:   pb.BuildItem_BUILD_STATUS_ABORTED,
}
