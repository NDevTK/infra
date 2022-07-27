// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package perms

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/resultdb/rdbperms"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
	"go.chromium.org/luci/server/auth/realms"
	"google.golang.org/grpc/codes"
)

func init() {
	rdbperms.PermListTestResults.AddFlags(realms.UsedInQueryRealms)
	rdbperms.PermListTestExonerations.AddFlags(realms.UsedInQueryRealms)
	rdbperms.PermGetArtifact.AddFlags(realms.UsedInQueryRealms)
	rdbperms.PermListArtifacts.AddFlags(realms.UsedInQueryRealms)
}

func TestQueryRealms(t *testing.T) {
	Convey("QueryRealms", t, func() {
		ctx := context.Background()

		ctx = auth.WithState(ctx, &authtest.FakeState{
			Identity: "user:someone@example.com",
			IdentityPermissions: []authtest.RealmPermission{
				{
					Realm:      "project1:realm1",
					Permission: rdbperms.PermListTestResults,
				},
				{
					Realm:      "project1:realm1",
					Permission: rdbperms.PermListTestExonerations,
				},
				{
					Realm:      "project1:realm1",
					Permission: rdbperms.PermGetArtifact,
				},
				{
					Realm:      "project1:realm2",
					Permission: rdbperms.PermListTestResults,
				},
				{
					Realm:      "project1:realm2",
					Permission: rdbperms.PermListTestExonerations,
				},
				{
					Realm:      "project2:realm1",
					Permission: rdbperms.PermListTestResults,
				},
				{
					Realm:      "project2:realm1",
					Permission: rdbperms.PermGetArtifact,
				},
				{
					Realm:      "project2:realm1",
					Permission: rdbperms.PermListArtifacts,
				},
			},
		})

		Convey("no permission specified", func() {
			realms, err := QueryRealms(ctx, "project1", "realm1", nil)
			So(err, ShouldErrLike, "at least one permission must be provided")
			So(realms, ShouldBeEmpty)
		})

		Convey("specified subRealm  without project", func() {
			realms, err := QueryRealms(ctx, "", "realm1", nil, rdbperms.PermListTestResults)
			So(err, ShouldErrLike, "project must be specified when the subRealm is specified")
			So(realms, ShouldBeEmpty)
		})

		Convey("global scope", func() {
			Convey("check single permission", func() {
				realms, err := QueryRealms(ctx, "", "", nil, rdbperms.PermListTestResults)
				So(err, ShouldBeNil)
				So(realms, ShouldResemble, []string{"project1:realm1", "project1:realm2", "project2:realm1"})
			})

			Convey("check multiple permissions", func() {
				realms, err := QueryRealms(ctx, "", "", nil, rdbperms.PermListTestResults, rdbperms.PermGetArtifact)
				So(err, ShouldBeNil)
				So(realms, ShouldResemble, []string{"project1:realm1", "project2:realm1"})
			})

			Convey("no matched realms", func() {
				realms, err := QueryRealms(ctx, "", "", nil, rdbperms.PermListTestExonerations, rdbperms.PermListArtifacts)
				So(err, ShouldErrLike, "caller does not have permissions", "in any projects")
				So(err, ShouldHaveAppStatus, codes.PermissionDenied)
				So(realms, ShouldBeEmpty)
			})
		})

		Convey("project scope", func() {
			Convey("check single permission", func() {
				realms, err := QueryRealms(ctx, "project1", "", nil, rdbperms.PermListTestResults)
				So(err, ShouldBeNil)
				So(realms, ShouldResemble, []string{"project1:realm1", "project1:realm2"})
			})

			Convey("check multiple permissions", func() {
				realms, err := QueryRealms(ctx, "project1", "", nil, rdbperms.PermListTestResults, rdbperms.PermGetArtifact)
				So(err, ShouldBeNil)
				So(realms, ShouldResemble, []string{"project1:realm1"})
			})

			Convey("no matched realms", func() {
				realms, err := QueryRealms(ctx, "project1", "", nil, rdbperms.PermListTestExonerations, rdbperms.PermListArtifacts)
				So(err, ShouldErrLike, "caller does not have permissions", "in project \"project1\"")
				So(err, ShouldHaveAppStatus, codes.PermissionDenied)
				So(realms, ShouldBeEmpty)
			})
		})

		Convey("realm scope", func() {
			Convey("check single permission", func() {
				realms, err := QueryRealms(ctx, "project1", "realm1", nil, rdbperms.PermListTestResults)
				So(err, ShouldBeNil)
				So(realms, ShouldResemble, []string{"project1:realm1"})
			})

			Convey("check multiple permissions", func() {
				realms, err := QueryRealms(ctx, "project1", "realm1", nil, rdbperms.PermListTestResults, rdbperms.PermGetArtifact)
				So(err, ShouldBeNil)
				So(realms, ShouldResemble, []string{"project1:realm1"})
			})

			Convey("no matched realms", func() {
				realms, err := QueryRealms(ctx, "project1", "realm1", nil, rdbperms.PermListTestExonerations, rdbperms.PermListArtifacts)
				So(err, ShouldErrLike, "caller does not have permission", "in realm \"project1:realm1\"")
				So(err, ShouldHaveAppStatus, codes.PermissionDenied)
				So(realms, ShouldBeEmpty)
			})
		})
	})
}
