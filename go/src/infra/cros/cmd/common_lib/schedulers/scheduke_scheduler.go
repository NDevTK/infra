// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package schedulers

import (
	"context"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/common"
)

// SchedukeScheduler defines a scheduler that schedules request(s) through
// Scheduke.
type SchedukeScheduler struct {
	*AbstractScheduler

	schedukeClient *common.SchedukeClient
}

func NewSchedukeSchedulerr() *SchedukeScheduler {
	absSched := NewAbstractScheduler(SchedukeSchedulerType)
	return &SchedukeScheduler{AbstractScheduler: absSched}
}

func (sc *SchedukeScheduler) Setup(ctx context.Context) error {
	//if sc.schedukeClient == nil {
	//	// TODO: pipe in env
	//	var env string
	//	var local bool
	//	client, err := common.NewSchedukeClient(ctx, env, local)
	//	if err != nil {
	//		return err
	//	}
	//	sc.schedukeClient = client
	//}
	return nil
}

func (sc *SchedukeScheduler) ScheduleRequest(ctx context.Context, req *buildbucketpb.ScheduleBuildRequest, step *build.Step) (*buildbucketpb.Build, error) {
	//schedukeReq, err := common.ScheduleBuildReqToSchedukeReq(req)
	//if err != nil {
	//	return nil, err
	//}
	//createTaskResponse, err := (*sc.schedukeClient).ScheduleExecution(schedukeReq)
	//if err != nil {
	//	return nil, err
	//}
	//// TODO: loop to get BBID, update build state via step
	return nil, nil
}
