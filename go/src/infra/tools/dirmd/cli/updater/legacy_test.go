// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package updater

import (
	"bytes"
	"encoding/json"
	"testing"

	"infra/tools/dirmd"
	dirmdpb "infra/tools/dirmd/proto"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLegacy(t *testing.T) {
	t.Parallel()

	Convey(`Legacy`, t, func() {
		m, err := dirmd.ReadMapping("testdata/root", dirmdpb.MappingForm_FULL)
		So(err, ShouldBeNil)
		actual := toLegacyFormat(m)
		So(jsonIndent(actual), ShouldEqual, jsonIndent([]byte(`{
			"AAA-README": [
				"",
				"This file is generated by infra.git/go/src/tools/dirmd/cli/updater",
				"by parsing DIR_METADATA files throughout the chromium source code.",
				"",
				"Manual edits of this file will be overwritten by an automated process."
			],
			"component-to-team":  {},
			"dir-to-component": {
				"subdir": "Some\u003eComponent(Linux)",
				"subdir/empty_subdir": "Some\u003eComponent(Linux)"
			},
			"dir-to-team": {
				".": "chromium-review@chromium.org",
				"subdir": "team-email@chromium.org",
				"subdir/empty_subdir": "team-email@chromium.org"
			},
			"teams-per-component": {}
		}`)))
	})
}

func jsonIndent(data []byte) string {
	buf := &bytes.Buffer{}
	err := json.Indent(buf, data, "", "  ")
	So(err, ShouldBeNil)
	return string(buf.Bytes())
}
