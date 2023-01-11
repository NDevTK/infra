// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package fake

import (
	"context"
	"fmt"
	"io/ioutil"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/proto/git"
	"go.chromium.org/luci/common/proto/gitiles"
	"google.golang.org/grpc"
)

// GitClient mocks the git.ClientInterface
type GitClient struct {
}

// GitTilesClient mocks the gitiles.GitilesClient
type GitTilesClient struct {
}

// GetFile mocks git.ClientInterface.GetFile()
func (gc *GitClient) GetFile(ctx context.Context, path string) (string, error) {
	if path == "test_git_path" {
		return GitData("../frontend/fake/dhcp_test.conf")
	} else if path == "test_enc_git_path" {
		return GitData("../frontend/fake/bots.cfg")
	}
	return "", errors.Reason("Unspecified mock path %s", path).Err()
}

// SwitchProject mocks git.ClientInterface.SwitchProject()
func (gc *GitClient) SwitchProject(ctx context.Context, project string) error {
	return nil
}

// Log mocks gitiles.GitilesClient.Log()
func (gc *GitTilesClient) Log(ctx context.Context, req *gitiles.LogRequest, opts ...grpc.CallOption) (res *gitiles.LogResponse, err error) {
	return &gitiles.LogResponse{
		Log: []*git.Commit{
			{Id: fmt.Sprintf("%s-%s", req.Project, req.Committish)},
		},
	}, nil
}

// GitData mocks a git file content based on a given filepath
func GitData(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
