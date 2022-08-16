// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package model

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSuspectJutification(t *testing.T) {
	t.Parallel()

	Convey("SuspectJutification", t, func() {
		justification := &SuspectJustification{}
		justification.AddItem(10, "a/b", "fileInLog", JustificationType_FAILURELOG)
		So(justification.GetScore(), ShouldEqual, 10)
		justification.AddItem(2, "c/d", "fileInDependency1", JustificationType_DEPENDENCY)
		So(justification.GetScore(), ShouldEqual, 12)
		justification.AddItem(8, "e/f", "fileInDependency2", JustificationType_DEPENDENCY)
		So(justification.GetScore(), ShouldEqual, 19)
	})
}
