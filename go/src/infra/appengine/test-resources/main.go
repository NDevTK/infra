// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"

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
		"/discovery.Discovery/*":                   rpcacl.All,
		"/test_resources.Stats/UpdateMetricsTable": dataOwnersGroup,
	}
)

type Client interface {
	UpdateSummary(context.Context, *api.UpdateMetricsTableRequest) (*api.UpdateMetricsTableResponse, error)
	UpdateDateSummary(context.Context, civil.Date) error
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
		return err
	}

	err = stats.Client.UpdateDateSummary(ctx, cDate)
	if err != nil {
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
		BqClient: bqClient,
	}
	err = client.Init(srv.Options.CloudProject)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type testResourcesServer struct {
	Client Client
}

func (s *testResourcesServer) UpdateMetricsTable(ctx context.Context, req *api.UpdateMetricsTableRequest) (*api.UpdateMetricsTableResponse, error) {
	resp, err := s.Client.UpdateSummary(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *testResourcesServer) ListComponents(ctx context.Context, req *api.ListComponentsRequest) (*api.ListComponentsResponse, error) {
	panic("Endpoint has not been implemented yet")
}

func (s *testResourcesServer) FetchDirectoryMetrics(ctx context.Context, req *api.FetchDirectoryMetricsRequest) (*api.FetchDirectoryMetricsResponse, error) {
	panic("Endpoint has not been implemented yet")
}

func (s *testResourcesServer) FetchTestMetrics(ctx context.Context, req *api.FetchTestMetricsRequest) (*api.FetchTestMetricsResponse, error) {
	panic("Endpoint has not been implemented yet")
}
