// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/types/known/anypb"

	kpb "infra/cmd/package_index/kythe/proto"
)

func TestGetClangUtil(t *testing.T) {
	t.Parallel()
	cwd, _ := os.Getwd()
	clangInfo := &clangUnitInfo{
		unit: clangUnit{
			Command: "foo clang++ bar baz",
		},
		filepathsFn: filepath.Join(cwd, "package_index_testdata",
			"input", "src", "out", "Debug", "gen", "main.pb.h"),
	}
	Convey("linux", t, func() {
		cu, err := getClangUnit(
			context.Background(),
			clangInfo,
			"rootPath",
			"out/dir",
			"corpus",
			"linux",
			"",
			&FileHashMap{})
		So(err, ShouldBeNil)
		So(cu, ShouldNotBeNil)

		details := &kpb.BuildDetails{
			BuildConfig: "linux",
		}
		detail, _ := anypb.New(details)
		detail.TypeUrl = "kythe.io/proto/kythe.proto.BuildDetails"
		So(cu.Argument, ShouldResemble,
			[]string{"clang++", "bar", "baz",
				"-DKYTHE_IS_RUNNING=1", "-w"})
	})
	Convey("linux-arm64", t, func() {
		cu, err := getClangUnit(
			context.Background(),
			clangInfo,
			"rootPath",
			"out/dir",
			"corpus",
			"linux",
			"arm64",
			&FileHashMap{})
		So(err, ShouldBeNil)
		So(cu, ShouldNotBeNil)

		details := &kpb.BuildDetails{
			BuildConfig: "linux",
		}
		detail, _ := anypb.New(details)
		detail.TypeUrl = "kythe.io/proto/kythe.proto.BuildDetails"
		So(cu.Argument, ShouldResemble,
			[]string{"clang++", "bar", "baz",
				"-target", "arm64",
				"-DKYTHE_IS_RUNNING=1", "-w"})
	})
	Convey("mac", t, func() {
		cu, err := getClangUnit(
			context.Background(),
			clangInfo,
			"rootPath",
			"out/dir",
			"corpus",
			"mac",
			"",
			&FileHashMap{})
		So(err, ShouldBeNil)
		So(cu, ShouldNotBeNil)

		details := &kpb.BuildDetails{
			BuildConfig: "linux",
		}
		detail, _ := anypb.New(details)
		detail.TypeUrl = "kythe.io/proto/kythe.proto.BuildDetails"
		So(cu.Argument, ShouldResemble,
			[]string{"clang++", "bar", "baz",
				"-target", "x86_64-apple-darwin20.6.0",
				"-DKYTHE_IS_RUNNING=1", "-w"})
	})
}
