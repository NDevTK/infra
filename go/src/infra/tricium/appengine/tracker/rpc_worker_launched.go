// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tracker

import (
	"fmt"
	"strings"

	ds "github.com/luci/gae/service/datastore"
	"github.com/luci/luci-go/common/logging"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"infra/tricium/api/admin/v1"
	"infra/tricium/api/v1"
	"infra/tricium/appengine/common"
	"infra/tricium/appengine/common/track"
)

// WorkerLaunched tracks the launch of a worker.
func (*trackerServer) WorkerLaunched(c context.Context, req *admin.WorkerLaunchedRequest) (*admin.WorkerLaunchedResponse, error) {
	if req.RunId == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "missing run ID")
	}
	if req.Worker == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "missing worker")
	}
	if req.IsolatedInputHash == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "missing isolated input hash")
	}
	if req.SwarmingTaskId == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "missing swarming task ID")
	}
	if err := workerLaunched(c, req); err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to track worker launched: %v", err)
	}
	return &admin.WorkerLaunchedResponse{}, nil
}

func workerLaunched(c context.Context, req *admin.WorkerLaunchedRequest) error {
	logging.Debugf(c, "[tracker] Worker launched (run ID: %d, worker: %s, taskID: %s, IsolatedInput: %s)", req.RunId, req.Worker, req.SwarmingTaskId, req.IsolatedInputHash)
	// Compute needed keys.
	runKey := ds.NewKey(c, "Run", "", req.RunId, nil)
	analyzerName := strings.Split(req.Worker, "_")[0]
	analyzerKey := ds.NewKey(c, "AnalyzerRun", analyzerName, 0, runKey)
	workerKey := ds.NewKey(c, "WorkerRun", req.Worker, 0, analyzerKey)
	return ds.RunInTransaction(c, func(c context.Context) (err error) {
		ops := []func() error{
			// Notify reporter.
			func() error {
				run := &track.Run{ID: req.RunId}
				if err := ds.Get(c, run); err != nil {
					return fmt.Errorf("failed to get Run entry (run ID: %d): %v", run.ID, err)
				}
				switch run.Reporter {
				case tricium.Reporter_GERRIT:
					// TOOD(emso): push notification to the Gerrit reporter
				default:
					// Do nothing.
				}
				return nil
			},
			// Update worker state to launched.
			func() error {
				wr := &track.WorkerResult{
					ID:             "1",
					Parent:         workerKey,
					State:          tricium.State_RUNNING,
					IsolatedInput:  req.IsolatedInputHash,
					SwarmingTaskID: req.SwarmingTaskId,
				}
				if err := ds.Put(c, wr); err != nil {
					return fmt.Errorf("failed to update WorkerResult: %v", err)
				}
				return nil
			},
			// Maybe update analyzer state to launched.
			func() error {
				ar := &track.AnalyzerResult{
					ID:     "1",
					Parent: analyzerKey,
				}
				if err := ds.Get(c, ar); err != nil {
					return fmt.Errorf("failed to get AnalyzerResult: %v", err)
				}
				if ar.State == tricium.State_PENDING {
					ar.State = tricium.State_RUNNING
					if err := ds.Put(c, ar); err != nil {
						return fmt.Errorf("failed to update AnalyzerResult to launched: %v", err)
					}
				}
				return nil
			},
		}
		return common.RunInParallel(ops)
	}, nil)
}
