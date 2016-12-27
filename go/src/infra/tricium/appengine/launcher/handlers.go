// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package launcher implements HTTP handlers for the launcher module.
package launcher

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/golang/protobuf/jsonpb"
	"github.com/google/go-querystring/query"
	ds "github.com/luci/gae/service/datastore"
	tq "github.com/luci/gae/service/taskqueue"
	"github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/server/router"

	admin "infra/tricium/api/admin/v1"
	"infra/tricium/api/v1"
	"infra/tricium/appengine/common"
	"infra/tricium/appengine/common/pipeline"
)

func queueHandler(ctx *router.Context) {
	c, r, w := ctx.Context, ctx.Request, ctx.Writer
	if err := r.ParseForm(); err != nil {
		logging.WithError(err).Errorf(c, "Launch queue handler encountered errors")
		w.WriteHeader(500)
		return
	}
	lr, err := pipeline.ParseLaunchRequest(r.Form)
	if err != nil {
		logging.WithError(err).Errorf(c, "Launch queue handler encountered errors")
		w.WriteHeader(501)
		return
	}
	logging.Infof(c, "[launcher] Launch request (run ID: %d)", lr.RunID)
	if err := launch(c, lr); err != nil {
		logging.WithError(err).Errorf(c, "Launch queue handler encountered errors")
		w.WriteHeader(502)
		return
	}
	logging.Infof(c, "[launcher] Successfully completed")
}

func launch(c context.Context, lr *pipeline.LaunchRequest) error {
	// Store workflow as 'Workflow' entity, using run ID as key.
	wf, err := readWorkflowConfig(lr.Project)
	if err != nil {
		return err
	}
	m := jsonpb.Marshaler{}
	wfs, err := m.MarshalToString(wf)
	if err != nil {
		return fmt.Errorf("Failed to marshal workflow: %v", err)
	}
	workflowKey := ds.NewKey(c, "Workflow", "", lr.RunID, nil)
	e := new(common.Entity)
	e.Value = []byte(wfs)
	if err := ds.Put(c, workflowKey, e); err != nil {
		return fmt.Errorf("Failed to store workflow: %v", err)
	}

	// Isolate initial intput.
	inputHash, err := isolateGitFileDetails(lr.Project, lr.GitRepo, lr.GitRef, lr.Path)
	if err != nil {
		return fmt.Errorf("Failed to isolate git file details: %v", err)
	}

	// Track progress, enqueue track request.
	vr, err := query.Values(&pipeline.TrackRequest{
		Kind:  pipeline.TrackWorkflowLaunched,
		RunID: lr.RunID,
	})
	if err != nil {
		return fmt.Errorf("Failed to encode reporter request: %v", err)
	}
	tr := tq.NewPOSTTask("/tracker/internal/queue", vr)
	if err := tq.Add(c, "tracker-queue", tr); err != nil {
		return fmt.Errorf("Failed to enqueue reporter request: %v", err)
	}

	// Trigger root workers, enqueue driver requests.
	for _, worker := range rootWorkers(wf) {
		vd, err := query.Values(&pipeline.DriverRequest{
			Kind:          pipeline.DriverTrigger,
			RunID:         lr.RunID,
			IsolatedInput: inputHash,
			Worker:        worker,
		})
		if err != nil {
			return fmt.Errorf("Failed to encode launch request: %v", err)
		}
		td := tq.NewPOSTTask("/driver/internal/queue", vd)
		if err := tq.Add(c, "driver-queue", td); err != nil {
			return fmt.Errorf("Failed to enqueue driver request: %v", err)
		}
	}
	return nil
}

// rootWorkers returns a list of root workers.
//
// Root workers are those workers in need of the initial Tricium
// data type, Git file details.
func rootWorkers(wf *admin.Workflow) []string {
	wl := []string{}
	for _, w := range wf.Workers {
		if w.Needs == tricium.Data_GIT_FILE_DETAILS {
			wl = append(wl, w.Name)
		}
	}
	return wl
}

func isolateGitFileDetails(project, gitRepo, gitRef string, paths []string) (string, error) {
	// TODO(emso): Create initial Tricium data, git file details.
	// TODO(emso): Isolate created Tricium data.
	return "abcedfg", nil
}

func readWorkflowConfig(project string) (*admin.Workflow, error) {
	// TODO(emso): Replace this dummy config with one read from luci-config.
	return &admin.Workflow{
		WorkerTopic:    "projects/tricium-dev/topics/worker-completion",
		ServiceAccount: "emso@chromium.org",
		Workers: []*admin.Worker{
			{
				Name:     "Hello_Ubuntu14.04_x86-64",
				Needs:    tricium.Data_GIT_FILE_DETAILS,
				Provides: tricium.Data_FILES,
				Platform: "Ubuntu14.04_x86-64",
				Dimensions: []string{
					"pool:Chrome",
					"os:Ubuntu-14.04",
					"cpu:x84-64",
				},
				Cmd: &tricium.Cmd{
					Exec: "echo",
					Args: []string{
						"'hello'",
					},
				},
				Deadline: 30,
			},
		},
	}, nil
}
