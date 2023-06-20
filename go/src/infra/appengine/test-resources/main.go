// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/proto/protowalk"
	"go.chromium.org/luci/grpc/appstatus"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/openid"
	"go.chromium.org/luci/server/auth/rpcacl"
	"go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/module"

	"infra/appengine/test-resources/api"
	"infra/appengine/test-resources/internal/testmetrics"
)

const (
	dataOwnersGroup = "mdb/chrome-browser-infra"
)

var (
	stats *testResourcesServer
	// RPC-level ACLs.
	rpcACL = rpcacl.Map{
		"/discovery.Discovery/*":                      rpcacl.All,
		"/test_resources.Stats/UpdateMetricsTable":    dataOwnersGroup,
		"/test_resources.Stats/ListComponents":        rpcacl.All,
		"/test_resources.Stats/FetchTestMetrics":      rpcacl.All,
		"/test_resources.Stats/FetchDirectoryMetrics": rpcacl.All,
	}
	// Data set to work with
	dataSet = flag.String(
		"data-set",
		"test_results",
		"The data set to use (e.g. test_results_test for testing).",
	)
)

type Client interface {
	UpdateSummary(ctx context.Context, fromDate civil.Date, toDate civil.Date) error
	ListComponents(ctx context.Context, req *api.ListComponentsRequest) (*api.ListComponentsResponse, error)
	FetchMetrics(ctx context.Context, req *api.FetchTestMetricsRequest) (*api.FetchTestMetricsResponse, error)
	FetchDirectoryMetrics(ctx context.Context, req *api.FetchDirectoryMetricsRequest) (*api.FetchDirectoryMetricsResponse, error)
}

func main() {
	modules := []module.Module{
		cron.NewModuleFromFlags(),
	}
	server.Main(nil, modules, func(srv *server.Server) error {
		client, err := setupClient(srv)
		if err != nil {
			return err
		}
		stats = &testResourcesServer{
			Client: client,
		}
		srv.Options.DefaultRequestTimeout = time.Minute * 10
		cron.RegisterHandler("update-daily-summary", updateDailySummary)

		// All RPC APIs.
		api.RegisterStatsServer(srv, stats)

		// Authentication methods for RPC APIs.
		srv.SetRPCAuthMethods([]auth.Method{
			// The preferred authentication method.
			&openid.GoogleIDTokenAuthMethod{
				AudienceCheck: openid.AudienceMatchesHost,
				SkipNonJWT:    true, // pass OAuth2 access tokens through
			},
			// Backward compatibility for the RPC Explorer and old clients.
			&auth.GoogleOAuth2Method{
				Scopes: []string{"https://www.googleapis.com/auth/userinfo.email"},
			},
		})

		// Per-RPC authorization interceptor.
		srv.RegisterUnifiedServerInterceptors(rpcacl.Interceptor(rpcACL))
		return nil
	})
}

func updateDailySummary(ctx context.Context) error {
	previousDateTime := time.Now().AddDate(0, 0, -1)
	date := previousDateTime.Format("2006-01-02")

	cDate, err := civil.ParseDate(date)
	if err != nil {
		logging.Errorf(ctx, "Failed parsing current date: %s", err)
		return err
	}

	err = stats.Client.UpdateSummary(ctx, cDate, cDate)
	if err != nil {
		logging.Errorf(ctx, "Failed updating current date: %s", err)
		return err
	}
	return nil
}

func setupClient(srv *server.Server) (*testmetrics.Client, error) {
	bqClient, err := bigquery.NewClient(srv.Context, srv.Options.CloudProject)
	if err != nil {
		return nil, err
	}
	var client = &testmetrics.Client{
		BqClient:  bqClient,
		ProjectId: srv.Options.CloudProject,
		DataSet:   *dataSet,
	}
	err = client.Init()
	if err != nil {
		return nil, err
	}
	return client, nil
}

type testResourcesServer struct {
	Client Client
}

func (s *testResourcesServer) UpdateMetricsTable(ctx context.Context, req *api.UpdateMetricsTableRequest) (*api.UpdateMetricsTableResponse, error) {
	if err := validateRequest(ctx, req); err != nil {
		return nil, appstatus.Errorf(codes.InvalidArgument, "%s", err.Error())
	}
	fromDate, err := civil.ParseDate(req.FromDate)
	if err != nil {
		return nil, appstatus.Errorf(codes.InvalidArgument, "%s", err.Error())
	}
	toDate, err := civil.ParseDate(req.ToDate)
	if err != nil {
		return nil, appstatus.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	err = s.Client.UpdateSummary(ctx, fromDate, toDate)

	if err != nil {
		return nil, err
	}
	return &api.UpdateMetricsTableResponse{}, nil
}

func (s *testResourcesServer) ListComponents(ctx context.Context, req *api.ListComponentsRequest) (*api.ListComponentsResponse, error) {
	resp, err := s.Client.ListComponents(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *testResourcesServer) FetchDirectoryMetrics(ctx context.Context, req *api.FetchDirectoryMetricsRequest) (*api.FetchDirectoryMetricsResponse, error) {
	if err := validateRequest(ctx, req); err != nil {
		return nil, appstatus.Errorf(codes.InvalidArgument, "%s", err.Error())
	}
	resp, err := s.Client.FetchDirectoryMetrics(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *testResourcesServer) FetchTestMetrics(ctx context.Context, req *api.FetchTestMetricsRequest) (*api.FetchTestMetricsResponse, error) {
	if err := validateRequest(ctx, req); err != nil {
		return nil, appstatus.Errorf(codes.InvalidArgument, "%s", err.Error())
	}
	resp, err := s.Client.FetchMetrics(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func validateRequest(ctx context.Context, req proto.Message) error {
	if procRes := protowalk.Fields(req, &protowalk.RequiredProcessor{}); procRes != nil {
		if resStrs := procRes.Strings(); len(resStrs) > 0 {
			logging.Infof(ctx, strings.Join(resStrs, ". "))
		}
		if err := procRes.Err(); err != nil {
			return err
		}
	}
	return nil
}
