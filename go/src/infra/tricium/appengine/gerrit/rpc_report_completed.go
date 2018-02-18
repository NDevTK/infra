// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gerrit

import (
	"bytes"
	"fmt"

	ds "go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/common/logging"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"infra/tricium/api/admin/v1"
	"infra/tricium/appengine/common"
	"infra/tricium/appengine/common/track"
)

// ReportCompleted implements the admin.Reporter interface.
func (r *gerritReporter) ReportCompleted(c context.Context, req *admin.ReportCompletedRequest) (*admin.ReportCompletedResponse, error) {
	logging.Debugf(c, "[gerrit-reporter] ReportCompleted request (run ID: %d)", req.RunId)
	if req.RunId == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "missing run ID")
	}
	if err := reportCompleted(c, req, GerritServer); err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to report completed to Gerrit: %v", err)
	}
	return &admin.ReportCompletedResponse{}, nil
}

func reportCompleted(c context.Context, req *admin.ReportCompletedRequest, gerrit API) error {
	request := &track.AnalyzeRequest{ID: req.RunId}
	var functionResults []*track.FunctionRunResult
	ops := []func() error{
		// Get Git details.
		func() error {
			// The Git repo and ref in the service request should correspond to the Gerrit
			// repo for the project. This request is typically done by the Gerrit poller.
			if err := ds.Get(c, request); err != nil {
				return fmt.Errorf("failed to get AnalyzeRequest entity (ID: %d): %v", req.RunId, err)
			}
			return nil
		},
		// Get function results.
		func() error {
			requestKey := ds.NewKey(c, "AnalyzeRequest", "", req.RunId, nil)
			runKey := ds.NewKey(c, "WorkflowRun", "", 1, requestKey)
			if err := ds.GetAll(c, ds.NewQuery("FunctionRunResult").Ancestor(runKey), &functionResults); err != nil {
				return fmt.Errorf("failed to get FunctionRunResult entities: %v", err)
			}
			return nil
		},
	}
	if err := common.RunInParallel(ops); err != nil {
		return err
	}
	if request.GerritReportingDisabled {
		logging.Infof(c, "Gerrit reporting disabled, not reporting completion (run ID: %s, project: %s)", req.RunId, request.Project)
		return nil
	}
	// Create result message.
	n := 0
	var buf bytes.Buffer
	for _, fr := range functionResults {
		n += fr.NumComments
		buf.WriteString(fmt.Sprintf("  %s: %d\n", fr.Name, fr.NumComments))
	}
	results := fmt.Sprintf("%d result", n)
	if n > 1 {
		results = fmt.Sprintf("%d results", n)
	}
	msg := fmt.Sprintf("Tricium finished analyzing patch set and found %s (run ID: %d).\n%s", results, req.RunId, buf.String())
	return gerrit.PostReviewMessage(c, request.GerritHost, request.GerritChange, request.GerritRevision, msg)
}
