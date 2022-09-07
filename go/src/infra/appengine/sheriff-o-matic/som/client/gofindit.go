// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package client

import (
	"context"
	"net/http"

	gfipb "go.chromium.org/luci/bisection/proto"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/auth"
)

const (
	gofinditHost = "chops-gofindit-dev.appspot.com"
)

type GoFinditClient struct {
	ServiceClient gfipb.GoFinditServiceClient
}

func (cl *GoFinditClient) QueryGoFinditResults(c context.Context, bbid int64, stepName string) (*gfipb.QueryAnalysisResponse, error) {
	logging.Infof(c, "Querying GoFindit result for build %d", bbid)
	req := &gfipb.QueryAnalysisRequest{
		BuildFailure: &gfipb.BuildFailure{
			Bbid:           bbid,
			FailedStepName: stepName,
		},
	}

	res, err := cl.ServiceClient.QueryAnalysis(c, req)
	if err != nil {
		logging.Errorf(c, "Cannot query analysis for build %d: %s", bbid, err)
		return nil, err
	}
	return res, nil
}

func NewGoFinditServiceClient(c context.Context, host string) (gfipb.GoFinditServiceClient, error) {
	t, err := auth.GetRPCTransport(c, auth.AsSelf)
	if err != nil {
		return nil, err
	}
	return gfipb.NewGoFinditServicePRPCClient(
		&prpc.Client{
			C:       &http.Client{Transport: t},
			Host:    host,
			Options: prpc.DefaultOptions(),
		}), nil
}
