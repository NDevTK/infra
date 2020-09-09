// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"infra/appengine/cr-rev/config"
	"infra/appengine/cr-rev/models"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/gae/service/datastore"
)

func redirectTestSetup() context.Context {
	ctx := gaetesting.TestingContext()
	ds := datastore.GetTestable(ctx)
	ds.Consistent(true)
	ds.AutoIndex(true)
	return ctx
}

func TestRedirects(t *testing.T) {
	r := newRedirectRules()
	Convey("number redirect", t, func() {
		ctx := redirectTestSetup()
		commits := []*models.Commit{
			{
				ID:             "chromium-chromium/src-0000000000000000000000000000000000000001",
				CommitHash:     "0000000000000000000000000000000000000001",
				Host:           "chromium",
				Repository:     "chromium/src",
				PositionNumber: 1,
				PositionRef:    "svn://svn.chromium.org/chrome",
			},
			{
				ID:             "chromium-chromium/src-0000000000000000000000000000000000000002",
				CommitHash:     "0000000000000000000000000000000000000002",
				Host:           "chromium",
				Repository:     "chromium/src",
				PositionNumber: 2,
				PositionRef:    "svn://svn.chromium.org/chrome/trunk",
			},
			{
				ID:             "chromium-chromium/src-0000000000000000000000000000000000000022",
				CommitHash:     "0000000000000000000000000000000000000022",
				Host:           "chromium",
				Repository:     "chromium/src",
				PositionNumber: 2,
				PositionRef:    "svn://svn.chromium.org/blink",
			},
			{
				ID:             "chromium-chromium/src-0000000000000000000000000000000000000222",
				CommitHash:     "0000000000000000000000000000000000000222",
				Host:           "chromium",
				Repository:     "foo",
				PositionNumber: 2,
				PositionRef:    "refs/heads/master",
			},
			{
				ID:             "chromium-chromium/src-0000000000000000000000000000000000000003",
				CommitHash:     "0000000000000000000000000000000000000003",
				Host:           "chromium",
				Repository:     "chromium/src",
				PositionNumber: 3,
				PositionRef:    "svn://svn.chromium.org/blink",
			},
			{
				ID:             "chromium-chromium/src-0000000000000000000000000000000000288197",
				CommitHash:     "0000000000000000000000000000000000288197",
				Host:           "chromium",
				Repository:     "chromium/src",
				PositionNumber: 288197,
				PositionRef:    "svn://svn.chromium.org/chrome/trunk/src",
			},
			{
				ID:             "chromium-chromium/src-0000000000000000000000000000000000291560",
				CommitHash:     "0000000000000000000000000000000000291560",
				Host:           "chromium",
				Repository:     "chromium/src",
				PositionNumber: 291560,
				PositionRef:    "refs/heads/master",
			},
			{
				ID:             "chromium-codesearch/chromium/src-0000000000000000000000000000000000291560",
				CommitHash:     "0000000000000000000000000000000000291560",
				Host:           "chromium",
				Repository:     "codesearch/chromium/src",
				PositionNumber: 291560,
				PositionRef:    "refs/heads/master",
			},
			{
				ID:             "foo-baz-0000000000000000000000000000000000291561",
				CommitHash:     "0000000000000000000000000000000000291561",
				Host:           "chromium",
				Repository:     "codesearch/chromium/src",
				PositionNumber: 291561,
				PositionRef:    "refs/heads/master",
			},
			{
				ID:             "foo-baz-0000000000000000000000000000000000291562",
				CommitHash:     "0000000000000000000000000000000000291562",
				Host:           "chromium",
				Repository:     "chromium/src",
				PositionNumber: 291562,
				PositionRef:    "refs/heads/main",
			},
			{
				ID:             "foo-baz-0000000000000000000000000000000000291563",
				CommitHash:     "0000000000000000000000000000000000291563",
				Host:           "chromium",
				Repository:     "chromium/src",
				PositionNumber: 291563,
				PositionRef:    "refs/heads/feature",
			},
		}
		datastore.Put(ctx, commits)

		Convey("svn position style", func() {
			Convey("release ", func() {
				url, err := r.findRedirectURL(ctx, "/1")
				So(err, ShouldBeNil)
				So(url, ShouldEqual, "https://chromium.googlesource.com/chromium/src/+/0000000000000000000000000000000000000001")
			})

			Convey("trunk", func() {
				url, err := r.findRedirectURL(ctx, "/2")
				So(err, ShouldBeNil)
				So(url, ShouldEqual, "https://chromium.googlesource.com/chromium/src/+/0000000000000000000000000000000000000002")
			})

			Convey("trunk src", func() {
				url, err := r.findRedirectURL(ctx, "/288197")
				So(err, ShouldBeNil)
				So(url, ShouldEqual, "https://chromium.googlesource.com/chromium/src/+/0000000000000000000000000000000000288197")
			})

			Convey("non chromium", func() {
				_, err := r.findRedirectURL(ctx, "/3")
				So(err, ShouldEqual, errNoMatch)
			})
		})

		Convey("git numberer", func() {
			Convey("with mirror", func() {
				url, err := r.findRedirectURL(ctx, "/291560")
				So(err, ShouldBeNil)
				So(url, ShouldEqual, "https://chromium.googlesource.com/chromium/src/+/0000000000000000000000000000000000291560")
			})

			Convey("not chromium repo", func() {
				_, err := r.findRedirectURL(ctx, "/291561")
				So(err, ShouldEqual, errNoMatch)
			})

			Convey("main branch", func() {
				url, err := r.findRedirectURL(ctx, "/291562")
				So(err, ShouldBeNil)
				So(url, ShouldEqual, "https://chromium.googlesource.com/chromium/src/+/0000000000000000000000000000000000291562")
			})

			Convey("non default branch", func() {
				_, err := r.findRedirectURL(ctx, "/291563")
				So(err, ShouldEqual, errNoMatch)
			})

		})
	})

	Convey("full hash redirect", t, func() {
		ctx := redirectTestSetup()
		commits := []*models.Commit{
			{
				ID:         "foo-bar-0000000000000000000000000000000000000000",
				CommitHash: "0000000000000000000000000000000000000000",
				Host:       "foo",
				Repository: "bar",
			},
			{
				ID:         "foo-bar/mirror-0000000000000000000000000000000000000000",
				CommitHash: "0000000000000000000000000000000000000000",
				Host:       "foo",
				Repository: "bar/mirror",
			},
			{
				ID:         "foo-baz-0000000000000000000000000000000000000001",
				CommitHash: "0000000000000000000000000000000000000001",
				Host:       "foo",
				Repository: "baz",
			},
			{
				ID:         "foo-baz/mirror-0000000000000000000000000000000000000001",
				CommitHash: "0000000000000000000000000000000000000001",
				Host:       "foo",
				Repository: "baz/not/indexed",
			},
			{
				ID:         "foo-baz/mirror-0000000000000000000000000000000000000002",
				CommitHash: "0000000000000000000000000000000000000002",
				Host:       "foo",
				Repository: "baz/not/indexed",
			},
		}
		datastore.Put(ctx, commits)

		err := config.Override(ctx, &config.Config{
			Hosts: []*config.Host{
				{
					Name: "foo",
					Repos: []*config.Repository{
						{
							Name:     "bar",
							Indexing: &config.Repository_Priority{Priority: true},
						},
						{
							Name:     "baz/not/indexed",
							Indexing: &config.Repository_DoNotIndex{DoNotIndex: true},
						},
					},
				},
			},
		})
		if err != nil {
			panic(fmt.Sprintf("Could not set context: %s", err.Error()))
		}

		Convey("respect priority", func() {
			url, err := r.findRedirectURL(
				ctx, "/0000000000000000000000000000000000000000")
			So(err, ShouldBeNil)
			So(url, ShouldEqual, "https://foo.googlesource.com/bar/+/0000000000000000000000000000000000000000")
		})

		Convey("avoid not indexed", func() {
			url, err := r.findRedirectURL(
				ctx, "/0000000000000000000000000000000000000001")
			So(err, ShouldBeNil)
			So(url, ShouldEqual, "https://foo.googlesource.com/baz/+/0000000000000000000000000000000000000001")
		})

		Convey("redirect to not indexed if only one", func() {
			url, err := r.findRedirectURL(
				ctx, "/0000000000000000000000000000000000000002")
			So(err, ShouldBeNil)
			So(url, ShouldEqual, "https://foo.googlesource.com/baz/not/indexed/+/0000000000000000000000000000000000000002")
		})
	})

	Convey("default not found", t, func() {
		ctx := redirectTestSetup()
		_, err := r.findRedirectURL(
			ctx, "/foo")
		So(err, ShouldEqual, errNoMatch)
	})
}
