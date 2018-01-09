// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"fmt"
	"strconv"

	ds "go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/common/logging"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"infra/tricium/api/v1"
	"infra/tricium/appengine/common/config"
	"infra/tricium/appengine/common/track"
)

// Progress implements Tricium.Progress.
func (r *TriciumServer) Progress(c context.Context, req *tricium.ProgressRequest) (*tricium.ProgressResponse, error) {
	project, runID, err := validateProgressRequest(c, req)
	if err != nil {
		return nil, err
	}
	// Run progress for a project.
	if project != "" {
		runProgress, errCode, err := projectProgress(c, project, config.LuciConfigServer)
		if err != nil {
			logging.WithError(err).Errorf(c, "project progress failed: %v, project: %s", err, project)
			return nil, grpc.Errorf(errCode, "failed to execute progress request")
		}
		logging.Infof(c, "[frontend] Project progress: %v", runProgress)
		return &tricium.ProgressResponse{
			RunProgress: runProgress,
		}, nil
	}
	// Analyzer progress for a run.
	runState, analyzerProgress, errCode, err := progress(c, runID)
	if err != nil {
		logging.WithError(err).Errorf(c, "progress failed: %v, run ID: %d", err, runID)
		return nil, grpc.Errorf(errCode, "failed to execute progress request")
	}
	logging.Infof(c, "[frontend] Analyzer progress: %v", analyzerProgress)
	return &tricium.ProgressResponse{
		RunId:            strconv.FormatInt(runID, 10),
		State:            runState,
		AnalyzerProgress: analyzerProgress,
	}, nil
}

func validateProgressRequest(c context.Context, req *tricium.ProgressRequest) (string, int64, error) {
	if req.Project != "" {
		return req.Project, 0, nil
	}
	if req.Consumer == tricium.Consumer_GERRIT {
		// Either Gerrit details or run ID should be given; if both are
		// given then they may be conflicting; if the run ID is given
		// then there should be no need to specify Gerrit details.
		if req.RunId != "" {
			return "", 0, grpc.Errorf(codes.InvalidArgument, "both Gerrit details and run ID given")
		}
		gd := req.GetGerritDetails()
		if gd == nil {
			return "", 0, grpc.Errorf(codes.InvalidArgument, "missing Gerrit details")
		}
		if gd.Host == "" {
			return "", 0, grpc.Errorf(codes.InvalidArgument, "missing Gerrit host")
		}
		if gd.Project == "" {
			return "", 0, grpc.Errorf(codes.InvalidArgument, "missing Gerrit project")
		}
		if gd.Change == "" {
			return "", 0, grpc.Errorf(codes.InvalidArgument, "missing Gerrit change ID")
		}
		if gd.Revision == "" {
			return "", 0, grpc.Errorf(codes.InvalidArgument, "missing Gerrit revision")
		}
		// Look up the run ID with the provided Gerrit change details.
		g := &GerritChangeToRunID{
			ID: gerritMappingID(gd.Host, gd.Project, gd.Change),
		}
		if err := ds.Get(c, g); err != nil {
			logging.WithError(err).Errorf(c, "failed to get GerritChangeToRunID entity: %v", err)
			return "", 0, grpc.Errorf(codes.InvalidArgument, "failed to find run ID for Gerrit change")
		}
		return "", g.RunID, nil
	}
	if req.RunId == "" {
		return "", 0, grpc.Errorf(codes.InvalidArgument, "missing run ID")
	}
	runID, err := strconv.ParseInt(req.RunId, 10, 64)
	if err != nil {
		logging.WithError(err).Errorf(c, "failed to parse run ID: %s", req.RunId)
		return "", 0, grpc.Errorf(codes.InvalidArgument, "invalid run ID")
	}
	return "", runID, nil
}

