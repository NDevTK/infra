// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cipd

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/errors"
	. "go.chromium.org/luci/common/testing/assertions"
)

type fakeClient struct {
	ensure func(ctx context.Context, serviceUrl, cipdRoot string, packages map[string]*Package) (map[string]string, error)
}

func (f *fakeClient) Ensure(ctx context.Context, serviceUrl, cipdRoot string, packages map[string]*Package) (map[string]string, error) {
	ensure := f.ensure
	if ensure != nil {
		return ensure(ctx, serviceUrl, cipdRoot, packages)
	}
	return nil, nil
}

func TestEnsure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Ensure", t, func() {

		cipdRoot := t.TempDir()

		Convey("fails if provided empty service URL", func() {
			resolvedPackages, err := Ensure(ctx, "", cipdRoot, map[string]*Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-version",
				},
			})

			So(err, ShouldErrLike, "empty serviceUrl")
			So(resolvedPackages, ShouldBeNil)
		})

		Convey("fails if provided empty CIPD root", func() {
			resolvedPackages, err := Ensure(ctx, "fake-url", "", map[string]*Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-version",
				},
			})

			So(err, ShouldErrLike, "empty cipdRoot")
			So(resolvedPackages, ShouldBeNil)
		})

		Convey("fails if provided empty packages", func() {
			resolvedPackages, err := Ensure(ctx, "fake-url", cipdRoot, nil)

			So(err, ShouldErrLike, "empty packages")
			So(resolvedPackages, ShouldBeNil)
		})

		Convey("fails if provided empty subdir", func() {
			resolvedPackages, err := Ensure(ctx, "fake-url", cipdRoot, map[string]*Package{
				"": {
					Name:    "fake-package",
					Version: "fake-version",
				},
			})

			So(err, ShouldErrLike, "empty subdir in packages")
			So(resolvedPackages, ShouldBeNil)
		})

		Convey("fails if provided nil package", func() {
			resolvedPackages, err := Ensure(ctx, "fake-url", cipdRoot, map[string]*Package{
				"fake-subdir": nil,
			})

			So(err, ShouldErrLike, `nil package for subdir "fake-subdir"`)
			So(resolvedPackages, ShouldBeNil)
		})

		Convey("fails if provided empty package name", func() {
			resolvedPackages, err := Ensure(ctx, "fake-url", cipdRoot, map[string]*Package{
				"fake-subdir": {
					Name:    "",
					Version: "fake-version",
				},
			})

			So(err, ShouldErrLike, `empty package name for subdir "fake-subdir"`)
			So(resolvedPackages, ShouldBeNil)
		})

		Convey("fails if provided empty package version", func() {
			resolvedPackages, err := Ensure(ctx, "fake-url", cipdRoot, map[string]*Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "",
				},
			})

			So(err, ShouldErrLike, `empty package version for subdir "fake-subdir"`)
			So(resolvedPackages, ShouldBeNil)
		})

		Convey("fails if ensuring packages fails", func() {
			factory := func(ctx context.Context) Client {
				return &fakeClient{ensure: func(ctx context.Context, serviceUrl, cipdRoot string, packages map[string]*Package) (map[string]string, error) {
					return nil, errors.New("test Ensure failure")
				}}
			}
			ctx := UseClientFactory(ctx, factory)

			resolvedPackages, err := Ensure(ctx, "fake-url", cipdRoot, map[string]*Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-version",
				},
			})

			So(err, ShouldErrLike, "test Ensure failure")
			So(resolvedPackages, ShouldBeNil)
		})

		Convey("returns resolved package information on success", func() {
			factory := func(ctx context.Context) Client {
				return &fakeClient{ensure: func(ctx context.Context, serviceUrl, cipdRoot string, packages map[string]*Package) (map[string]string, error) {
					return map[string]string{"fake-subdir": "fake-instance-id"}, nil
				}}
			}
			ctx := UseClientFactory(ctx, factory)

			resolvedPackages, err := Ensure(ctx, "fake-url", cipdRoot, map[string]*Package{
				"fake-subdir": {
					Name:    "fake-package",
					Version: "fake-version",
				},
			})

			So(err, ShouldBeNil)
			So(resolvedPackages, ShouldResemble, map[string]*ResolvedPackage{
				"fake-subdir": {
					Name:             "fake-package",
					RequestedVersion: "fake-version",
					ActualVersion:    "fake-instance-id",
				},
			})
		})

	})
}

func TestUnmarshalEnsureJsonOut(t *testing.T) {

	Convey("unmarshallEnsureJsonOut decodes valid ensure json out", t, func() {
		jsonOutContents := []byte(`{
			"result": {
				"exe": [
					{
						"package": "fake-package",
						"instance_id": "fake-instance-id"
					}
				]
			}
		}`)

		out, err := unmarshalEnsureJsonOut(jsonOutContents)

		So(err, ShouldBeNil)
		So(out, ShouldResemble, &jsonOut{
			Result: map[string][]jsonPackage{
				"exe": {
					{
						Package:    "fake-package",
						InstanceId: "fake-instance-id",
					},
				},
			},
		})

	})
}
