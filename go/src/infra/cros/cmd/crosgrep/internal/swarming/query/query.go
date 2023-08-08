// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"bytes"
	"context"
	"text/template"

	"cloud.google.com/go/bigquery"

	"infra/cros/cmd/crosgrep/internal/swarming/logging"
)

// MustMakeTemplate takes the name of a template and the body and produces a template.
// In the event of an error it panics. Templates are not intended to be created dynamically.
func mustMakeTemplate(name string, body string) *template.Template {
	return template.Must(template.New(name).Parse("{{$tick := \"`\"}}" + body))
}

// InstantiateSQLQuery takes a template, a normalizer function, and a bundle of parameters and
// creates a SQL query as a string.
func instantiateSQLQuery(ctx context.Context, template *template.Template, params interface{}) (string, error) {
	var out bytes.Buffer
	if err := template.Execute(&out, params); err != nil {
		return "", err
	}
	return out.String(), nil
}

// RunSQL takes a Bigquery Client and a sql query and returns an iterator over
// the result set.
func RunSQL(ctx context.Context, client *bigquery.Client, sql string) (*bigquery.RowIterator, error) {
	logging.Debugf(ctx, "Running SQL query: %s\n", sql)
	query := client.Query(sql)
	it, err := query.Read(ctx)
	return it, err
}

// MustExpandTick takes a string containing {{$tick}} instead of ` and replaces occurrences of
// {{$tick}} with `. In the event that the string is malformed as a text/template, MustExpandTick
// will panic. This function is intended to be used with strings that are compile-time constants.
func mustExpandTick(body string) string {
	template := mustMakeTemplate("", "{{$tick := \"`\"}}"+body)
	out := &bytes.Buffer{}
	if err := template.Execute(out, nil); err != nil {
		panic(err.Error())
	}
	return out.String()
}
