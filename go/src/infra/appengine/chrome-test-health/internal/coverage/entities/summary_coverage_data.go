// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

type SummaryCoverageData struct {
	DataType                string `datastore:"data_type"`
	Path                    string `datastore:"path"`
	Bucket                  string `datastore:"bucket"`
	Builder                 string `datastore:"builder"`
	ModifierId              int64  `datastore:"modifier_id"`
	Data                    []byte `datastore:"data"`
	GitilesCommitProject    string `datastore:"gitiles_commit.project"`
	GitilesCommitRef        string `datastore:"gitiles_commit.ref"`
	GitilesCommitRevision   string `datastore:"gitiles_commit.revision"`
	GitilesCommitServerHost string `datastore:"gitiles_commit.server_host"`
}
