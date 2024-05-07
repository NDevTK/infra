// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bqwrapper

import (
	"context"
	"testing"

	"cloud.google.com/go/bigquery"

	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
)

// TestSimple just tests writing a record and reading it back through UniversalRowQuery.
func TestSimple(t *testing.T) {
	t.Parallel()

	ctx := memory.Use(context.Background())
	datastore.GetTestable(ctx).Consistent(true)

	memBQ, err := MakeMemBQ(ctx)
	if err != nil {
		t.Error(err)
	}

	err = memBQ.Put(ctx, "some-project", "some-dataset", "some-table", []bigquery.ValueSaver{
		&testValueSaver{
			data: map[string]any{
				"key":   400,
				"value": false,
			},
		},
	})
	if err != nil {
		t.Error(err)
	}

	var records []*EmulatedBigqueryRecord
	if err := datastore.GetAll(ctx, memBQ.UniversalRowQuery(ctx), &records); err != nil {
		t.Error(err)
	}

	if len(records) != 1 {
		t.Errorf("records has unexpected length %d", len(records))
	}

	vals := records[0].Extra["key"].Slice()
	val := vals[0].String()
	if val != "PTInt(400)" {
		t.Errorf("unexpected value for key from only record: %s", val)
	}
}

type testValueSaver struct {
	data map[string]any
}

func (saver *testValueSaver) Save() (map[string]bigquery.Value, string, error) {
	out := map[string]bigquery.Value{}
	for k, v := range saver.data {
		out[k] = v
	}
	return out, "", nil
}
