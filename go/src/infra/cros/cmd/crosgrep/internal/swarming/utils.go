// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package swarming

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/crosgrep/internal/swarming/logging"
)

// BQRow is an alias for the type of a bigquery row.
// We use a map for a bqRow instead of a []bigquery.Value so that
// individual records with a result set retain the name of the column.
type bqRow = []bigquery.Value

// TmplPreamble contains definitions that will be used in SQL templates such as `
// A literal ` cannot appear in a raw string.
const tmplPreamble = "{{$tick := \"`\"}}"

// TemplateOrPanic is a helper function that creates a template or panics.
func templateOrPanic(name string, body string) *template.Template {

	return template.Must(template.New(name).Parse(body))
}

// TemplateToString converts a template and its arguments to a string or fails
// if it cannot.
func templateToString(tmpl *template.Template, input interface{}) (string, error) {
	out := &bytes.Buffer{}
	if err := tmpl.Execute(out, input); err != nil {
		return "", err
	}
	return out.String(), nil
}

// GetRowIterator returns a row iterator from a sql query.
func getRowIterator(ctx context.Context, client *bigquery.Client, sql string) (*bigquery.RowIterator, error) {
	logging.Debugf(ctx, "GetRowIterator %s\n", strings.ReplaceAll(sql, "\n", "\t"))
	q := client.Query(sql)
	it, err := q.Read(ctx)
	return it, errors.Annotate(err, "get iterator").Err()
}

// ToBotName adds a crossk- prefix if one does not already exist.
func toBotName(s string) string {
	if strings.HasPrefix(s, "crossk-") {
		return s
	}
	return fmt.Sprintf("crossk-%s", s)
}
