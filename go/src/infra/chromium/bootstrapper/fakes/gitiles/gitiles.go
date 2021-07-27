// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gitiles

import (
	"context"
	"fmt"
	"infra/chromium/bootstrapper/gitiles"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/proto/git"
	gitilespb "go.chromium.org/luci/common/proto/gitiles"
	"google.golang.org/grpc"
)

type Revision struct {
	// Parent is the commit ID of the parent revision.
	//
	// If Parent is empty, the root revision of the project will be the
	// parent revision (except for the root revision of the project).
	Parent string

	// Files maps file paths to their new contents at the revision.
	//
	// For non-root revisions, missing keys will have the same contents as
	// in the parent revision. For root-revisions, a missing key indicates
	// the file does not exist at the revision. A nil value indicates the
	// file does not exist at the revision.
	Files map[string]*string
}

// Project is the fake data for a gitiles project.
type Project struct {
	// Refs maps refs to their revision.
	//
	// Missing keys will have a default revision computed. An empty string
	// value indicates that the ref does not exist.
	Refs map[string]string

	// Revisions maps commit IDs to the revision of the repo.
	//
	// Missing keys will have a default fake revision. A nil value indicates
	// no revision with the commit ID exists.
	Revisions map[string]*Revision

	// The commit ID of the project's root revision
	//
	// If empty, the root revision will be an empty commit.
	RootId string
}

// Host is the fake data for a gitiles host.
type Host struct {
	// Projects maps project names to their details.
	//
	// Missing keys will have a default fake project. A nil value indicates
	// the the project does not exist.
	Projects map[string]*Project
}

// Client is the client that will serve fake data for a given host.
type Client struct {
	hostname string
	gitiles  *Host
}

// Factory creates a factory that returns RPC clients that use fake data to
// respond to requests.
//
// The fake data is taken from the fakes argument, which is a map from host
// names to the Host instances containing the fake data for the host. Missing
// keys will have a default Host. A nil value indicates that the given host is
// not a gitiles instance.
func Factory(fakes map[string]*Host) gitiles.GitilesClientFactory {
	return func(ctx context.Context, host string) (gitiles.GitilesClient, error) {
		fake, ok := fakes[host]
		if !ok {
			fake = &Host{}
		} else if fake == nil {
			return nil, errors.Reason("%s is not a gitiles host", host).Err()
		}
		return &Client{host, fake}, nil
	}
}

func (c *Client) getProject(projectName string) (*Project, error) {
	project, ok := c.gitiles.Projects[projectName]
	if !ok {
		project = &Project{}
	} else if project == nil {
		return nil, errors.Reason("unknown project %#v on host %#v", projectName, c.hostname).Err()
	}
	return project, nil
}

func (c *Client) Log(ctx context.Context, request *gitilespb.LogRequest, options ...grpc.CallOption) (*gitilespb.LogResponse, error) {
	if request.PageSize != 1 {
		panic(errors.Reason("unexpected page_size in LogRequest: %d", request.PageSize))
	}
	project, err := c.getProject(request.Project)
	if err != nil {
		return nil, err
	}
	commitId, ok := project.Refs[request.Committish]
	if !ok {
		commitId = fmt.Sprintf("fake-revision|%s|%s|%s", c.hostname, request.Project, request.Committish)
	} else if commitId == "" {
		return nil, errors.Reason("unknown ref %#v for project %#v on host %#v", request.Committish, request.Project, c.hostname).Err()
	}
	return &gitilespb.LogResponse{
		Log: []*git.Commit{
			{
				Id: commitId,
			},
		},
	}, nil
}

func (c *Client) DownloadFile(ctx context.Context, request *gitilespb.DownloadFileRequest, options ...grpc.CallOption) (*gitilespb.DownloadFileResponse, error) {
	project, err := c.getProject(request.Project)
	if err != nil {
		return nil, err
	}
	revision, ok := project.Revisions[request.Committish]
	if !ok {
		revision = &Revision{}
	} else if revision == nil {
		return nil, errors.Reason("unknown revision %#v of project %#v on host %#v", request.Committish, request.Project, c.hostname).Err()
	}
	for {
		_, ok = revision.Files[request.Path]
		if ok {
			break
		}
		parent := revision.Parent
		if parent == "" || parent == project.RootId {
			revision, ok = project.Revisions[project.RootId]
			if !ok {
				revision = &Revision{}
			} else if revision == nil {
				panic("root revision is set to non-existent revision")
			}
			break
		}
		revision, ok = project.Revisions[parent]
		if !ok {
			revision = &Revision{}
		} else if revision == nil {
			panic("parent revision is set to non-existent revision")
		}
	}
	contents := revision.Files[request.Path]
	if contents == nil {
		return nil, errors.Reason("unknown file %#v at revision %#v of project %#v on host %#v", request.Path, request.Committish, request.Project, c.hostname).Err()
	}
	return &gitilespb.DownloadFileResponse{
		Contents: *contents,
	}, nil
}
