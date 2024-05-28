// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package database

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"go.chromium.org/luci/common/logging"
)

// BuildQueryFilter builds a WHERE clause based on a filter string.
//
// Parse filter from string. Assume filters are separated by "AND" based on
// AIP-160.
func BuildQueryFilter(ctx context.Context, filter string) (string, []interface{}) {
	var (
		filterExprs []string
		filterArgs  []interface{}
		queryFilter string
	)

	logging.Debugf(ctx, "BuildQueryFilter: processing filter string %s", filter)
	for _, filterPart := range strings.Split(filter, "AND") {
		trimmedFilter := strings.TrimSpace(filterPart)
		parts := strings.FieldsFunc(trimmedFilter, func(r rune) bool {
			return r == '=' || r == '!' || r == '>' || r == '<' || unicode.IsSpace(r)
		})

		if len(parts) != 2 {
			logging.Warningf(ctx, "BuildQueryFilter: invalid filter format: %s", trimmedFilter)
			break
		}

		field := parts[0]
		value := parts[1]
		operator := strings.TrimSpace(trimmedFilter[len(field):(len(trimmedFilter) - len(value))]) // extract the operator
		logging.Debugf(ctx, "BuildQueryFilter: processing %s %s %s", field, operator, value)

		position := fmt.Sprintf("$%d", len(filterArgs)+1)
		filterArgs = append(filterArgs, value)

		switch operator {
		case "=":
			filterExprs = append(filterExprs, fmt.Sprintf("%s = %s", field, position))
		case "!=":
			filterExprs = append(filterExprs, fmt.Sprintf("%s != %s", field, position))
		case ">":
			filterExprs = append(filterExprs, fmt.Sprintf("%s > %s", field, position))
		case "<":
			filterExprs = append(filterExprs, fmt.Sprintf("%s < %s", field, position))
		case ">=":
			filterExprs = append(filterExprs, fmt.Sprintf("%s >= %s", field, position))
		case "<=":
			filterExprs = append(filterExprs, fmt.Sprintf("%s <= %s", field, position))
		default:
			logging.Warningf(ctx, "BuildQueryFilter: unsupported filter: %s %s %s", field, operator, value)
			continue
		}
	}

	if len(filterExprs) > 0 {
		queryFilter += `
		WHERE ` + strings.Join(filterExprs, " AND ")
	}
	return queryFilter, filterArgs
}
