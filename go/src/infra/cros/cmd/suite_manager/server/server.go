// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package server implements the suite_manager grpc service.
package server

import (
	"context"
	"log/slog"

	smpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/suite_manager"
)

// InitServer returns an active SuiteManagerServiceServer type.
func InitServer() smpb.SuiteManagerServiceServer {
	slog.Info("Initializing SuiteManager Service client.")

	return &server{}
}

// server implements the SuiteManagerServiceServer from the suite_manager proto
// definition.
type server struct {
	smpb.UnimplementedSuiteManagerServiceServer
}

// GetConfig fetches a single config from the long term storage location for
// SuiteScheduler configs.
func (s *server) GetConfig(ctx context.Context, request *smpb.GetConfigRequest) (*smpb.GetConfigResponse, error) {
	slog.Info("Called GetConfig()")
	return nil, nil
}

// FetchConfigs returns SuiteScheduler configs from long term storage based
// on the filter parameters provided.
func (s *server) FetchConfigs(ctx context.Context, request *smpb.FetchConfigsRequest) (*smpb.FetchConfigsResponse, error) {
	slog.Info("Called FetchConfigs()")
	return nil, nil
}

// AddConfig adds a new config to long term storage.
//
// NOTE: Duplicate config names are not allowed.
func (s *server) AddConfig(ctx context.Context, request *smpb.AddConfigRequest) (*smpb.AddConfigResponse, error) {
	slog.Info("Called AddConfig()")
	return nil, nil
}

// DeleteConfig removes a config from long term storage
func (s *server) DeleteConfig(ctx context.Context, request *smpb.DeleteConfigRequest) (*smpb.DeleteConfigResponse, error) {
	slog.Info("Called DeleteConfig()")
	return nil, nil
}

// UpdateConfig updates fields in a config already in long term storage.
func (s *server) UpdateConfig(ctx context.Context, request *smpb.UpdateConfigRequest) (*smpb.UpdateConfigResponse, error) {
	slog.Info("Called UpdateConfig()")
	return nil, nil
}

// FetchBuildTargets returns a list of build targets according to the current
// lab_config.cfg.
func (s *server) FetchBuildTargets(ctx context.Context, request *smpb.FetchBuildTargetsRequest) (*smpb.FetchBuildTargetsResponse, error) {
	slog.Info("Called FetchBuildTargets()")
	return nil, nil
}

// EstimateImpact returns the estimated impact of a config in hours. If
// multiple configs are given then it returns the difference between the two
// using the logic:  <config impact> - <compare_config impact> = diff.
func (s *server) EstimateImpact(ctx context.Context, request *smpb.EstimateImpactRequest) (*smpb.EstimateImpactResponse, error) {
	slog.Info("Called EstimateImpact()")
	return nil, nil
}

// BatchEstimateImpact returns the estimated impact of a config in hours. If
// multiple configs are given then it returns the difference between the two
// using the logic:  <config impact> - <compare_config impact> = diff
func (s *server) BatchEstimateImpact(ctx context.Context, request *smpb.BatchEstimateImpactRequest) (*smpb.BatchEstimateImpactResponse, error) {
	slog.Info("Called BatchEstimateImpact()")
	return nil, nil
}

// ProposeChange adds a change proposal into the active queue.
//
// TODO(b/339518338): Implement after a new control flow is worked out for
// proposal storage and accessing.
//
// NOTE: This service should disallow multiple open proposals for a single
// config.
func (s *server) ProposeChange(ctx context.Context, request *smpb.ProposeChangeRequest) (*smpb.ProposeChangeResponse, error) {
	slog.Info("Called ProposeChange()")
	return nil, nil
}

// UpdateProposal modifies an active proposal.
//
// TODO(b/339518338): Implement after a new control flow is worked out for
// proposal storage and accessing.
func (s *server) UpdateProposal(ctx context.Context, request *smpb.UpdateProposalRequest) (*smpb.UpdateProposalResponse, error) {
	slog.Info("Called UpdateProposal()")
	return nil, nil
}
