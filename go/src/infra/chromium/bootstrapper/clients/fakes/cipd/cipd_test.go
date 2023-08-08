// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cipd

import (
	"context"
	real "infra/chromium/bootstrapper/clients/cipd"
	"infra/chromium/util"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/common/testing/testfs"
)

func collect(cipdRoot, subdir string) map[string]string {
	layout, err := testfs.Collect(filepath.Join(cipdRoot, subdir))
	util.PanicOnError(err)
	return layout
}

func TestEnsure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Client.Ensure", t, func() {
		cipdRoot := t.TempDir()

		Convey("returns pin for a package by default", func() {
			client := Client{}

			packageVersions, err := client.Ensure(ctx, "fake-url", cipdRoot, map[string]*real.Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-version",
				},
			})

			So(err, ShouldBeNil)
			So(packageVersions, ShouldContainKey, "fake-subdir")
			So(packageVersions["fake-subdir"], ShouldNotBeEmpty)
		})

		Convey("fails for a nil package", func() {
			client := Client{map[string]*Package{
				"fake-package": nil,
			}}

			packageVersions, err := client.Ensure(ctx, "fake-url", cipdRoot, map[string]*real.Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-version",
				},
			})

			So(err, ShouldErrLike, `unknown package "fake-package"`)
			So(packageVersions, ShouldBeNil)
		})

		Convey("fails for an a version mapping to an empty instance ID", func() {
			client := Client{map[string]*Package{
				"fake-package": {
					Refs: map[string]string{
						"fake-version": "",
					},
				},
			}}

			packageVersions, err := client.Ensure(ctx, "fake-url", cipdRoot, map[string]*real.Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-version",
				},
			})

			So(err, ShouldErrLike, `unknown version "fake-version" of package "fake-package"`)
			So(packageVersions, ShouldBeNil)
		})

		Convey("returns pin for version mapping to provided instance ID", func() {
			client := Client{map[string]*Package{
				"fake-package": {
					Refs: map[string]string{
						"fake-version": "fake-instance-id",
					},
				},
			}}

			packageVersions, err := client.Ensure(ctx, "fake-url", cipdRoot, map[string]*real.Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-version",
				},
			})

			So(err, ShouldBeNil)
			So(packageVersions, ShouldContainKey, "fake-subdir")
			So(packageVersions["fake-subdir"], ShouldEqual, "fake-instance-id")
		})

		Convey("fails for a non-existent instance ID", func() {
			client := Client{map[string]*Package{
				"fake-package": {
					Instances: map[string]*PackageInstance{
						"fake-instance-id": nil,
					},
				},
			}}

			packageVersions, err := client.Ensure(ctx, "fake-url", cipdRoot, map[string]*real.Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-instance-id",
				},
			})

			So(err, ShouldErrLike, `unknown version "fake-instance-id" of package "fake-package"`)
			So(packageVersions, ShouldBeNil)
		})

		Convey("returns pin for instance ID", func() {
			client := Client{map[string]*Package{
				"fake-package": {
					Instances: map[string]*PackageInstance{
						"fake-instance-id": {},
					},
				},
			}}

			packageVersions, err := client.Ensure(ctx, "fake-url", cipdRoot, map[string]*real.Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-instance-id",
				},
			})

			So(err, ShouldBeNil)
			So(packageVersions, ShouldContainKey, "fake-subdir")
			So(packageVersions["fake-subdir"], ShouldEqual, "fake-instance-id")
		})

		Convey("deploys specified files", func() {
			client := Client{map[string]*Package{
				"fake-package": {
					Instances: map[string]*PackageInstance{
						"fake-instance-id": {
							Contents: map[string]string{
								"infra/config/recipes.cfg": "fake-recipes.cfg",
								"recipes/foo.py":           "fake-recipe-foo",
							},
						},
					},
				},
			}}

			_, err := client.Ensure(ctx, "fake-url", cipdRoot, map[string]*real.Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-instance-id",
				},
			})

			So(err, ShouldBeNil)
			layout := collect(cipdRoot, "fake-subdir")
			So(layout, ShouldResemble, map[string]string{
				"infra/config/recipes.cfg": "fake-recipes.cfg",
				"recipes/foo.py":           "fake-recipe-foo",
			})
		})

	})
}

func TestIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("CIPD using fake factory", t, func() {

		cipdRoot := t.TempDir()

		ctx := real.UseClientFactory(ctx, Factory(nil))

		Convey("succeeds when calling EnsurePackages", func() {
			packages, err := real.Ensure(ctx, "fake-url", cipdRoot, map[string]*real.Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-version",
				},
			})

			So(err, ShouldBeNil)
			So(packages, ShouldContainKey, "fake-subdir")
			pkg := packages["fake-subdir"]
			So(pkg.Name, ShouldEqual, "fake-package")
			So(pkg.RequestedVersion, ShouldEqual, "fake-version")
			So(pkg.ActualVersion, ShouldNotBeEmpty)
		})

	})
}
