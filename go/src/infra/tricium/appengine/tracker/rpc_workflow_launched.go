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
	"infra/tricium/appengine/common/config"
	"infra/tricium/appengine/common/track"
)

// TrackerServer represents the Tricium pRPC Tracker server.
type trackerServer struct{}

var server = &trackerServer{}

// WorkflowLaunched tracks the launch of a workflow.
func (*trackerServer) WorkflowLaunched(c context.Context, req *admin.WorkflowLaunchedRequest) (*admin.WorkflowLaunchedResponse, error) {
	if req.RunId == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "missing run ID")
	}
	if err := workflowLaunched(c, req, config.WorkflowCache); err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to track workflow launched: %v", err)
	}
	return &admin.WorkflowLaunchedResponse{}, nil
}

func workflowLaunched(c context.Context, req *admin.WorkflowLaunchedRequest, wp config.WorkflowCacheAPI) error {
	wf, err := wp.GetWorkflow(c, req.RunId)
	if err != nil {
		return fmt.Errorf("failed to read workflow config: %v", err)
	}
	// Prepare analyzer and worker invocation tracking entries to store.
	aw := extractAnalyzerWorkerStructure(c, wf)
	logging.Infof(c, "Extracted analyzer/worker entries for tracking: %#v", aw)
	return ds.RunInTransaction(c, func(c context.Context) (err error) {
		run := &track.Run{ID: req.RunId}
		if err := ds.Get(c, run); err != nil {
			return fmt.Errorf("failed to retrieve run entry (run ID: %d): %v", run.ID, err)
		}
		runKey := ds.KeyForObj(c, run)
		ops := []func() error{
			// Notify reporter.
			func() error {
				switch run.Reporter {
				case tricium.Reporter_GERRIT:
					// TOOD(emso): push notification to the Gerrit reporter
				default:
					// Do nothing.
				}
				return nil
			},
			// Update Run state to launched by setting RunResult state to running.
			func() error {
				r := &track.RunResult{
					ID:     "1",
					Parent: runKey,
					State:  tricium.State_RUNNING,
				}
				if err := ds.Put(c, r); err != nil {
					return fmt.Errorf("failed to mark workflow as launched: %v", err)
				}
				return nil
			},
			// Store analyzer and worker invocation entries for tracking.
			func() error {
				entities := make([]interface{}, 0, len(aw))
				for _, v := range aw {
					v.Analyzer.Parent = runKey
					analyzerKey := ds.KeyForObj(c, v.Analyzer)
					entities = append(entities, []interface{}{
						v.Analyzer,
						&track.AnalyzerResult{
							ID:     "1",
							Parent: analyzerKey,
							Name:   v.Analyzer.ID,
							State:  tricium.State_RUNNING,
						},
					}...)
					for _, worker := range v.Workers {
						worker.Parent = analyzerKey
						entities = append(entities, worker)
						workerKey := ds.KeyForObj(c, worker)
						entities = append(entities, []interface{}{
							worker,
							&track.WorkerResult{
								ID:     "1",
								Name:   worker.ID,
								Parent: workerKey,
								State:  tricium.State_RUNNING,
							},
						}...)
					}
				}
				if err := ds.Put(c, entities); err != nil {
					return fmt.Errorf("failed to store analyzer and worker entries: %v", err)
				}
				return nil
			},
		}
		return common.RunInParallel(ops)
	}, nil)
}

type analyzerToWorkers struct {
	Analyzer *track.AnalyzerRun
	Workers  []*track.WorkerRun
}

// extractAnalyzerWorkerStructure extracts analyzer-*worker structure from workflow config.
func extractAnalyzerWorkerStructure(c context.Context, wf *admin.Workflow) map[string]*analyzerToWorkers {
	m := map[string]*analyzerToWorkers{}
	for _, w := range wf.Workers {
		analyzer := strings.Split(w.Name, "_")[0]
		a, ok := m[analyzer]
		if !ok {
			a = &analyzerToWorkers{
				Analyzer: &track.AnalyzerRun{
					ID: analyzer,
				},
			}
			m[analyzer] = a
		}
		aw := &track.WorkerRun{
			ID:                w.Name,
			IsolateServerURL:  wf.IsolateServer,
			SwarmingServerURL: wf.SwarmingServer,
			Platform:          w.ProvidesForPlatform,
		}
		for _, n := range w.Next {
			aw.Next = append(aw.Next, n)
		}
		a.Workers = append(a.Workers, aw)
		logging.Debugf(c, "Found analyzer/worker: %v", a)
	}
	return m
}
