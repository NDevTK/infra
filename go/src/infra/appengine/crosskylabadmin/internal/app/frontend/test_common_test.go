// Copyright 2018 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
)

func TestBotForDUT(t *testing.T) {
	Convey("empty dimensions", t, func() {
		So(BotForDUT("dut1", "", ""), ShouldResemble, &swarming.SwarmingRpcsBotInfo{
			BotId: "bot_dut1",
			Dimensions: []*swarming.SwarmingRpcsStringListPair{
				{Key: "dut_state", Value: []string{""}},
				{Key: "dut_id", Value: []string{"dut1"}},
				{Key: "dut_name", Value: []string{"dut1-host"}},
			},
		})
	})

	Convey("non-trivial dimensions with whitespace", t, func() {
		So(BotForDUT("dut1", "fake_state", "a: x, y ; b :z"), ShouldResemble, &swarming.SwarmingRpcsBotInfo{
			BotId: "bot_dut1",
			Dimensions: []*swarming.SwarmingRpcsStringListPair{
				{Key: "a", Value: []string{"x", "y"}},
				{Key: "b", Value: []string{"z"}},
				{Key: "dut_state", Value: []string{"fake_state"}},
				{Key: "dut_id", Value: []string{"dut1"}},
				{Key: "dut_name", Value: []string{"dut1-host"}},
			},
		},
		)
	})
}
