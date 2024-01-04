// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package rpc_services

import (
	"context"
	"errors"
	"strings"

	"infra/cros/satlab/common/run"

	pb "go.chromium.org/chromiumos/infra/proto/go/satlabrpcserver"
)

// BUG_ID is a key for the run suite tag.
const BUG_ID = "bug_id"

// RunStorageQual run a storage qualification suite
func (s *SatlabRpcServiceServer) RunStorageQual(
	ctx context.Context,
	in *pb.RunStorageQualRequest,
) (*pb.RunStorageQualResponse, error) {
	if err := s.validateServices(); err != nil {
		return nil, err
	}

	bugId := strings.TrimSpace(in.GetBugId())

	// if bug id is empty, return an error.
	if bugId == "" {
		return nil, errors.New("bug id is empty")
	}

	r := &run.Run{
		Suite:     in.GetSuite(),
		Model:     in.GetModel(),
		Board:     in.GetBoard(),
		Milestone: in.GetMilestone(),
		Build:     in.GetBuild(),
		Pool:      in.GetPool(),
		AddedDims: parseDims(in.GetDims()),
		Tags: map[string]string{
			BUG_ID: bugId,
		},
		TRV2: true,
	}
	buildLink, err := r.TriggerRun(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.RunStorageQualResponse{BuildLink: buildLink}, nil
}
