// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package track implements shared tracking functionality for the Tricium service modules.
//
// Overview diagram:
//
//    +-----------------+
//    |AnalyzeRequest   |
//    |id=<generated_id>|
//    +---+-------------+
//        |
//        +----------------------+
//        |                      |
//    +---+----------------+ +---+-------+
//    |AnalyzeRequestResult| |WorkflowRun|
//    |id=1                | |id=1       |
//    +---+----------------+ +-----------+
//                               |
//                               +----------------------+
//                               |                      |
//                           +---+-------------+ +---+-------------+
//                           |WorkflowRunResult| |AnalyzerRun      |
//                           |id=1             | |id=<analyzerName>|
//                           +-----------------+ +---+-------------+
//                                                   |
//                               +-------------------+
//                               |                   |
//                           +---+-------------+ +---+----------------------+
//                           |AnalyzerRunResult| |WorkerRun                 |
//                           |id=1             | |id=<analyzerName_platform>|
//                           +-----------------+ +---+----------------------+
//                                                   |
//                                          +--------+---------+
//                                          |                  |
//                                      +---+-----------+ +----+------------+
//                                      |WorkerRunResult| |Comment          |
//                                      |id=1           | |id=<generated_id>|
//                                      +---------------+ +----+------------+
//                                                             |
//                                          +------------------+
//                                          |                  |
//                                       +--+-------------+ +--+------------+
//                                       |CommentSelection| |CommentFeedback|
//                                       |id=1            | |id=2           |
//                                       +-----------   --+ +---------------+
//
package track

import (
	"time"

	ds "github.com/luci/gae/service/datastore"

	"infra/tricium/api/v1"
)

// AnalyzeRequest represents one Tricium Analyze RPC request.
//
// Immutable root entry.
type AnalyzeRequest struct {

	// LUCI datastore ID field with generated value.
	ID int64 `gae:"$id"`

	// Time when the corresponding request was received, time recorded in the reporter.
	Received time.Time

	// The project of the request.
	Project string

	// Reporter to use for progress updates and results.
	Reporter tricium.Reporter

	// File paths listed in the request.
	Paths []string `gae:",noindex"`

	// Git repository hosting files in the request.
	GitRepo string `gae:",noindex"`

	// Git ref to use in the Git repo.
	GitRef string `gae:",noindex"`
}

// AnalyzeRequestResult tracks the state of an Analyze request.
//
// Mutable entity.
type AnalyzeRequestResult struct {

	// LUCI datastore ID field with value 1.
	ID string `gae:"$id"`

	// LUCI datastore parent field with 'AnalyzeRequest' as parent.
	Parent *ds.Key `gae:"$parent"`

	// State of the Analyze request; running, success, or failure.
	State tricium.State
}

// WorkflowRun declares a request to execute a Tricium workflow.
//
// Immutable root of the complete workflow execution.
type WorkflowRun struct {

	// LUCI datastore ID field with value 1.
	ID string `gae:"$id"`

	// LUCI datastore parent field with 'AnalyzeRequest' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Name of analyzers included in this workflow.
	//
	// Included here to allow for direct access without queries.
	Analyzers []string

	// Isolate server URL.
	IsolateServerURL string `gae:",noindex"`

	// Swarming server URL.
	SwarmingServerURL string `gae:",noindex"`
}

// WorkflowRunResult tracks the state of a workflow run.
//
// Mutable entity.
type WorkflowRunResult struct {

	// LUCI datastore ID field with value 1.
	ID string `gae:"$id"`

	// LUCI datastore parent field with 'WorkflowRun' as parent.
	Parent *ds.Key `gae:"$parent"`

	// State of the parent request; running, success, or failure.
	//
	// This state is an aggregation of the run state of triggered analyzers.
	State tricium.State
}

