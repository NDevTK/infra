// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package external

import (
	"context"
	"net/http"

	authclient "go.chromium.org/luci/auth"
	gitilesapi "go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/proto/gitiles"
	"go.chromium.org/luci/server/auth"
	"google.golang.org/grpc"
)

// GitTilesClient exposes a subset of gitiles.GitilesClient
type GitTilesClient interface {
	Log(ctx context.Context, in *gitiles.LogRequest, opts ...grpc.CallOption) (*gitiles.LogResponse, error)
	DownloadFile(ctx context.Context, in *gitiles.DownloadFileRequest, opts ...grpc.CallOption) (*gitiles.DownloadFileResponse, error)
}

type gitTilesClientImpl struct {
	client gitiles.GitilesClient
}

// Log implements gitiles.GitilesClient.Log()
func (gc *gitTilesClientImpl) Log(ctx context.Context, req *gitiles.LogRequest) (*gitiles.LogResponse, error) {
	return gc.client.Log(ctx, req)
}

// DownloadFile implements gitiles.GitilesClient.DownloadFile()
func (gc *gitTilesClientImpl) DownloadFile(ctx context.Context, in *gitiles.DownloadFileRequest) (*gitiles.DownloadFileResponse, error) {
	return gc.client.DownloadFile(ctx, in)
}

// GetGitilesClient returns the GitilesClient for the given host.
func GetGitilesClient(ctx context.Context, gitilesHost string) (gitiles.GitilesClient, error) {
	t, err := auth.GetRPCTransport(ctx, auth.AsSelf, auth.WithScopes(authclient.OAuthScopeEmail, gitilesapi.OAuthScope))
	if err != nil {
		return nil, err
	}
	return gitilesapi.NewRESTClient(&http.Client{Transport: t}, gitilesHost, true)
}
