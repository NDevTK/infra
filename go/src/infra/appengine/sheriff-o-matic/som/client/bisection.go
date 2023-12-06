// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package client

import (
	"context"
	"net/http"

	bisectionpb "go.chromium.org/luci/bisection/proto/v1"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/auth"
)

type BisectionClient struct {
	ServiceClient bisectionpb.AnalysesClient
}

func (cl *BisectionClient) QueryBisectionResults(c context.Context, bbid int64, stepName string) (*bisectionpb.QueryAnalysisResponse, error) {
	logging.Infof(c, "Querying LUCI Bisection results for build %d", bbid)
	req := &bisectionpb.QueryAnalysisRequest{
		BuildFailure: &bisectionpb.BuildFailure{
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

func (cl *BisectionClient) BatchGetTestAnalyses(c context.Context, req *bisectionpb.BatchGetTestAnalysesRequest) (*bisectionpb.BatchGetTestAnalysesResponse, error) {
	return cl.ServiceClient.BatchGetTestAnalyses(c, req)
}

func NewBisectionServiceClient(c context.Context, host string) (bisectionpb.AnalysesClient, error) {
	t, err := auth.GetRPCTransport(c, auth.AsSelf)
	if err != nil {
		return nil, err
	}
	return bisectionpb.NewAnalysesPRPCClient(
		&prpc.Client{
			C:       &http.Client{Transport: t},
			Host:    host,
			Options: prpc.DefaultOptions(),
		}), nil
}
