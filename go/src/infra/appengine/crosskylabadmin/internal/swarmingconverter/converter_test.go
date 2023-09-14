// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package swarmingconverter

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	swarmingv1 "go.chromium.org/luci/common/api/swarming/swarming/v1"
	swarmingv2 "go.chromium.org/luci/swarming/proto/api_v2"
)

func TestConvertSwarmingRpcsBotInfo(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   *swarmingv1.SwarmingRpcsBotInfo
		out  *swarmingv2.BotInfo
	}{
		{
			name: "empty",
			in:   nil,
			out:  nil,
		},
		{
			name: "non-empty",
			in: &swarmingv1.SwarmingRpcsBotInfo{
				BotId:           "a",
				TaskId:          "b",
				ExternalIp:      "c",
				AuthenticatedAs: "d",
				FirstSeenTs:     "",
				IsDead:          true,
				LastSeenTs:      "",
				Quarantined:     true,
				MaintenanceMsg:  "e",
				Dimensions:      nil,
				TaskName:        "f",
				Version:         "g",
				State:           "e",
				Deleted:         true,
			},
			out: &swarmingv2.BotInfo{
				BotId:           "a",
				TaskId:          "b",
				ExternalIp:      "c",
				AuthenticatedAs: "d",
				FirstSeenTs:     nil,
				IsDead:          true,
				LastSeenTs:      nil,
				Quarantined:     true,
				MaintenanceMsg:  "e",
				Dimensions:      nil,
				TaskName:        "f",
				Version:         "g",
				State:           "e",
				Deleted:         true,
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expected := tt.out
			actual := ConvertSwarmingRpcsBotInfo(tt.in)
			if diff := cmp.Diff(expected, actual, protocmp.Transform()); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}
