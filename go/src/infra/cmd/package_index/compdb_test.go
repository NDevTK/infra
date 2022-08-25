package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/ptypes"
	. "github.com/smartystreets/goconvey/convey"

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
			&FileHashMap{})
		So(err, ShouldBeNil)
		So(cu, ShouldNotBeNil)

		details := &kpb.BuildDetails{
			BuildConfig: "linux",
		}
		detail, _ := ptypes.MarshalAny(details)
		detail.TypeUrl = "kythe.io/proto/kythe.proto.BuildDetails"
		So(cu.Argument, ShouldResemble,
			[]string{"clang++", "bar", "baz",
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
			&FileHashMap{})
		So(err, ShouldBeNil)
		So(cu, ShouldNotBeNil)

		details := &kpb.BuildDetails{
			BuildConfig: "linux",
		}
		detail, _ := ptypes.MarshalAny(details)
		detail.TypeUrl = "kythe.io/proto/kythe.proto.BuildDetails"
		So(cu.Argument, ShouldResemble,
			[]string{"clang++", "bar", "baz",
				"-target", "x86_64-apple-darwin20.6.0",
				"-DKYTHE_IS_RUNNING=1", "-w"})
	})
}
