// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"

	tricium "infra/tricium/api/v1"
)

// These tests read from files on the filesystem, so modifying the tests may
// require modifying the example test files.
const (
	baseDir          = "test"
	gLinks           = "src/g_links.md"
	goLinks          = "src/go_links.md"
	httpsURLs        = "src/https_urls.md"
	httpURL          = "src/http_single_url.md"
	multipleHTTPURLs = "src/http_multiple_urls.md"
)

func TestHTTPSChecker(t *testing.T) {

	Convey("Produces no comment for g/ link", t, func() {
		results := &tricium.Data_Results{}
		checkHTTPS(baseDir, gLinks, results)
		So(results.Comments, ShouldBeNil)
	})

	Convey("Produces no comment for file with go/ link", t, func() {
		results := &tricium.Data_Results{}
		checkHTTPS(baseDir, goLinks, results)
		So(results.Comments, ShouldBeNil)
	})

	Convey("Produces no comment for file with httpsURLs", t, func() {
		results := &tricium.Data_Results{}
		checkHTTPS(baseDir, httpsURLs, results)
		So(results.Comments, ShouldBeNil)
	})

	Convey("Flags a single http URL", t, func() {
		results := &tricium.Data_Results{}
		checkHTTPS(baseDir, httpURL, results)
		So(results.Comments, ShouldNotBeNil)
		So(results.Comments[0], ShouldResembleProto, &tricium.Data_Comment{
			Category:  "HttpsCheck/Warning",
			Message:   ("Nit: Replace http:// URLs with https://"),
			Path:      httpURL,
			StartLine: 5,
			EndLine:   5,
			StartChar: 7,
			EndChar:   24,
		})
	})

	Convey("Flags multiple http URLs", t, func() {
		results := &tricium.Data_Results{}
		checkHTTPS(baseDir, multipleHTTPURLs, results)
		So(len(results.Comments), ShouldEqual, 2)
		So(results.Comments[1], ShouldResembleProto, &tricium.Data_Comment{
			Category:  "HttpsCheck/Warning",
			Message:   ("Nit: Replace http:// URLs with https://"),
			Path:      multipleHTTPURLs,
			StartLine: 9,
			EndLine:   9,
			StartChar: 7,
			EndChar:   26,
		})
	})
}
