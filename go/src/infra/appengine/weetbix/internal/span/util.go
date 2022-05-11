// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package span

import (
	"bytes"
	"text/template"

	"cloud.google.com/go/spanner"
	"go.chromium.org/luci/common/errors"
)

// GenerateStatement generates a spanner statement from a text template.
func GenerateStatement(tmpl *template.Template, name string, input interface{}) (spanner.Statement, error) {
	sql := &bytes.Buffer{}
	err := tmpl.ExecuteTemplate(sql, name, input)
	if err != nil {
		return spanner.Statement{}, errors.Annotate(err, "failed to generate statement: %s", name).Err()
	}
	return spanner.NewStatement(sql.String()), nil
}
