// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package migrator

import (
	"context"
)

// API is the expected plugin interface from migrator plugins.
//
// The plugin may export any of the following functions by name.
type API interface {
	// FindProblems allows you to report problems about a Project, or about
	// certain configuration files within the project.
	//
	// If the method finds issues which warrant followup, it should use
	// proj.Report and/or proj.ConfigFiles()["filename"].Report. Reporting one or
	// more problems will cause the migrator tool to set up a checkout for this
	// project.
	//
	// Logging is set up for this context, and will be diverted to a per-project
	// logfile.
	//
	// This function should panic on error.
	FindProblems(ctx context.Context, proj Project)

	// ApplyFix allows you to attempt to automatically fix problems within a repo.
	//
	// Note that for real implementations you may want to keep details on the
	// `impl` struct; this will let you carry over information from
	// FindProblems.
	//
	// Logging is set up for this context, and will be diverted to a per-project
	// logfile.
	//
	// This function should panic on error.
	ApplyFix(ctx context.Context, repo Repo)
}

// InstantiateAPI is the symbol that plugins must export.
//
// It should return a new instance of API.
//
// If this returns nil, it has the effect of a plugin which:
//    FindProblems reports a generic problem "FindProblems not defined".
//    ApplyFix does nothing.
type InstantiateAPI func() API
