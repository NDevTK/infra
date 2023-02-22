// Copyright 2023 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package scopes

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// Testing param methods.
func TestParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Simple read", t, func() {
		m := map[string]any{
			"k1": "v1",
			"k2": "v2",
			"k3": "v3",
		}
		ctx = WithParams(ctx, m)
		copy := GetParamCopy(ctx)
		So(m, ShouldResemble, copy)
	})
	Convey("Read an existent key", t, func() {
		m := map[string]any{
			"k4": "v4",
		}
		ctx = WithParams(ctx, m)
		v, ok := GetParam(ctx, "k4")
		So(ok, ShouldEqual, true)
		So(v, ShouldEqual, "v4")
	})
	Convey("Read nonexistent key", t, func() {
		m := map[string]any{
			"k5": "v5",
		}
		ctx = WithParams(ctx, m)
		_, ok := GetParam(ctx, "k6")
		So(ok, ShouldEqual, false)

	})
	Convey("Read nonexistent key (2)", t, func() {
		_, ok := GetParam(ctx, "k6")
		So(ok, ShouldEqual, false)
	})
}
