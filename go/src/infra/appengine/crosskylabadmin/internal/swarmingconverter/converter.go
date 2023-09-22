// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package swarmingconverter

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	swarmingv1 "go.chromium.org/luci/common/api/swarming/swarming/v1"
	swarmingv2 "go.chromium.org/luci/swarming/proto/api_v2"
)

func stringToTimestampPb(a string) *timestamppb.Timestamp {
	t, err := time.Parse(time.RFC3339, a)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
}

// ConvertSwarmingRpcsStringListPair converts a string list pair.
func ConvertSwarmingRpcsStringListPair(p *swarmingv1.SwarmingRpcsStringListPair) *swarmingv2.StringListPair {
	if p == nil {
		return nil
	}

	return &swarmingv2.StringListPair{
		Key:   p.Key,
		Value: p.Value,
	}
}

// ConvertSwarmingRpcsStringListPair converts an array of string list pairs.
func ConvertSwarmingRpcsStringListPairs(p []*swarmingv1.SwarmingRpcsStringListPair) []*swarmingv2.StringListPair {
	out := make([]*swarmingv2.StringListPair, 0, len(p))
	for _, v := range p {
		out = append(out, ConvertSwarmingRpcsStringListPair(v))
	}
	return out
}

// ConvertSwarmingRpcsBotInfo converts a v1 bot info into a v2 bot info.
func ConvertSwarmingRpcsBotInfo(i *swarmingv1.SwarmingRpcsBotInfo) *swarmingv2.BotInfo {
	if i == nil {
		return nil
	}

	return &swarmingv2.BotInfo{
		BotId:           i.BotId,
		TaskId:          i.TaskId,
		ExternalIp:      i.ExternalIp,
		AuthenticatedAs: i.AuthenticatedAs,
		FirstSeenTs:     stringToTimestampPb(i.FirstSeenTs),
		IsDead:          i.IsDead,
		LastSeenTs:      stringToTimestampPb(i.LastSeenTs),
		Quarantined:     i.Quarantined,
		MaintenanceMsg:  i.MaintenanceMsg,
		Dimensions:      ConvertSwarmingRpcsStringListPairs(i.Dimensions),
		TaskName:        i.TaskName,
		Version:         i.Version,
		State:           i.State,
		Deleted:         i.Deleted,
	}
}

// ConvertSwarmingRpcsBotInfos converts an array of bot infos.
func ConvertSwarmingRpcsBotInfos(p []*swarmingv1.SwarmingRpcsBotInfo) []*swarmingv2.BotInfo {
	out := make([]*swarmingv2.BotInfo, 0, len(p))
	for _, v := range p {
		out = append(out, ConvertSwarmingRpcsBotInfo(v))
	}
	return out
}
