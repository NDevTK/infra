// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// package server implements the server to handle pRPC requests.
package server

import (
	"context"
	"infra/appengine/gofindit/compilefailureanalysis/heuristic"
	gfim "infra/appengine/gofindit/model"
	gfipb "infra/appengine/gofindit/proto"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GoFinditServer implements the proto service GoFinditService.
type GoFinditServer struct{}

// GetAnalysis returns the analysis given the analysis id
func (server *GoFinditServer) GetAnalysis(c context.Context, req *gfipb.GetAnalysisRequest) (*gfipb.Analysis, error) {
	analysis := &gfim.CompileFailureAnalysis{
		Id: req.AnalysisId,
	}
	switch err := datastore.Get(c, analysis); err {
	case nil:
		//continue
	case datastore.ErrNoSuchEntity:
		return nil, status.Errorf(codes.NotFound, "Analysis %d not found: %v", req.AnalysisId, err)
	default:
		return nil, status.Errorf(codes.Internal, "Error in retrieving analysis: %s", err)
	}
	result, err := GetAnalysisResult(c, analysis)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error getting analysis result: %s", err)
	}
	return result, nil
}

// QueryAnalysis returns the analysis given a query
func (server *GoFinditServer) QueryAnalysis(c context.Context, req *gfipb.QueryAnalysisRequest) (*gfipb.QueryAnalysisResponse, error) {
	if err := validateQueryAnalysisRequest(req); err != nil {
		return nil, err
	}
	if req.BuildFailure.FailedStepName != "compile" {
		return nil, status.Errorf(codes.Unimplemented, "only compile failures are supported")
	}
	bbid := req.BuildFailure.GetBbid()
	logging.Infof(c, "QueryAnalysis for build %d", bbid)

	analysis, err := GetAnalysisForBuild(c, bbid)
	if err != nil {
		logging.Errorf(c, "Could not query analysis for build %d: %s", bbid, err)
		return nil, status.Errorf(codes.Internal, "failed to get analysis for build %d: %s", bbid, err)
	}
	if analysis == nil {
		logging.Infof(c, "No analysis for build %d", bbid)
		return nil, status.Errorf(codes.NotFound, "analysis not found for build %d", bbid)
	}
	analysispb, err := GetAnalysisResult(c, analysis)
	if err != nil {
		logging.Errorf(c, "Could not get analysis data for build %d: %s", bbid, err)
		return nil, status.Errorf(codes.Internal, "failed to get analysis data %s", err)
	}

	res := &gfipb.QueryAnalysisResponse{
		Analyses: []*gfipb.Analysis{analysispb},
	}
	return res, nil
}

// TriggerAnalysis triggers an analysis for a failure
func (server *GoFinditServer) TriggerAnalysis(c context.Context, req *gfipb.TriggerAnalysisRequest) (*gfipb.TriggerAnalysisResponse, error) {
	// TODO(nqmtuan): Implement this
	return nil, nil
}

// UpdateAnalysis updates the information of an analysis.
// At the mean time, it is only used for update the bugs associated with an
// analysis.
func (server *GoFinditServer) UpdateAnalysis(c context.Context, req *gfipb.UpdateAnalysisRequest) (*gfipb.Analysis, error) {
	// TODO(nqmtuan): Implement this
	return nil, nil
}

// GetAnalysisResult returns an analysis for pRPC from CompileFailureAnalysis
func GetAnalysisResult(c context.Context, analysis *gfim.CompileFailureAnalysis) (*gfipb.Analysis, error) {
	result := &gfipb.Analysis{
		AnalysisId:      analysis.Id,
		Status:          analysis.Status,
		CreatedTime:     timestamppb.New(analysis.CreateTime),
		EndTime:         timestamppb.New(analysis.EndTime),
		FirstFailedBbid: analysis.FirstFailedBuildId,
		LastPassedBbid:  analysis.LastPassedBuildId,
	}

	// Check whether the analysis has an associated first failed build
	if analysis.FirstFailedBuildId != 0 {
		// Add details from first failed build
		firstFailedBuild, err := GetBuild(c, analysis.FirstFailedBuildId)
		if err != nil {
			return nil, err
		}
		if firstFailedBuild != nil {
			result.Builder = &buildbucketpb.BuilderID{
				Project: firstFailedBuild.Project,
				Bucket:  firstFailedBuild.Bucket,
				Builder: firstFailedBuild.Builder,
			}
			result.BuildFailureType = firstFailedBuild.BuildFailureType
		}
	}

	heuristicAnalysis, err := GetHeuristicAnalysis(c, analysis)
	if err != nil {
		return nil, err
	}
	if heuristicAnalysis == nil {
		// No heuristic analysis associated with the compile failure analysis
		return result, nil
	}

	suspects, err := GetSuspects(c, heuristicAnalysis)
	if err != nil {
		return nil, err
	}

	pbSuspects := make([]*gfipb.HeuristicSuspect, len(suspects))
	for i, suspect := range suspects {
		pbSuspects[i] = &gfipb.HeuristicSuspect{
			GitilesCommit:   &suspect.GitilesCommit,
			ReviewUrl:       suspect.ReviewUrl,
			Score:           int32(suspect.Score),
			Justification:   suspect.Justification,
			ConfidenceLevel: heuristic.GetConfidenceLevel(suspect.Score),
		}
	}
	heuristicResult := &gfipb.HeuristicAnalysisResult{
		Status:    heuristicAnalysis.Status,
		StartTime: timestamppb.New(heuristicAnalysis.StartTime),
		EndTime:   timestamppb.New(heuristicAnalysis.EndTime),
		Suspects:  pbSuspects,
	}

	result.HeuristicResult = heuristicResult

	// TODO (nqmtuan): query for nth-section result

	// TODO (aredulla): get culprit actions, such as the revert CL for the culprit
	//                  and any related bugs

	return result, nil
}

// validateQueryAnalysisRequest checks if the request is valid.
func validateQueryAnalysisRequest(req *gfipb.QueryAnalysisRequest) error {
	if req.BuildFailure == nil {
		return status.Errorf(codes.InvalidArgument, "BuildFailure must not be empty")
	}
	if req.BuildFailure.GetBbid() == 0 {
		return status.Errorf(codes.InvalidArgument, "BuildFailure bbid must not be empty")
	}
	return nil
}
