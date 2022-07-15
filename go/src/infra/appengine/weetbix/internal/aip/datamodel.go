// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package aip contains utilities used to comply with API Improvement
// Proposals (AIPs) from https://google.aip.dev/. This includes
// an AIP-160 filter parser and SQL generator and AIP-132 order by
// clause parser and SQL generator.
package aip

import (
	"fmt"
	"strings"
)

// Column represents the schema of a Database column.
type Column struct {
	// The externally-visible name of the column. This may be used in AIP-160
	// filters and order by clauses.
	name string

	// The database name of the column.
	// Important: Only assign assign safe constants to this field.
	// User input MUST NOT flow to this field, as it will be used directly
	// in SQL statements and would allow the user to perform SQL injection
	// attacks.
	databaseName string

	// Whether this column can be sorted on.
	sortable bool

	// Whether this column can be filtered on.
	filterable bool

	// ImplicitFilter controls whether this field is searched implicitly
	// in AIP-160 filter expressions.
	implicitFilter bool
}

// Table represents the schema of a Database table, view or query.
type Table struct {
	// The columns in the database table.
	columns []*Column

	// A mapping from externally-visible column name to the column
	// definition. The column name used as a key is in lowercase.
	columnByName map[string]*Column
}

// FilterableColumnByName returns the database name of the filterable column
// with the given externally-visible name.
func (t *Table) FilterableColumnByName(name string) (*Column, error) {
	col := t.columnByName[strings.ToLower(name)]
	if col != nil && col.filterable {
		return col, nil
	}

	columnNames := []string{}
	for _, column := range t.columns {
		if column.filterable {
			columnNames = append(columnNames, column.name)
		}
	}
	return nil, fmt.Errorf("no filterable field named %q, valid fields are %s", name, strings.Join(columnNames, ", "))
}

// SortableColumnByName returns the sortable database column
// with the given externally-visible name.
func (t *Table) SortableColumnByName(name string) (*Column, error) {
	col := t.columnByName[strings.ToLower(name)]
	if col != nil && col.sortable {
		return col, nil
	}

	columnNames := []string{}
	for _, column := range t.columns {
		if column.sortable {
			columnNames = append(columnNames, column.name)
		}
	}
	return nil, fmt.Errorf("no sortable field named %q, valid fields are %s", name, strings.Join(columnNames, ", "))
}
