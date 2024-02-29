// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"context"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/server/secrets"
)

func TestGetSecret(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := &secrets.SecretManagerStore{
		CloudProject: "test-project",
	}
	ctx = secrets.Use(ctx, store)

	Convey("Test GetSecret", t, func() {
		Convey("success - valid secret location in base64", func() {
			secret, err := GetSecret(ctx, "devsecret://cGFzc3dvcmQ")
			So(err, ShouldBeNil)
			So(secret, ShouldEqual, "password")
		})
		Convey("fail - invalid secret location with padding", func() {
			secret, err := GetSecret(ctx, "devsecret://cGFzc3dvcmQ=")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "not base64 encoding")
			So(secret, ShouldBeEmpty)
		})
	})
}

func TestGetEnvVar(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Test GetEnvVar", t, func() {
		Convey("success - env var found", func() {
			err := os.Setenv("FOO", "BAR")
			So(err, ShouldBeNil)

			v, err := GetEnvVar(ctx, "FOO")
			So(err, ShouldBeNil)
			So(v, ShouldEqual, "BAR")

			err = os.Setenv("FOO", "")
			So(err, ShouldBeNil)
		})
		Convey("fail - env var not found", func() {
			v, err := GetEnvVar(ctx, "NO_FOO")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "environment variable not set")
			So(v, ShouldBeEmpty)
		})
	})
}
