// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"infra/libs/cipkg_new/core"
	"infra/libs/cipkg_new/testutils"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessURL(t *testing.T) {
	Convey("Test action processor for url", t, func() {
		ap := NewActionProcessor("", testutils.NewMockPackageManage(""))

		url := &core.ActionURLFetch{
			Url:           "https://host.not.exist/123",
			HashAlgorithm: core.HashAlgorithm_HASH_SHA256,
			HashValue:     "abcdef",
		}

		pkg, err := ap.Process(&core.Action{
			Metadata: &core.Action_Metadata{Name: "url"},
			Deps:     []*core.Action_Dependency{ReexecDependency()},
			Spec:     &core.Action_Url{Url: url},
		})
		So(err, ShouldBeNil)

		So(pkg.Dependencies, ShouldHaveLength, 1)
		So(pkg.Derivation.Args[0], ShouldStartWith, pkg.Dependencies[0].Package.Handler.OutputDirectory())
		checkReexecArg(pkg.Derivation.Args, url)
	})
}

func TestExecuteURL(t *testing.T) {
	Convey("Test execute action url", t, func() {
		ctx := context.Background()
		dst := testutils.NewAferoMemMapFs()

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "something")
		}))
		defer s.Close()

		Convey("Test download file", func() {
			a := &core.ActionURLFetch{
				Url: s.URL,
			}

			err := ActionURLFetchExecutor(ctx, a, dst)
			So(err, ShouldBeNil)

			{
				f, err := dst.Open("file")
				So(err, ShouldBeNil)
				b, err := io.ReadAll(f)
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "something")
			}
		})

		Convey("Test download file with hash verify", func() {
			a := &core.ActionURLFetch{
				Url:           s.URL,
				HashAlgorithm: core.HashAlgorithm_HASH_SHA256,
				HashValue:     "3fc9b689459d738f8c88a3a48aa9e33542016b7a4052e001aaa536fca74813cb",
			}

			err := ActionURLFetchExecutor(ctx, a, dst)
			So(err, ShouldBeNil)

			{
				f, err := dst.Open("file")
				So(err, ShouldBeNil)
				b, err := io.ReadAll(f)
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "something")
			}
		})

		Convey("Test download file with hash verify failed", func() {
			a := &core.ActionURLFetch{
				Url:           s.URL,
				HashAlgorithm: core.HashAlgorithm_HASH_SHA256,
				HashValue:     "abcdef",
			}

			err := ActionURLFetchExecutor(ctx, a, dst)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "hash mismatch")
		})
	})
}

func TestReexecURL(t *testing.T) {
	Convey("Test re-execute action processor for url", t, func() {
		ap := NewActionProcessor("", testutils.NewMockPackageManage(""))

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "something")
		}))
		defer s.Close()

		pkg, err := ap.Process(&core.Action{
			Metadata: &core.Action_Metadata{Name: "url"},
			Deps:     []*core.Action_Dependency{ReexecDependency()},
			Spec: &core.Action_Url{Url: &core.ActionURLFetch{
				Url:           s.URL,
				HashAlgorithm: core.HashAlgorithm_HASH_SHA256,
				HashValue:     "3fc9b689459d738f8c88a3a48aa9e33542016b7a4052e001aaa536fca74813cb",
			}},
		})
		So(err, ShouldBeNil)

		dst := testutils.NewAferoMemMapFs()
		runWithDrv(dst, pkg.Derivation)

		{
			f, err := dst.Open("out/file")
			So(err, ShouldBeNil)
			b, err := io.ReadAll(f)
			So(err, ShouldBeNil)
			So(string(b), ShouldEqual, "something")
		}
	})
}
