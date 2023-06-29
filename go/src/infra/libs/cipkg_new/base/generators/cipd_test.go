// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package generators

import (
	"context"
	"testing"

	"infra/libs/cipkg_new/core"
	"infra/libs/cipkg_new/testutils"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/cipd/client/cipd/ensure"
)

func TestCIPDExport(t *testing.T) {
	Convey("Test cipd export", t, func() {
		ctx := context.Background()
		plats := Platforms{}

		g := &CIPDExport{
			Ensure: ensure.File{
				PackagesBySubdir: map[string]ensure.PackageSlice{
					"": {
						{PackageTemplate: "infra/3pp/tools/git", UnresolvedVersion: "version:2@2.36.1.chromium.8"},
					},
				},
			},
			ServiceURL: "http://something",
		}
		a, err := g.Generate(ctx, plats)
		So(err, ShouldBeNil)

		cipd := testutils.Assert[*core.Action_Cipd](t, a.Spec)
		So(cipd.Cipd.EnsureFile, ShouldEqual, "infra/3pp/tools/git  version:2@2.36.1.chromium.8\n")
		So(cipd.Cipd.Env, ShouldResemble, []string{
			"CIPD_SERVICE_URL=http://something",
		})
	})
}
