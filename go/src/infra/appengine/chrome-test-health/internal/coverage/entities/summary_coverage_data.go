// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/datastore"
)

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

// Get the SummaryCoverageData entity by creating key from the given args.
// See here for more details about SummaryCoverageData's format:
// https://source.chromium.org/chromium/infra/infra/+/main:appengine/findit/model/code_coverage.py;l=439
func (s *SummaryCoverageData) Get(ctx context.Context, client *datastore.Client, host string, project string, ref string, revision string, dataType string, path string, bucket string, builder string) error {
	keyStr := fmt.Sprintf("%s$%s$%s$%s$%s$%s$%s$%s$0", host, project, ref, revision, dataType, path, bucket, builder)
	keyLiteral := datastore.NameKey("SummaryCoverageData", keyStr, nil)
	err := client.Get(ctx, keyLiteral, s)

	if err != nil {
		return errors.New("Unable to fetch record with the given arguments")
	}

	return nil
}
