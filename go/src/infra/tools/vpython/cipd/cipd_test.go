// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package cipd

import (
	"bytes"
	"testing"

	"infra/tools/vpython/api/env"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWriteEnsureFile(t *testing.T) {
	t.Parallel()

	Convey(`Test manifest generation`, t, func() {
		var buf bytes.Buffer

		Convey(`Invalid packages will panic.`, func() {
			pkg := &env.Spec_Package{
				Path: "foo/bar",
			}
			So(validatePackage(pkg), ShouldNotBeNil)
			So(func() {
				_ = writeEnsureFile(&buf, []*env.Spec_Package{pkg})
			}, ShouldPanic)
		})

		Convey(`Can write a manifest for packages`, func() {
			err := writeEnsureFile(&buf, []*env.Spec_Package{
				&env.Spec_Package{"foo/bar", "baz"},
				&env.Spec_Package{"pants", "key:value"},
			})
			So(err, ShouldBeNil)
			So(buf.String(), ShouldResemble, "foo/bar baz\npants key:value\n")
		})
	})
}
