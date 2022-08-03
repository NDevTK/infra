// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/appengine/weetbix/internal/config"
	"infra/appengine/weetbix/internal/perms"
	pb "infra/appengine/weetbix/proto/v1"
)

type projectServer struct{}

func NewProjectsServer() *pb.DecoratedProjects {
	return &pb.DecoratedProjects{
		Prelude:  checkAllowedPrelude,
		Service:  &projectServer{},
		Postlude: gRPCifyAndLogPostlude,
	}
}

func (*projectServer) GetConfig(ctx context.Context, req *pb.GetProjectConfigRequest) (*pb.ProjectConfig, error) {
	project, err := parseProjectConfigName(req.Name)
	if err != nil {
		return nil, invalidArgumentError(errors.Annotate(err, "name").Err())
	}

	if err := perms.VerifyProjectPermissions(ctx, project, perms.PermGetConfig); err != nil {
		return nil, err
	}

	// Fetch a recent project configuration.
	// (May be a recent value that was cached.)
	cfg, err := readProjectConfig(ctx, project)
	if err != nil {
		return nil, err
	}

	response := &pb.ProjectConfig{
		Name: fmt.Sprintf("projects/%s/config", project),
		Monorail: &pb.ProjectConfig_Monorail{
			Project:       cfg.Config.Monorail.Project,
			DisplayPrefix: cfg.Config.Monorail.DisplayPrefix,
		},
	}
	return response, nil
}

func (*projectServer) List(ctx context.Context, request *pb.ListProjectsRequest) (*pb.ListProjectsResponse, error) {
	projects, err := config.Projects(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "fetching project configs").Err()
	}

	readableProjects := make([]string, 0, len(projects))
	for project := range projects {
		hasAccess, err := perms.HasProjectPermission(ctx, project, perms.PermGetConfig)
		if err != nil {
			return nil, err
		}
		if hasAccess {
			readableProjects = append(readableProjects, project)
		}
	}

	// Return projects in a stable order.
	sort.Strings(readableProjects)

	return &pb.ListProjectsResponse{
		Projects: createProjectPbs(readableProjects),
	}, nil
}

func createProjectPbs(projects []string) []*pb.Project {
	projectsPbs := make([]*pb.Project, 0, len(projects))
	for _, project := range projects {
		projectsPbs = append(projectsPbs, &pb.Project{
			Name:        fmt.Sprintf("projects/%s", project),
			DisplayName: strings.Title(project),
			Project:     project,
		})
	}
	return projectsPbs
}
