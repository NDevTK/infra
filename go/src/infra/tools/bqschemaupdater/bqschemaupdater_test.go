// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	pb "infra/libs/bqschema/tabledef"
)

func TestTableDef(t *testing.T) {
	want := &pb.TableDef{
		Dataset: pb.TableDef_AGGREGATED,
		TableId: "test_table",
		Fields: []*pb.FieldSchema{
			{
				Name:        "field1",
				Type:        pb.Type_RECORD,
				Description: "test field",
				Schema: []*pb.FieldSchema{
					{
						Name:        "nested",
						Type:        pb.Type_STRING,
						Description: "nested",
					},
				},
			},
			{
				Name:        "field2",
				Type:        pb.Type_INTEGER,
				Description: "test field 2",
			},
		},
	}
	buf, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got := tableDef(bytes.NewReader(buf))
	if !(reflect.DeepEqual(got, want)) {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func TestUpdateFromTableDef(t *testing.T) {
	ctx := context.Background()
	ts := localTableStore{}
	datasetID := pb.TableDef_AGGREGATED.ID()
	tableID := "test_table"

	field := &pb.FieldSchema{
		Name:        "test_field",
		Description: "test description",
		Type:        pb.Type_STRING,
	}
	anotherField := &pb.FieldSchema{
		Name:        "field_2",
		Description: "another field",
		Type:        pb.Type_STRING,
	}
	tcs := [][]*pb.FieldSchema{
		{field},
		{field, anotherField},
	}
	for _, tc := range tcs {
		td := &pb.TableDef{
			Dataset: pb.TableDef_AGGREGATED,
			TableId: tableID,
			Fields:  tc,
		}
		err := updateFromTableDef(ctx, ts, td)
		if err != nil {
			t.Fatal(err)
		}
		got, err := ts.getTableMetadata(ctx, datasetID, tableID)
		if err != nil {
			t.Fatal(err)
		}
		want := &bigquery.TableMetadata{Schema: pb.BQSchema(tc)}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got: %v; want: %v", got, want)
		}
	}
}