func projectProgress(c context.Context, project string, cp config.ProviderAPI) ([]*tricium.RunProgress, codes.Code, error) {
	sc, err := cp.GetServiceConfig(c)
	if err != nil {
		return nil, codes.Internal, fmt.Errorf("failed to get service config: %v", err)
	}
	pd := tricium.LookupProjectDetails(sc, project)
	if pd == nil {
		return nil, codes.InvalidArgument, fmt.Errorf("unknown project")
	}
	var runProgress []*tricium.RunProgress
	var requests []*track.AnalyzeRequest
	if err := ds.GetAll(c, ds.NewQuery("AnalyzeRequest").Eq("Project", project), &requests); err != nil {
		return nil, codes.Internal, fmt.Errorf("failed to retrieve AnalyzeRequest entities: %v", err)
	}
	// TODO(emso): gather info from found analyze requests
	// TODO(emso): sort result on state
	return runProgress, codes.OK, nil
}

func progress(c context.Context, runID int64) (tricium.State, []*tricium.AnalyzerProgress, codes.Code, error) {
	requestKey := ds.NewKey(c, "AnalyzeRequest", "", runID, nil)
	requestRes := &track.AnalyzeRequestResult{ID: 1, Parent: requestKey}
	if err := ds.Get(c, requestRes); err != nil {
		return tricium.State_PENDING, nil, codes.InvalidArgument, fmt.Errorf("failed to get AnalyzeRequestResult: %v", err)
	}
	run := &track.WorkflowRun{ID: 1, Parent: requestKey}
	if err := ds.Get(c, run); err != nil {
		return tricium.State_PENDING, nil, codes.Internal, fmt.Errorf("failed to get AnalyzeRequestResult: %v", err)
	}
	runKey := ds.KeyForObj(c, run)
	// TODO(emso): extract a common GetAnalyzerRunsForWorkflowRun function
	var analyzers []*track.AnalyzerRun
	for _, analyzerName := range run.Analyzers {
		analyzers = append(analyzers, &track.AnalyzerRun{ID: analyzerName, Parent: runKey})
	}
	logging.Debugf(c, "Reading results for analyzers: %v, run: %v", analyzers, run)
	if err := ds.Get(c, analyzers); err != nil {
		return tricium.State_PENDING, nil, codes.Internal, fmt.Errorf("failed to get AnalyzerRun entities: %v", err)
	}
	var workerResults []*track.WorkerRunResult
	for _, analyzer := range analyzers {
		analyzerKey := ds.KeyForObj(c, analyzer)
		for _, workerName := range analyzer.Workers {
			workerKey := ds.NewKey(c, "WorkerRun", workerName, 0, analyzerKey)
			workerResults = append(workerResults, &track.WorkerRunResult{ID: 1, Parent: workerKey})
		}
	}
	logging.Debugf(c, "Reading worker results for %v", workerResults)
	if err := ds.Get(c, workerResults); err != nil && err != ds.ErrNoSuchEntity {
		return tricium.State_PENDING, nil, codes.Internal, fmt.Errorf("failed to get WorkerRunResult entities: %v", err)
	}
	res := []*tricium.AnalyzerProgress{}
	for _, wr := range workerResults {
		p := &tricium.AnalyzerProgress{
			Analyzer:    wr.Analyzer,
			Platform:    wr.Platform,
			State:       wr.State,
			NumComments: int32(wr.NumComments),
		}
		if len(wr.SwarmingTaskID) > 0 {
			p.SwarmingUrl = run.SwarmingServerURL
			p.SwarmingTaskId = wr.SwarmingTaskID
		}
		res = append(res, p)
	}
	// Monitor progress requests per project and run ID.
	request := &track.AnalyzeRequest{ID: runID}
	if err := ds.Get(c, request); err != nil {
		return requestRes.State, res, codes.Internal, fmt.Errorf("failed to get AnalyzeRequest: %v", err)
	}
	progressRequestCount.Add(c, 1, request.Project, strconv.FormatInt(runID, 10))
	return requestRes.State, res, codes.OK, nil
}
