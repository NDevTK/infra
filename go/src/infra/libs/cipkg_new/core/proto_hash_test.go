// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package core

import (
	"crypto/sha256"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestProtoHash(t *testing.T) {
	Convey("stable hash for proto message", t, func() {
		h := sha256.New()
		Convey("derivation", func() {
			err := StableHash(h, &Derivation{})
			So(err, ShouldBeNil)
			So(fmt.Sprintf("%x", h.Sum(nil)), ShouldEqual, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
		})
		Convey("action with extension", func() {
			a, err := anypb.New(&ActionURLFetch{})
			So(err, ShouldBeNil)
			err = StableHash(h, &Action{
				Spec: &Action_Extension{
					Extension: a,
				},
			})
			So(err, ShouldBeNil)
		})
		Convey("action with invalid extension", func() {
			err := StableHash(h, &Action{
				Spec: &Action_Extension{},
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid empty type URL")
		})
		Convey("unknown field", func() {
			drv := &Derivation{}
			drv.ProtoReflect().SetUnknown([]byte{0})
			err := StableHash(h, drv)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unknown fields cannot be hashed")
		})
	})
}
