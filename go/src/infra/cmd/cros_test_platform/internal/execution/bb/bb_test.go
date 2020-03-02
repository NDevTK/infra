// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bb

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	swarming_api "go.chromium.org/luci/common/api/swarming/swarming/v1"

	"infra/libs/skylab/request"
)

// stubSwarming implements skylab_api.Swarming.
type stubSwarming struct {
	botExists bool
}

func (f *stubSwarming) CreateTask(context.Context, *swarming_api.SwarmingRpcsNewTaskRequest) (*swarming_api.SwarmingRpcsTaskRequestMetadata, error) {
	return nil, nil
}

func (f *stubSwarming) GetResults(context.Context, []string) ([]*swarming_api.SwarmingRpcsTaskResult, error) {
	return nil, nil
}

func (f *stubSwarming) BotExists(context.Context, []*swarming_api.SwarmingRpcsStringPair) (bool, error) {
	return f.botExists, nil
}

func (f *stubSwarming) GetTaskURL(string) string {
	return ""
}

func (f *stubSwarming) setCannedBotExistsResponse(b bool) {
	f.botExists = b
}

func TestNonExistentBot(t *testing.T) {
	Convey("When arguments ask for a non-existent bot", t, func() {
		var swarming stubSwarming
		swarming.setCannedBotExistsResponse(false)
		skylab := &bbSkylabClient{
			swarmingClient: &swarming,
		}
		Convey("the validation fails.", func() {
			exists, err := skylab.ValidateArgs(context.Background(), &request.Args{})
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
		})
	})
}

func TestExistingBot(t *testing.T) {
	Convey("When arguments ask for an existing bot", t, func() {
		var swarming stubSwarming
		swarming.setCannedBotExistsResponse(true)
		skylab := &bbSkylabClient{
			swarmingClient: &swarming,
		}
		Convey("the validation passes.", func() {
			exists, err := skylab.ValidateArgs(context.Background(), &request.Args{})
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
		})
	})
}
