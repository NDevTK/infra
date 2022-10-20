// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package aip

import (
	"fmt"
	"strings"
)

// MergeWithDefaultOrder merges the specified order with the given
// defaultOrder. The merge occurs as follows:
//   - Ordering specified in `order` takes precedence.
//   - For columns not specified in the `order` that appear in `defaultOrder`,
//     ordering is applied in the order they apply in defaultOrder.
func MergeWithDefaultOrder(defaultOrder []OrderBy, order []OrderBy) []OrderBy {
	result := make([]OrderBy, 0, len(order)+len(defaultOrder))
	seenColumns := make(map[string]struct{})
	for _, o := range order {
		result = append(result, o)
		seenColumns[strings.ToLower(o.Name)] = struct{}{}
	}
	for _, o := range defaultOrder {
		if _, ok := seenColumns[strings.ToLower(o.Name)]; !ok {
			result = append(result, o)
		}
	}
	return result
}

// OrderByClause returns a Standard SQL Order by clause, including
// "ORDER BY" and trailing new line (if an order is specified).
// If no order is specified, returns "".
//
// The returned order clause is safe against SQL injection; only
// strings appearing from Table appear in the output.
func (t *Table) OrderByClause(order []OrderBy) (string, error) {
	if len(order) == 0 {
		return "", nil
	}
	seenColumns := make(map[string]struct{})
	var result strings.Builder
	result.WriteString("ORDER BY ")
	for i, o := range order {
		if i > 0 {
			result.WriteString(", ")
		}
		column, err := t.SortableColumnByName(o.Name)
		if err != nil {
			return "", err
		}
		if _, ok := seenColumns[column.databaseName]; ok {
			return "", fmt.Errorf("field appears in order_by multiple times: %q", o.Name)
		}
		seenColumns[column.databaseName] = struct{}{}
		result.WriteString(column.databaseName)
		if o.Descending {
			result.WriteString(" DESC")
		}
	}
	result.WriteString("\n")
	return result.String(), nil
}
