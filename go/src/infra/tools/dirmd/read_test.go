// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dirmd

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	dirmdpb "infra/tools/dirmd/proto"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestRead(t *testing.T) {
	t.Parallel()

	testDataKey := "go/src/infra/tools/dirmd/testdata"
	dummyRepos := map[string]*dirmdpb.Repo{
		".": {Mixins: map[string]*dirmdpb.Metadata{}},
	}

	mxKey := testDataKey + "/mixins"
	dummyMixinRepos := map[string]*dirmdpb.Repo{
		".": {Mixins: map[string]*dirmdpb.Metadata{
			"//" + mxKey + "/FOO_METADATA": {
				Monorail: &dirmdpb.Monorail{
					Project:   "chromium",
					Component: "foo",
				},
			},
		}},
	}

	Convey(`ReadMapping`, t, func() {
		ctx := context.Background()
		rootKey := testDataKey + "/root"

		Convey(`Original`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_ORIGINAL, false, "testdata/root")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey: {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
					rootKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						// OS was not inherited
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Resultdb: &dirmdpb.ResultDB{
							Tags: []string{
								"feature:read-later",
								"feature:another-one",
							},
						},
					},
					rootKey + "/subdir_with_files": {
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Overrides: []*dirmdpb.MetadataOverride{
							{
								FilePatterns: []string{
									"*.txt",
								},
								Metadata: &dirmdpb.Metadata{
									Monorail: &dirmdpb.Monorail{
										Component: "Some>Other>Component",
									},
								},
							},
							{
								FilePatterns: []string{
									"*.json",
								},
								Metadata: &dirmdpb.Metadata{
									Mixins: []string{
										"//" + mxKey + "/FOO_METADATA",
									},
								},
							},
						},
					},
					rootKey + "/subdir_with_owners": {
						TeamEmail: "team-email@chromium.org",
						// OS was not inherited
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
					// "subdir_with_owners/empty_subdir" is not present because it has
					// no metadata.
				},
				Files: map[string]*dirmdpb.Metadata{
					rootKey + "/subdir_with_files/dummy.json": {
						Mixins: []string{
							"//" + mxKey + "/FOO_METADATA",
						},
					},
					rootKey + "/subdir_with_files/dummy.txt": {
						Monorail: &dirmdpb.Monorail{
							Component: "Some>Other>Component",
						},
					},
				},
				Repos: dummyMixinRepos,
			})
		})

		Convey(`Original with two dirs`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_ORIGINAL, false, "testdata/root/subdir", "testdata/root/subdir_with_owners")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey: {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
					rootKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						// OS was not inherited
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Resultdb: &dirmdpb.ResultDB{
							Tags: []string{
								"feature:read-later",
								"feature:another-one",
							},
						},
					},
					rootKey + "/subdir_with_owners": {
						TeamEmail: "team-email@chromium.org",
						// OS was not inherited
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
					// "subdir_with_owners/empty_subdir" is not present because it has
					// no metadata.
				},
				Repos: dummyRepos,
			})
		})

		Convey(`Original with two dirs metadata`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_ORIGINAL, false, "testdata/root/subdir", "testdata/root/subdir_with_files")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey: {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
					rootKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						// OS was not inherited
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Resultdb: &dirmdpb.ResultDB{
							Tags: []string{
								"feature:read-later",
								"feature:another-one",
							},
						},
					},
					rootKey + "/subdir_with_files": {
						// TeamEmail and OS were not inherited
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Overrides: []*dirmdpb.MetadataOverride{
							{
								FilePatterns: []string{
									"*.txt",
								},
								Metadata: &dirmdpb.Metadata{
									Monorail: &dirmdpb.Monorail{
										Component: "Some>Other>Component",
									},
								},
							},
							{
								FilePatterns: []string{
									"*.json",
								},
								Metadata: &dirmdpb.Metadata{
									Mixins: []string{
										"//" + mxKey + "/FOO_METADATA",
									},
								},
							},
						},
					},
					// "subdir_with_owners/nested_dir" is not present because it has
					// no metadata.
				},
				Files: map[string]*dirmdpb.Metadata{
					rootKey + "/subdir_with_files/dummy.json": {
						Mixins: []string{
							"//" + mxKey + "/FOO_METADATA",
						},
					},
					rootKey + "/subdir_with_files/dummy.txt": {
						Monorail: &dirmdpb.Monorail{
							Component: "Some>Other>Component",
						},
					},
				},
				Repos: dummyMixinRepos,
			})
		})

		Convey(`Full`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_FULL, false, "testdata/root")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey: {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
					rootKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Resultdb: &dirmdpb.ResultDB{
							Tags: []string{
								"feature:read-later",
								"feature:another-one",
							},
						},
					},
					rootKey + "/subdir_with_files": {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
					rootKey + "/subdir_with_files/nested_dir": {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
					rootKey + "/subdir_with_owners": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
					rootKey + "/subdir_with_owners/empty_subdir": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
				},
				Files: map[string]*dirmdpb.Metadata{
					rootKey + "/subdir_with_files/dummy.txt": {
						Monorail: &dirmdpb.Monorail{
							Component: "Some>Other>Component",
						},
					},
					rootKey + "/subdir_with_files/dummy.json": {
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "foo",
						},
					},
				},
				Repos: dummyMixinRepos,
			})
		})

		Convey(`Computed`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_COMPUTED, false, "testdata/root")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey: {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
					rootKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Resultdb: &dirmdpb.ResultDB{
							Tags: []string{
								"feature:read-later",
								"feature:another-one",
							},
						},
					},
					rootKey + "/subdir_with_files": {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
					rootKey + "/subdir_with_owners": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
				},
				Files: map[string]*dirmdpb.Metadata{
					rootKey + "/subdir_with_files/dummy.txt": {
						Monorail: &dirmdpb.Monorail{
							Component: "Some>Other>Component",
						},
					},
					rootKey + "/subdir_with_files/dummy.json": {
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "foo",
						},
					},
				},
				Repos: dummyMixinRepos,
			})
		})

		Convey(`Computed, not from root`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_COMPUTED, false, "testdata/root/subdir")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey: {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
					rootKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Resultdb: &dirmdpb.ResultDB{
							Tags: []string{
								"feature:read-later",
								"feature:another-one",
							},
						},
					},
				},
				Repos: dummyRepos,
			})
		})

		Convey(`Computed, only DIR_METADATA`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_COMPUTED, true, "testdata/root")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey: {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
					rootKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Resultdb: &dirmdpb.ResultDB{
							Tags: []string{
								"feature:read-later",
								"feature:another-one",
							},
						},
					},
					rootKey + "/subdir_with_files": {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
				},
				Files: map[string]*dirmdpb.Metadata{
					rootKey + "/subdir_with_files/dummy.txt": {
						Monorail: &dirmdpb.Monorail{
							Component: "Some>Other>Component",
						},
					},
					rootKey + "/subdir_with_files/dummy.json": {
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "foo",
						},
					},
				},
				Repos: dummyMixinRepos,
			})
		})

		Convey(`Computed, from a symlink`, func() {
			if runtime.GOOS == "windows" {
				return
			}
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_COMPUTED, false, "testdata/sym_root")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey: {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
					rootKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Resultdb: &dirmdpb.ResultDB{
							Tags: []string{
								"feature:read-later",
								"feature:another-one",
							},
						},
					},
					rootKey + "/subdir_with_files": {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
					rootKey + "/subdir_with_owners": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
				},
				Files: map[string]*dirmdpb.Metadata{
					rootKey + "/subdir_with_files/dummy.txt": {
						Monorail: &dirmdpb.Monorail{
							Component: "Some>Other>Component",
						},
					},
					rootKey + "/subdir_with_files/dummy.json": {
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "foo",
						},
					},
				},
				Repos: dummyMixinRepos,
			})
		})

		Convey(`Computed, with mixin`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_COMPUTED, false, "testdata/mixins")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					mxKey: {
						TeamEmail: "team-email@chromium.org",
					},
					mxKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium", // from FOO_METADATA
							Component: "bar",      // from BAR_METADATA
						},
					},
				},
				Repos: map[string]*dirmdpb.Repo{
					".": {
						Mixins: map[string]*dirmdpb.Metadata{
							"//" + mxKey + "/FOO_METADATA": {
								Monorail: &dirmdpb.Monorail{
									Project:   "chromium",
									Component: "foo",
								},
							},
							"//" + mxKey + "/BAR_METADATA": {
								Monorail: &dirmdpb.Monorail{
									Component: "bar",
								},
							},
						},
					},
				},
			})
		})

		Convey(`Sparse`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_SPARSE, false, "testdata/root/subdir")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmdpb.OS_LINUX,
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Resultdb: &dirmdpb.ResultDB{
							Tags: []string{
								"feature:read-later",
								"feature:another-one",
							},
						},
					},
				},
				Repos: dummyRepos,
			})
		})

		Convey(`Sparse, only DIR_METADATA`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_SPARSE, true, "testdata/root/subdir_with_owners/")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					// Include inherited metadata from root/DIR_METADATA, the content of
					// its OWNERS file is not included.
					rootKey + "/subdir_with_owners": {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
				},
				Repos: dummyRepos,
			})
		})

		Convey(`Sparse, with mixins`, func() {
			mxKey := testDataKey + "/mixins"
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_SPARSE, false, "testdata/mixins/subdir")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					mxKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium", // from FOO_METADATA
							Component: "bar",      // from BAR_METADATA
						},
					},
				},
				Repos: map[string]*dirmdpb.Repo{
					".": {
						Mixins: map[string]*dirmdpb.Metadata{
							"//" + mxKey + "/FOO_METADATA": {
								Monorail: &dirmdpb.Monorail{
									Project:   "chromium",
									Component: "foo",
								},
							},
							"//" + mxKey + "/BAR_METADATA": {
								Monorail: &dirmdpb.Monorail{
									Component: "bar",
								},
							},
						},
					},
				},
			})
		})

		Convey(`Sparse, from a symlink`, func() {
			if runtime.GOOS == "windows" {
				return
			}
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_SPARSE, false, "testdata/sym_root")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey: {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
				},
				Repos: dummyRepos,
			})
		})

		Convey(`Reduced`, func() {
			m, err := ReadMapping(ctx, dirmdpb.MappingForm_REDUCED, false, "testdata/root")
			So(err, ShouldBeNil)
			So(m.Proto(), ShouldResembleProto, &dirmdpb.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					rootKey: {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmdpb.OS_LINUX,
					},
					rootKey + "/subdir": {
						TeamEmail: "team-email@chromium.org",
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Resultdb: &dirmdpb.ResultDB{
							Tags: []string{
								"feature:read-later",
								"feature:another-one",
							},
						},
					},
					rootKey + "/subdir_with_files": {
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
						Overrides: []*dirmdpb.MetadataOverride{
							{
								FilePatterns: []string{
									"*.txt",
								},
								Metadata: &dirmdpb.Metadata{
									Monorail: &dirmdpb.Monorail{
										Component: "Some>Other>Component",
									},
								},
							},
							{
								FilePatterns: []string{
									"*.json",
								},
								Metadata: &dirmdpb.Metadata{
									Mixins: []string{
										"//" + mxKey + "/FOO_METADATA",
									},
								},
							},
						},
					},
					rootKey + "/subdir_with_owners": {
						TeamEmail: "team-email@chromium.org",
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
				},
				Files: map[string]*dirmdpb.Metadata{
					rootKey + "/subdir_with_files/dummy.txt": {
						Monorail: &dirmdpb.Monorail{
							Component: "Some>Other>Component",
						},
					},
					rootKey + "/subdir_with_files/dummy.json": {
						Monorail: &dirmdpb.Monorail{
							Project:   "chromium",
							Component: "foo",
						},
					},
				},
				Repos: dummyMixinRepos,
			})
		})
	})
}

func TestRemoveRedundantDirs(t *testing.T) {
	t.Parallel()

	Convey("TestRemoveRedundantDirs", t, func() {
		actual := removeRedundantDirs(
			filepath.FromSlash("x/y2/z"),
			filepath.FromSlash("a"),
			filepath.FromSlash("a/b"),
			filepath.FromSlash("x/y1"),
			filepath.FromSlash("x/y2"),
		)
		So(actual, ShouldResemble, []string{
			filepath.FromSlash("a"),
			filepath.FromSlash("x/y1"),
			filepath.FromSlash("x/y2"),
		})
	})
}
