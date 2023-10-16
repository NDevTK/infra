// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import (
	"context"
	"fmt"
	"infra/appengine/chrome-test-health/datastorage"
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

// Get function fetches the PostSubmitReport entity by creating key from the given args.
// See here for more details about PostSubmitReport entity:
// https://source.chromium.org/chromium/infra/infra/+/main:appengine/findit/model/code_coverage.py;drc=da0f8e0369a013173b31b6744b411c2bd9edd9df;l=331
func (p *PostsubmitReport) Get(ctx context.Context, client datastorage.IDataClient, host string, project string, ref string, revision string, bucket string, builder string, modifierID string) error {
	keyStr := fmt.Sprintf("%s$%s$%s$%s$%s$%s$%s", host, project, ref, revision, bucket, builder, modifierID)
	err := client.Get(ctx, p, "PostsubmitReport", keyStr)
	if err != nil {
		return fmt.Errorf("PostsubmitReport: %w", err)
	}
	return nil
}
