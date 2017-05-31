// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package track implements shared tracking functionality for the Tricium service modules.
//
// Overview diagram:
//
// +-----------------+
// |Run              |
// |id=<generated_id>|
// +-----------------+
//     |
//     +-----------+
//     |           |
// +---+-----+ +---+--------------+
// |RunResult| |AnalyzerRun       |
// |id=1     | |id=<analyzer_name>|
// +---------+ +------------------+
//                 |
//                 +----------------+
//                 |                |
//             +---+----------+ +---+-----------------------+
//             |AnalyzerResult| |WorkerRun                  |
//             |id=1          | |id=<analyzer_name-platform>|
//             +--------------+ +---------------------------+
//                                   |
//                                   +--------------+
//                                   |              |
//                               +---+--------+ +---+-------------+
//                               |WorkerResult| |Comment          |
//                               |id=1        | |id=<generated id>|
//                               +------------+ +----+------------+
//                                                   |
//                                                   |
//                                              +----+--------+
//                                              |CommentResult|
//                                              |id=1         |
//                                              +-------------+
//
package track

import (
	"time"

	ds "github.com/luci/gae/service/datastore"

	"infra/tricium/api/v1"
)

// Run tracks the processing of one analysis request.
//
// Immutable entry with no parent.
type Run struct {

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

// RunResult tracks the state of a run.
//
// Mutable entry.
type RunResult struct {

	// LUCI datastore ID field with value 1.
	ID     string  `gae:"$id"`

	// LUCI datastore parent field with 'Run' as parent.
	Parent *ds.Key `gae:"$parent"`

	// State of the parent run; running, success, or failure.
	//
	// This state is an aggregation of the run state of triggered analyzers.
	State tricium.State
}

// AnalyzerRun tracks the execution of an analyzer.
//
// Immutable entry.
type AnalyzerRun struct {

	// LUCI datastore ID field with the name of the analyzer.
	// 
	// The workflow for a run may have several workers for one analyzer,
	// each running on different platforms.
	ID     string  `gae:"$id"`

	// LUCI datastore parent field with 'Run' as parent.
	Parent *ds.Key `gae:"$parent"`
}

// AnalyzerResult tracks the state of an analyzer run.
//
// Mutable entry.
type AnalyzerResult struct {

	// LUCI datastore ID field with value 1.
	ID     string  `gae:"$id"`

	// LUCI datastore parent field with 'AnalyzerRun' as parent.
	Parent *ds.Key `gae:"$parent"`

	// State of the parent analyzer run; running, success, or failure.
	//
	// This state is an aggregation of the run state of triggered analyzer workers.
	State tricium.State
}

// WorkerRun tracks the execution of an analyzer worker.
//
// Immutable entry.
type WorkerRun struct {

	// LUCI datastore ID field with the name of the worker as value.
	//
	// This name is the same as that used in the workflow configuration.
	ID     string  `gae:"$id"`

	// LUCI datastore parent field with 'AnalyzerRun' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Platform this worker is producing results for.
	Platform tricium.Platform_Name

	// Names of workers succeeding this worker in the workflow.
	Next []string `gae:",noindex"`

	// Isolate server URL.
	IsolateServerURL string `gae:",noindex"`

	// Swarming server URL.
	SwarmingURL string `gae:",noindex"`
}

// WorkerResult tracks the state of a worker run.
//
// Mutable entry.
type WorkerResult struct {

	// LUCI datastore ID field with value 1.
	ID string  `gae:"$id"`

	// LUCI datastore parent field with 'WorkerRun' as parent.
	Parent *ds.Key `gae:"$parent"`

	// State of the parent worker run; running, success, or failure.
	State tricium.State

	// Hash to the isolated input provided to the corresponding swarming task.
	IsolatedInput string `gae:",noindex"`

	// Hash to the isolated output collected from the corresponding swarming task.
	IsolatedOutput string `gae:",noindex"`

	// Swarming task ID.
	TaskID string `gae:",noindex"`

	// Exit code of the corresponding swarming task.
	ExitCode int

	// Number of comments produced by this worker.
	NumComments int

	// Tricium result encoded as JSON.
	Result string `gae:",noindex"`
}

// Comment tracks a result comment from a worker.
//
// Immutable entry.
type Comment struct {

	// LUCI datastore ID field with generated value.
	ID     string  `gae:"$id"`

	// LUCI datastore parent field with 'WorkerRun' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Comment encoded as JSON.
	//
	// The comment should follow the tricium.Data_Comment format.
	// TODO(emso): Consider storing structured comment data.
	Comment string `gae:",noindex"`

	// Comment category with subcategories.
	//
	// This includes the analyzer name, e.g., clang-tidy/llvm-header-guard.
	Category string

	// Platforms this comment applies to.
	//
	// This is a int64 bit map using the tricium.Platform_Name number values for platforms.
	Platforms int64
}

// CommentResult tracks inclusion and feedback for a comment.
//
// If an analyzer runs workers on more than one platform, then the results
// for the analyzer are merged, including a selection of comments via the 'Included'
// field.
//
// Mutable entry.
type CommentResult struct {

	// LUCI datastore ID field with value 1.
	ID     string  `gae:"$id"`

	// LUCI datastore parent field with 'Comment' as parent.
	Parent *ds.Key `gae:"$parent"`

	// Whether this comments was included in the overall result of the enclosing run.
	//
	// All comments are included by default, but comments may need to be merged
	// in the case when comments for a category are produced for multiple platforms.
	Included bool

	// Number of 'not useful' clicks.
	NotUseful int

	// Links to more information about why the comment was found not useful.
	//
	// This should typically be a link to a Monorail issue.
	NotUsefulIssueURL []string
}
