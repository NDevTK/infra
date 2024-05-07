// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bqwrapper

import (
	"context"
	"errors"

	"cloud.google.com/go/bigquery"

	"go.chromium.org/luci/gae/service/datastore"
)

type MemBQ struct{}

var _ BQIf = &MemBQ{}

// MakeMemBQ makes a new MemBQ.
func MakeMemBQ(ctx context.Context) (*MemBQ, error) {
	if datastore.GetTestable(ctx) == nil {
		return nil, errors.New("MemBQ must be created with a testable datastore in the context")
	}
	return &MemBQ{}, nil
}

// Put writes a record to the fake datastore and thus the fake BigQuery.
func (mbq *MemBQ) Put(ctx context.Context, projectID string, dataset string, table string, data []bigquery.ValueSaver) error {
	mustBeTestable(ctx)

	var entities []*EmulatedBigqueryRecord
	for _, valueSaver := range data {
		row, _, err := valueSaver.Save()
		if err != nil {
			return err
		}
		entities = append(entities, &EmulatedBigqueryRecord{
			ProjectID: projectID,
			Dataset:   dataset,
			Table:     table,
			Extra:     rowToPropertyMap(row),
		})
	}

	return datastore.Put(ctx, entities)
}

// UniversalRowQuery gives a query over the row records.
func (mbq *MemBQ) UniversalRowQuery(ctx context.Context) *datastore.Query {
	mustBeTestable(ctx)
	return datastore.NewQuery(emulatedBigqueryRecordKind)
}

// Private function mustBeTestable checks that we are, in fact, in an environment with a testable datastore.
//
// Trying to use the prod datastore as a poorly emulated BigQuery would be bad.
func mustBeTestable(ctx context.Context) {
	if datastore.GetTestable(ctx) == nil {
		// It would be kind of cool if you could indentionally use this thing with the production datastore.
		// I'm not sure it would be useful, but it would be cool.
		panic("MemBQ is for testing contexts only! It cannot be used with the production datastore.")
	}
}

const emulatedBigqueryRecordKind = "emulatedBigqueryRecordKind"

// EmulatedBigqueryRecord is a record in datastore that emulates a bigquery row.
//
// Fields used internally by the emulator are prefixed with FOUR underscores. ____
type EmulatedBigqueryRecord struct {
	_kind string `gae:"$kind,emulatedBigqueryRecordKind"`
	// Make this public so unit tests can consume them.
	ProjectID string `gae:"____project_id"`
	Dataset   string `gae:"____dataset"`
	Table     string `gae:"____table"`
	// Extra *has* to be exported or the datastore ORM will crash.
	Extra datastore.PropertyMap `gae:",extra"`
}

// Appease staticcheck.
var _ = (EmulatedBigqueryRecord{})._kind

func rowToPropertyMap(row map[string]bigquery.Value) datastore.PropertyMap {
	out := make(datastore.PropertyMap, len(row))
	for k, v := range row {
		out[k] = datastore.MkProperty(v)
	}
	return out
}
