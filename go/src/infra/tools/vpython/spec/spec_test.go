// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spec

import (
	"testing"

	"infra/tools/vpython/api/env"

	. "github.com/luci/luci-go/common/testing/assertions"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNormalizeAndHash(t *testing.T) {
	t.Parallel()

	pkgFoo := &env.Spec_Package{Path: "foo", Version: "1"}
	pkgBar := &env.Spec_Package{Path: "bar", Version: "2"}
	pkgBaz := &env.Spec_Package{Path: "baz", Version: "3"}

	Convey(`Test manifest generation`, t, func() {
		var spec env.Spec

		Convey(`Will normalize an empty spec`, func() {
			So(Normalize(&spec), ShouldBeNil)
			So(spec, ShouldResemble, env.Spec{})
		})

		Convey(`Will normalize to sorted order.`, func() {
			spec.Wheel = []*env.Spec_Package{pkgFoo, pkgBar, pkgBaz}
			So(Normalize(&spec), ShouldBeNil)
			So(spec, ShouldResemble, env.Spec{
				Wheel: []*env.Spec_Package{pkgBar, pkgBaz, pkgFoo},
			})

			So(Hash(&spec), ShouldEqual, "904b958b73206e64ce267e2b6df5fa0f1223e348327dba5e9a5f6421bb42cadb")
		})

		Convey(`Will fail to normalize if there are duplicate wheels.`, func() {
			spec.Wheel = []*env.Spec_Package{pkgFoo, pkgFoo, pkgBar, pkgBaz}
			So(Normalize(&spec), ShouldErrLike, "duplicate spec entries")

			// Even if the versions differ.
			fooClone := *pkgFoo
			fooClone.Version = "other"
			spec.Wheel = []*env.Spec_Package{pkgFoo, &fooClone, pkgBar, pkgBaz}
			So(Normalize(&spec), ShouldErrLike, "duplicate spec entries")
		})
	})
}
