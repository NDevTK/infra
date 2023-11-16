// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package wheels

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/vpython/api/vpython"
)

func TestGeneratingEnsureFile(t *testing.T) {
	Convey("Test generate ensure file", t, func() {
		ef, err := ensureFileFromVPythonSpec(&vpython.Spec{
			Wheel: []*vpython.Spec_Package{
				{Name: "pkg1", Version: "version1"},
				{Name: "pkg2", Version: "version2"},
			},
		}, nil)
		So(err, ShouldBeNil)
		So(ef.PackagesBySubdir["wheels"], ShouldResemble, ensure.PackageSlice{
			{PackageTemplate: "pkg1", UnresolvedVersion: "version1"},
			{PackageTemplate: "pkg2", UnresolvedVersion: "version2"},
		})

	})
	Convey("Test duplicated wheels", t, func() {
		Convey("Same version", func() {
			ef, err := ensureFileFromVPythonSpec(&vpython.Spec{
				Wheel: []*vpython.Spec_Package{
					{Name: "pkg1", Version: "version1"},
					{Name: "pkg1", Version: "version1"},
				},
			}, nil)
			So(err, ShouldBeNil)
			So(ef.PackagesBySubdir["wheels"], ShouldResemble, ensure.PackageSlice{
				{PackageTemplate: "pkg1", UnresolvedVersion: "version1"},
			})
		})
		Convey("Different version", func() {
			_, err := ensureFileFromVPythonSpec(&vpython.Spec{
				Wheel: []*vpython.Spec_Package{
					{Name: "pkg1", Version: "version1"},
					{Name: "pkg1", Version: "version2"},
				},
			}, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "multiple versions for package")
		})
	})
}
