// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package aip

import "strings"

type ColumnBuilder struct {
	column Column
}

// NewColumn starts building a new column.
func NewColumn() *ColumnBuilder {
	return &ColumnBuilder{}
}

// WithName specifies the user-visible name of the column.
func (c *ColumnBuilder) WithName(name string) *ColumnBuilder {
	c.column.name = name
	return c
}

// WithDatabaseName specifies the database name of the column.
// Important: Only pass safe values (e.g. compile-time constants) to this
// field.
// User input MUST NOT flow to this field, as it will be used directly
// in SQL statements and would allow the user to perform SQL injection
// attacks.
func (c *ColumnBuilder) WithDatabaseName(name string) *ColumnBuilder {
	c.column.databaseName = name
	return c
}

// Sortable specifies this column can be sorted on.
func (c *ColumnBuilder) Sortable() *ColumnBuilder {
	c.column.sortable = true
	return c
}

// Filterable specifies this column can be filtered on.
func (c *ColumnBuilder) Filterable() *ColumnBuilder {
	c.column.filterable = true
	return c
}

// FilterableImplicitly specifies this column can be filtered on implicitly.
// This means that AIP-160 filter expressions not referencing any
// particular field will try to search in this column.
func (c *ColumnBuilder) FilterableImplicitly() *ColumnBuilder {
	c.column.filterable = true
	c.column.implicitFilter = true
	return c
}

// Build returns the built column.
func (c *ColumnBuilder) Build() *Column {
	result := &Column{}
	*result = c.column
	return result
}

type TableBuilder struct {
	columns []*Column
}

// NewTable starts building a new table.
func NewTable() *TableBuilder {
	return &TableBuilder{}
}

// WithColumns specifies the columns in the table.
func (t *TableBuilder) WithColumns(columns ...*Column) *TableBuilder {
	t.columns = columns
	return t
}

// Build returns the built table.
func (t *TableBuilder) Build() *Table {
	columnByName := make(map[string]*Column)
	for _, c := range t.columns {
		lowerName := strings.ToLower(c.name)
		if _, ok := columnByName[lowerName]; ok {
			panic("multiple columns with the same name: " + lowerName)
		}
		columnByName[strings.ToLower(c.name)] = c
	}

	return &Table{
		columns:      t.columns,
		columnByName: columnByName,
	}
}
