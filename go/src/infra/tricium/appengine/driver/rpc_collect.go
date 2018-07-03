// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package driver

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	tq "go.chromium.org/gae/service/taskqueue"
	"go.chromium.org/luci/common/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"infra/tricium/api/admin/v1"
	"infra/tricium/api/v1"
	"infra/tricium/appengine/common"
	"infra/tricium/appengine/common/config"
)

// Collect processes one collect request to the Tricium driver.
func (*driverServer) Collect(c context.Context, req *admin.CollectRequest) (*admin.CollectResponse, error) {
	logging.Infof(c, "[driver]: Received collect request (run ID: %d, worker: %s, task ID: %s)", req.RunId, req.Worker, req.TaskId)
	if err := validateCollectRequest(req); err != nil {
		return nil, err
	}
	if err := collect(c, req, config.WorkflowCache, common.SwarmingServer, common.IsolateServer); err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to collect: %v", err)
	}
	return &admin.CollectResponse{}, nil
}

func validateCollectRequest(req *admin.CollectRequest) error {
	if req.RunId == 0 {
		return grpc.Errorf(codes.InvalidArgument, "missing run ID")
	}
	if req.Worker == "" {
		return grpc.Errorf(codes.InvalidArgument, "missing worker name")
	}
	return nil
}

func collect(c context.Context, req *admin.CollectRequest, wp config.WorkflowCacheAPI, sw common.SwarmingAPI, isolator common.IsolateAPI) error {
	wf, err := wp.GetWorkflow(c, req.RunId)
	if err != nil {
		return fmt.Errorf("failed to read workflow config: %v", err)
	}
	isolatedOutput, exitCode, err := sw.Collect(c, wf.SwarmingServer, req.TaskId)
	if err != nil {
		return fmt.Errorf("failed to collect swarming task result: %v", err)
	}
	if isolatedOutput == "" {
		// No isolated output was found. The task may not be done yet.
		// Try to re-enqueue.
		if err = enqueueCollectRequest(c, req); err != nil {
			return err
		}
		return nil
	}
	w, err := wf.GetWorker(req.Worker)
	if err != nil {
		return fmt.Errorf("failed to get worker output type: %v", err)
	}

	// Worker state.
	workerState := tricium.State_SUCCESS
	if exitCode != 0 {
		workerState = tricium.State_FAILURE
	}

	// Mark worker as done.
	b, err := proto.Marshal(&admin.WorkerDoneRequest{
		RunId:              req.RunId,
		Worker:             req.Worker,
		IsolatedOutputHash: isolatedOutput,
		Provides:           w.Provides,
		State:              workerState,
	})
	if err != nil {
		return fmt.Errorf("failed to encode worker done request: %v", err)
	}
	t := tq.NewPOSTTask("/tracker/internal/worker-done", nil)
	t.Payload = b
	if err := tq.Add(c, common.TrackerQueue, t); err != nil {
		return fmt.Errorf("failed to enqueue track request: %v", err)
	}

	// Abort here if worker failed and mark descendants as failures.
	if workerState == tricium.State_FAILURE {
		logging.Warningf(c, "Execution of worker failed, exit code: %d, worker: %s, run ID: %s", exitCode, req.Worker, req.RunId)
		var tasks []*tq.Task
		for _, worker := range wf.GetWithDescendants(req.Worker) {
			if worker == req.Worker {
				continue
			}
			// Mark descendant worker as done and failed.
			b, err := proto.Marshal(&admin.WorkerDoneRequest{
				RunId:  req.RunId,
				Worker: worker,
				State:  tricium.State_ABORTED,
			})
			if err != nil {
				return fmt.Errorf("failed to encode worker done request: %v", err)
			}
			t := tq.NewPOSTTask("/tracker/internal/worker-done", nil)
			t.Payload = b
			tasks = append(tasks, t)
		}
		if err := tq.Add(c, common.TrackerQueue, tasks...); err != nil {
			return fmt.Errorf("failed to enqueue track request: %v", err)
		}
		return nil
	}

	// Create layered isolated input, include the input in the collect request and
	// massage the isolated output into new isolated input.
	isolatedInput, err := isolator.LayerIsolates(c, wf.IsolateServer, req.IsolatedInputHash, isolatedOutput)
	if err != nil {
		return fmt.Errorf("failed layer isolates: %v", err)
	}

	// Enqueue trigger requests for successors.
	for _, worker := range wf.GetNext(req.Worker) {
		b, err := proto.Marshal(&admin.TriggerRequest{
			RunId:             req.RunId,
			IsolatedInputHash: isolatedInput,
			Worker:            worker,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal successor trigger request: %v", err)
		}
		t := tq.NewPOSTTask("/driver/internal/trigger", nil)
		t.Payload = b
		if err := tq.Add(c, common.DriverQueue, t); err != nil {
			return fmt.Errorf("failed to enqueue collect request: %v", err)
		}
	}
	return nil
}
