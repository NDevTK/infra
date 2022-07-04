// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExtractorUtil(t *testing.T) {
	Convey("NormalizeFilePath", t, func() {
		data := map[string]string{
			"../a/b/c.cc":    "a/b/c.cc",
			"a/b/./c.cc":     "a/b/c.cc",
			"a/b/../c.cc":    "a/c.cc",
			"a\\b\\.\\c.cc":  "a/b/c.cc",
			"a\\\\b\\\\c.cc": "a/b/c.cc",
			"//a/b/c.cc":     "a/b/c.cc",
		}
		for fp, nfp := range data {
			So(NormalizeFilePath(fp), ShouldEqual, nfp)
		}
	})

	Convey("GetCanonicalFileName", t, func() {
		data := map[string]string{
			"../a/b/c.cc":   "c",
			"a/b/./d.dd":    "d",
			"a/b/c.xx":      "c",
			"a/b/c_impl.xx": "c",
		}
		for fp, name := range data {
			So(GetCanonicalFileName(fp), ShouldEqual, name)
		}
	})

	Convey("StripExtensionAndCommonSuffixFromFileName", t, func() {
		data := map[string]string{
			"a_file_impl_mac_test.cc": "a_file",
			"src/b_file_x11_ozone.h":  "src/b_file",
			"c_file.cc":               "c_file",
		}
		for k, v := range data {
			So(StripExtensionAndCommonSuffixFromFileName(k), ShouldEqual, v)
		}
	})

}
