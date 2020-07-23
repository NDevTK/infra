// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"

	m "infra/tools/migrator"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging"
)

type impl struct{}

// FindProblems allows you to report problems about a Project, or about
// certain configuration files within the project.
//
// If the method finds issues which warrant followup, it should use proj.Report
// and/or proj.ConfigFiles()["filename"].Report. Reporting one or more problems
// will cause the migrator tool to set up a checkout for this project.
//
// Logging is set up for this context, and will be diverted to a per-project
// logfile.
//
// This function should panic on error.
func (impl) FindProblems(ctx context.Context, proj m.Project) {
	if proj.ID() == "chromium" {
		proj.Report("CHROMIUM", "it's chromium, bruh")
		return
	}

	logging.Infof(ctx, "Finding problems in %s", proj.ID())
	cfgFile, ok := proj.ConfigFiles()["cr-buildbucket.cfg"]
	if !ok {
		logging.Infof(ctx, "No cr-buildbucket.cfg")
		return
	}

	bbConfig := &bbpb.BuildbucketCfg{}
	cfgFile.TextPb(bbConfig)

	for _, b := range bbConfig.Buckets {
		if b.GetSwarming().GetBuilderDefaults() != nil {
			cfgFile.Report(
				"BUILDER_DEFAULTS",
				fmt.Sprintf("Bucket %s defines builder defaults.", b.Name),
				m.MetadataOption("bucketname", b.Name))
		}
		for _, sw := range b.GetSwarming().GetBuilders() {
			if len(sw.Mixins) != 0 {
				cfgFile.Report(
					"BUILDER_MIXINS",
					fmt.Sprintf("Builder %s/%s uses mixins.", b.Name, sw.Name),
					m.MetadataOption("bucketname", b.Name),
					m.MetadataOption("buildername", sw.Name))
			}
		}
	}
}

// ApplyFix allows you to attempt to automatically fix problems within a repo.
//
// Note that for real implementations you may want to keep details on the `impl`
// struct; this will let you carry over information from ReportProblems.
//
// Logging is set up for this context, and will be diverted to a per-project
// logfile.
//
// This function should panic on error.
func (impl) ApplyFix(ctx context.Context, repo m.Repo) {
	sh := repo.Shell()
	if sh.Stat("main.star") != nil {
		sh.Run("./main.star", m.TieStderr)
	}
}

// InstantiateAPI implements the migrator's plugin API.
func InstantiateAPI() m.API { return impl{} }

// Type assertion to make sure it's type-conformant
var _ m.InstantiateAPI = InstantiateAPI

func main() {
	panic("This is meant to be run as a Go plugin for the `migrator` tool.")
}
