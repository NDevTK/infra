// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package scopes

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// Testing param methods.
func TestConfigScope(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Happy case", t, func() {
		ctx := WithConfigScope(ctx)
		PutConfigParam(ctx, "key1", "hello")
		v, ok := ReadConfigParam(ctx, "key1")
		So(ok, ShouldBeTrue)
		So(v, ShouldEqual, "hello")
	})
	Convey("Read non existent key", t, func() {
		ctx := WithConfigScope(ctx)
		_, ok := ReadConfigParam(ctx, "key1")
		So(ok, ShouldBeFalse)
	})
	Convey("Read all keys", t, func() {
		ctx := WithConfigScope(ctx)
		PutConfigParam(ctx, "key1", "hello1")
		PutConfigParam(ctx, "key2", "hello2")
		PutConfigParam(ctx, "key3", "hello3")
		v, ok := ReadConfigParam(ctx, "key1")
		So(ok, ShouldBeTrue)
		So(v, ShouldEqual, "hello1")
		v, ok = ReadConfigParam(ctx, "key2")
		So(ok, ShouldBeTrue)
		So(v, ShouldEqual, "hello2")
		v, ok = ReadConfigParam(ctx, "key3")
		So(ok, ShouldBeTrue)
		So(v, ShouldEqual, "hello3")
	})
	Convey("Try to put before initilize", t, func() {
		PutConfigParam(ctx, "key1", "hello")
		_, ok := ReadConfigParam(ctx, "key1")
		So(ok, ShouldBeFalse)
	})
}
