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
)

func TestFetchURLs(t *testing.T) {
	Convey("Test fetch urls", t, func() {
		ctx := context.Background()
		plats := Platforms{}

		g := &FetchURLs{
			Metadata: &core.Action_Metadata{Name: "urls"},
			URLs: map[string]FetchURL{
				"something1": {
					URL:  "https://host/path1",
					Mode: 0o777,
				},
				"dir1/something2": {
					URL:           "https://host/path2",
					HashAlgorithm: core.HashAlgorithm_HASH_MD5,
					HashValue:     "abcdef",
				},
			},
		}
		a, err := g.Generate(ctx, plats)
		So(err, ShouldBeNil)

		url := testutils.Assert[*core.Action_Copy](t, a.Spec)
		So(url.Copy.Files, ShouldResemble, map[string]*core.ActionFilesCopy_Source{
			"something1": {
				Content: &core.ActionFilesCopy_Source_Output_{
					Output: &core.ActionFilesCopy_Source_Output{Name: "urls_2o025r0794", Path: "file"},
				},
				Mode: 0o777,
			},
			"dir1/something2": {
				Content: &core.ActionFilesCopy_Source_Output_{
					Output: &core.ActionFilesCopy_Source_Output{Name: "urls_om04u163h4", Path: "file"},
				},
				Mode: 0o666,
			},
		})

		{
			So(a.Deps, ShouldHaveLength, 3)
			for _, d := range a.Deps[1:] {
				u := testutils.Assert[*core.Action_Url](t, d.Action.Spec)
				switch d.Action.Metadata.Name {
				case "urls_2o025r0794":
					So(u.Url, testutils.ShouldEqualProto, &core.ActionURLFetch{
						Url: "https://host/path1",
					})
				case "urls_om04u163h4":
					So(u.Url, testutils.ShouldEqualProto, &core.ActionURLFetch{
						Url:           "https://host/path2",
						HashAlgorithm: core.HashAlgorithm_HASH_MD5,
						HashValue:     "abcdef",
					})
				}
			}
		}
	})
}
