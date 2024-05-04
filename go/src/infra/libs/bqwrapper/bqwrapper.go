// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package bqwrapper is a wrapper around the bigquery API.
//
// It does some things a little differently. The real bigquery go library uses a fluent interface.
// This library does not.
//
// It does not expose very much (only what is necessary for the projects that are using it).
// If you need more functionality, please send a CL.
//
// Anyway, the wrapper interface has two implementations:
//
// 1) CloudBQ -- the production one.
// 2) MemBQ   -- the testing one, which is implemented on top of the datastore emulator that we already have.
//
// I have made some attempt to mimic the semantics of the real BigQuery, but I didn't try very hard.
// This package is mostly there to allow unit code that uses bigquery to be unit-tested in a reasonable way.
package bqwrapper

import (
	"context"

	"cloud.google.com/go/bigquery"
)

// BQIf is the bigquery wrapper interface.
//
// It exposes the subset of the bigquery interface that can be used while testing and in prod.
type BQIf interface {
	// Put inserts some records into Bigquery.
	//
	// Note that you always specify the projectID, dataset, and table.
	//
	// The Save method on each element of data MUST expose datastore-compatible types.
	//
	// The exact rules for how a go type gets converted to datastore by this library will
	// probably change in the future, depending on what exactly we try to store in BigQuery
	// in practice.
	Put(ctx context.Context, projectID string, dataset string, table string, data []bigquery.ValueSaver) error
}