// AnalyzerRun declares a request to execute an analyzer.
//
// Immutable entity.
type AnalyzerRun struct {

	// LUCI datastore ID field with the name of the analyzer.
	//
	// The workflow for a run may have several workers for one analyzer,
	// each running on different platforms.
	ID string `gae:"$id"`

	// LUCI datastore parent field with 'WorkflowRun' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Name of workers launched for this analyzer.
	//
	// Included her to allow for direct access without queries.
	Workers []string
}

// AnalyzerRunResult tracks the state of an analyzer run.
//
// Mutable entity.
type AnalyzerRunResult struct {

	// LUCI datastore ID field with value 1.
	ID string `gae:"$id"`

	// LUCI datastore parent field with 'AnalyzerRun' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Name of analyzer.
	Name string

	// State of the parent analyzer run; running, success, or failure.
	//
	// This state is an aggregation of the run state of triggered analyzer workers.
	State tricium.State
}

// WorkerRun declare a request to execute an analyzer worker.
//
// Immutable entity.
type WorkerRun struct {

	// LUCI datastore ID field with the name of the worker as value.
	//
	// This name is the same as that used in the workflow configuration.
	ID string `gae:"$id"`

	// LUCI datastore parent field with 'AnalyzerRun' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Platform this worker is producing results for.
	Platform tricium.Platform_Name

	// Names of workers succeeding this worker in the workflow.
	Next []string `gae:",noindex"`
}

// WorkerRunResult tracks the state of a worker run.
//
// Mutable entity.
type WorkerRunResult struct {

	// LUCI datastore ID field with value 1.
	ID string `gae:"$id"`

	// LUCI datastore parent field with 'WorkerRun' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Name of worker.
	Name string

	// Analyzer this worker is running.
	Analyzer string

	// Platform this worker is running on.
	Platform tricium.Platform_Name

	// State of the parent worker run; running, success, or failure.
	State tricium.State

	// Hash to the isolated input provided to the corresponding swarming task.
	IsolatedInput string `gae:",noindex"`

	// Hash to the isolated output collected from the corresponding swarming task.
	IsolatedOutput string `gae:",noindex"`

	SwarmingTaskID string `gae:",noindex"`

	// Exit code of the corresponding swarming task.
	ExitCode int

	// Number of comments produced by this worker.
	NumComments int

	// Tricium result encoded as JSON.
	Result string `gae:",noindex"`
}

// Comment tracks a comment generated by a worker.
//
// Immutable entity.
type Comment struct {

	// LUCI datastore ID field with generated value.
	ID string `gae:"$id"`

	// LUCI datastore parent field with 'WorkerRun' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Comment encoded as JSON.
	//
	// The comment should follow the tricium.Data_Comment format.
	// TODO(emso): Consider storing structured comment data.
	Comment []byte `gae:",noindex"`

	// Comment category with subcategories.
	//
	// This includes the analyzer name, e.g., clang-tidy/llvm-header-guard.
	Category string

	// Platforms this comment applies to.
	//
	// This is a int64 bit map using the tricium.Platform_Name number values for platforms.
	Platforms int64
}

// CommentSelection tracks selection of comments.
//
// When an analyzer has several workers running the analyzer using different configurations
// the resulting comments are merged to avoid duplication of results for users.
//
// Mutable entity.
type CommentSelection struct {

	// LUCI datastore ID field with value 1.
	ID string `gae:"$id"`

	// LUCI datastore parent field with 'Comment' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Whether this comments was included in the overall result of the enclosing request.
	//
	// All comments are included by default, but comments may need to be merged
	// in the case when comments for a category are produced for multiple platforms.
	Included bool
}

// CommentFeedback tracks 'not useful' user feedback for a comment.
//
// Mutable entity.
type CommentFeedback struct {

	// LUCI datastore ID field with value 1.
	ID string `gae:"$id"`

	// LUCI datastore parent field with 'Comment' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Number of 'not useful' clicks.
	NotUseful int

	// Links to more information about why the comment was found not useful.
	//
	// This should typically be a link to a Monorail issue.
	NotUsefulIssueURL []string
}
