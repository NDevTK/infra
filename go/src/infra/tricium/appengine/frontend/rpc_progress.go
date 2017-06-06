// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"fmt"
	"strconv"
	"strings"

	ds "github.com/luci/gae/service/datastore"
	"github.com/luci/luci-go/common/logging"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"infra/tricium/api/v1"
	"infra/tricium/appengine/common/track"
)

// Progress implements Tricium.Progress.
func (r *TriciumServer) Progress(c context.Context, req *tricium.ProgressRequest) (*tricium.ProgressResponse, error) {
	if req.RunId == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "missing run ID")
	}
	runID, err := strconv.ParseInt(req.RunId, 10, 64)
	if err != nil {
		logging.WithError(err).Errorf(c, "failed to parse run ID: %s", req.RunId)
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid run ID")
	}
	runState, analyzerProgress, err := progress(c, runID)
	if err != nil {
		logging.WithError(err).Errorf(c, "progress failed: %v, run ID: %d", err, runID)
		return nil, grpc.Errorf(codes.Internal, "failed to execute progress request")
	}
	logging.Infof(c, "[frontend] Analyzer progress: %v", analyzerProgress)
	return &tricium.ProgressResponse{
		State:            runState,
		AnalyzerProgress: analyzerProgress,
	}, nil
}

func progress(c context.Context, runID int64) (tricium.State, []*tricium.AnalyzerProgress, error) {
	runKey := ds.NewKey(c, "WorkflowRun", "", runID, nil)
	runRes := &track.WorkflowRunResult{ID: "1", Parent: runKey}
	if err := ds.Get(c, runRes); err != nil {
		return tricium.State_PENDING, nil, fmt.Errorf("failed to get WorkflowRunResult: %v", err)
	}
	var workers []*track.WorkerRun
	q := ds.NewQuery("WorkerRun").Ancestor(runKey)
	if err := ds.GetAll(c, q, &workers); err != nil {
		return tricium.State_PENDING, nil, fmt.Errorf("failed to get WorkerRun: %v", err)
	}
	res := []*tricium.AnalyzerProgress{}
	for _, w := range workers {
		workerKey := ds.KeyForObj(c, w)
		wr := &track.WorkerRunResult{ID: "1", Parent: workerKey}
		if err := ds.Get(c, wr); err != nil {
			return tricium.State_PENDING, nil, fmt.Errorf("failed to get WorkerResult: %v", err)
		}
		res = append(res, &tricium.AnalyzerProgress{
			Analyzer:       extractAnalyzerName(w.ID),
			Platform:       w.Platform,
			State:          wr.State,
			SwarmingTaskId: fmt.Sprintf("%s/task?id=%s", w.SwarmingServerURL, wr.SwarmingTaskID),
			NumComments:    int32(wr.NumComments),
		})
	}
	return runRes.State, res, nil
}

func extractAnalyzerName(worker string) string {
	parts := strings.SplitN(worker, "_", 2)
	if len(parts) == 0 {
		return worker
	}
	return parts[0]
}
