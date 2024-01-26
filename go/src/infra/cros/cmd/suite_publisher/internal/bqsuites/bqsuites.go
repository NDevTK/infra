// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package bqsuites provides a common interface for publishing
// Suites and SuiteSets to BigQuery.
package bqsuites

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"

	"infra/cros/cmd/suite_publisher/internal/suite"
)

// PublishInfo is the information needed to publish a Suite or SuiteSet
// to BigQuery.
type PublishInfo struct {
	Suite         suite.CentralizedSuite
	BuildTarget   string
	CrosMilestone string
	CrosVersion   string
}

// PublishSuite publishes a suite to BigQuery using the provided inserter.
func PublishSuite(ctx context.Context, inserter *bigquery.Inserter, suite *PublishInfo) error {
	return inserter.Put(ctx, suite)
}

// Save implements the ValueSaver interface so that write operations can write
// from a PublishInfo struct to the database.
func (p *PublishInfo) Save() (map[string]bigquery.Value, string, error) {

	ret := map[string]bigquery.Value{
		"id":             p.Suite.ID(),
		"cros_milestone": p.CrosMilestone,
		"cros_version":   p.CrosVersion,
		"build_target":   p.BuildTarget,
	}
	if err := saveMetadata(p.Suite.Metadata(), ret); err != nil {
		return nil, "", err
	}
	switch p.Suite.Type() {
	case suite.SuiteType:
		ret["test_ids"] = p.Suite.Tests()
	case suite.SuiteSetType:
		ret["suites"] = p.Suite.Suites()
		ret["suite_sets"] = p.Suite.SuiteSets()
	}
	dedupeID := fmt.Sprintf("%s.%s.%s", ret["id"], ret["build_target"], ret["cros_version"])
	return ret, dedupeID, nil
}

func saveMetadata(metadata *suite.Metadata, v map[string]bigquery.Value) error {
	if metadata == nil {
		return fmt.Errorf("expected non-nil metadata")
	}
	v["owners"] = metadata.Owners
	v["criteria"] = metadata.Criteria
	v["bug_component"] = metadata.BugComponent
	return nil
}
