// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package testutils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func Assert[T any](tb testing.TB, x any) T {
	tb.Helper()
	ret, ok := x.(T)
	convey.So(ok, convey.ShouldBeTrue)
	return ret
}
