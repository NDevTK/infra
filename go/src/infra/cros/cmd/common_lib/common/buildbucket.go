// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"fmt"
	"net/http"

	"go.chromium.org/luci/auth"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"
)

// NewBBClient creates new bb client.
func NewBBClient(ctx context.Context) (buildbucketpb.BuildsClient, error) {
	hClient, err := HttpClient(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "create buildbucket client").Err()
	}
	pClient := &prpc.Client{
		C:    hClient,
		Host: "cr-buildbucket.appspot.com",
	}
	return buildbucketpb.NewBuildsPRPCClient(pClient), nil
}

// HttpClient creates a http client.
func HttpClient(ctx context.Context) (*http.Client, error) {
	a := auth.NewAuthenticator(ctx, auth.SilentLogin, auth.Options{
		Scopes: []string{auth.OAuthScopeEmail},
	})
	h, err := a.Client()
	if err != nil {
		return nil, errors.Annotate(err, "create http client").Err()
	}
	return h, nil
}

// TestRunnerBuilderID returns builderid for test_runner.
func TestRunnerBuilderID() *buildbucketpb.BuilderID {
	return &buildbucketpb.BuilderID{
		Project: "chromeos",
		Bucket:  "test_runner",
		Builder: "test_runner",
	}
}

// BBUrl returns the Buildbucket URL of the task.
func BBUrl(builderID *buildbucketpb.BuilderID, bbId int64) string {
	return fmt.Sprintf("https://ci.chromium.org/p/%s/builders/%s/%s/b%d", builderID.Project, builderID.Bucket, builderID.Builder, bbId)
}
