// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import (
	"time"
)

type DependencyRepository struct {
	Path       string `datastore:"path"`
	ServerHost string `datastore:"server_host"`
	Project    string `datastore:"project"`
	Revision   string `datastore:"revision"`
}

type PostsubmitReport struct {
	Bucket                  string                 `datastore:"bucket"`
	BuildId                 int64                  `datastore:"build_id"`
	Builder                 string                 `datastore:"builder"`
	CommitPosition          int64                  `datastore:"commit_position"`
	CommitTimestamp         time.Time              `datastore:"commit_timestamp"`
	GitilesCommitProject    string                 `datastore:"gitiles_commit.project"`
	GitilesCommitRef        string                 `datastore:"gitiles_commit.ref"`
	GitilesCommitRevision   string                 `datastore:"gitiles_commit.revision"`
	GitilesCommitServerHost string                 `datastore:"gitiles_commit.server_host"`
	Manifest                []DependencyRepository `datastore:"manifest"`
	ModifierId              int64                  `datastore:"modifier_id"`
	SummaryMetrics          []byte                 `datastore:"summary_metrics"`
	Visible                 bool                   `datastore:"visible"`
}
