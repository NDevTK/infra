// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tabledef

import (
	"cloud.google.com/go/bigquery"
)

// BQSchema constructs a bigquery.Schema from a []*TableDef.FieldSchema
func BQSchema(fields []*FieldSchema) bigquery.Schema {
	var s bigquery.Schema
	for _, f := range fields {
		s = append(s, bqField(f))
	}
	return s
}

func bqField(f *FieldSchema) *bigquery.FieldSchema {
	fs := &bigquery.FieldSchema{
		Name:        f.Name,
		Description: f.Description,
		Type:        bigquery.FieldType(f.Type.String()),
		Repeated:    f.IsRepeated,
		Required:    f.IsRequired,
	}
	if fs.Type == bigquery.RecordFieldType {
		fs.Schema = BQSchema(f.Schema)
	}
	return fs
}
