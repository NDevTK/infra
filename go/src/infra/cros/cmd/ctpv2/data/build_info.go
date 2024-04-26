// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package data

import (
	"go.chromium.org/chromiumos/config/go/test/api"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
)

type BuildRequest struct {
	Key                  string
	ShardNum             int
	ScheduleBuildRequest *buildbucketpb.ScheduleBuildRequest
	OriginalTrReq        *TrRequest
	SuiteInfo            *api.SuiteInfo
	Err                  error
}
