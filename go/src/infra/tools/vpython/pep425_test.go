// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/luci/luci-go/vpython/api/vpython"

	. "github.com/smartystreets/goconvey/convey"
)

func mkTag(version, abi, arch string) *vpython.Environment_Pep425Tag {
	return &vpython.Environment_Pep425Tag{
		Version: version,
		Abi:     abi,
		Arch:    arch,
	}
}

func TestPEP425TagSelector(t *testing.T) {
	t.Parallel()

	Convey(`Testing PEP425 tag selection`, t, func() {
		for _, tc := range []struct {
			goOS     string
			tags     []*vpython.Environment_Pep425Tag
			template map[string]string
		}{
			{"linux", nil, nil},

			{
				"linux",
				[]*vpython.Environment_Pep425Tag{
					mkTag("py2", "none", "any"),
					mkTag("py27", "none", "any"),
					mkTag("cp27", "cp27mu", "linux_x86_64"),
					mkTag("cp27", "cp27mu", "manylinux1_x86_64"),
					mkTag("cp27", "none", "manylinux1_x86_64"),
				},
				map[string]string{
					"py_tag":     "cp27-cp27mu-manylinux1_x86_64",
					"py_version": "cp27",
					"py_abi":     "cp27mu",
					"py_arch":    "manylinux1_x86_64",
				},
			},

			{
				"darwin",
				[]*vpython.Environment_Pep425Tag{
					mkTag("cp27", "cp27m", "macosx_10_12_x86_64"),
					mkTag("cp27", "cp27m", "macosx_10_12_fat64"),
					mkTag("cp27", "cp27m", "macosx_10_12_fat32"),
					mkTag("cp27", "cp27m", "macosx_10_12_intel"),
					mkTag("cp27", "cp27m", "macosx_10_10_intel"),
					mkTag("cp27", "cp27m", "macosx_10_9_fat64"),
					mkTag("cp27", "cp27m", "macosx_10_9_fat32"),
					mkTag("cp27", "cp27m", "macosx_10_9_universal"),
					mkTag("cp27", "cp27m", "macosx_10_8_fat32"),
					mkTag("cp27", "cp27m", "macosx_10_8_universal"),
					mkTag("cp27", "cp27m", "macosx_10_6_intel"),
					mkTag("cp27", "cp27m", "macosx_10_6_fat64"),
					mkTag("cp27", "cp27m", "macosx_10_6_fat32"),
					mkTag("cp27", "cp27m", "macosx_10_6_universal"),
					mkTag("cp27", "cp27m", "macosx_10_5_universal"),
					mkTag("cp27", "cp27m", "macosx_10_4_intel"),
					mkTag("cp27", "cp27m", "macosx_10_4_fat32"),
					mkTag("cp27", "cp27m", "macosx_10_1_universal"),
					mkTag("cp27", "cp27m", "macosx_10_0_fat32"),
					mkTag("cp27", "cp27m", "macosx_10_0_universal"),
					mkTag("cp27", "none", "macosx_10_12_x86_64"),
					mkTag("cp27", "none", "macosx_10_12_intel"),
					mkTag("cp27", "none", "macosx_10_12_fat64"),
					mkTag("cp27", "none", "macosx_10_9_universal"),
					mkTag("cp27", "none", "macosx_10_8_x86_64"),
					mkTag("cp27", "none", "macosx_10_8_intel"),
					mkTag("cp27", "none", "macosx_10_7_intel"),
					mkTag("cp27", "none", "macosx_10_7_fat64"),
					mkTag("cp27", "none", "macosx_10_7_fat32"),
					mkTag("cp27", "none", "macosx_10_6_universal"),
					mkTag("cp27", "none", "macosx_10_5_x86_64"),
					mkTag("cp27", "none", "macosx_10_5_intel"),
					mkTag("cp27", "none", "macosx_10_3_fat32"),
					mkTag("cp27", "none", "macosx_10_3_universal"),
					mkTag("cp27", "none", "macosx_10_2_fat32"),
					mkTag("py2", "none", "macosx_10_4_intel"),
					mkTag("cp27", "none", "any"),
				},
				map[string]string{
					"py_tag":     "cp27-cp27m-macosx_10_4_intel",
					"py_version": "cp27",
					"py_abi":     "cp27m",
					"py_arch":    "macosx_10_4_intel",
				},
			},

			{
				"exampleOS",
				[]*vpython.Environment_Pep425Tag{
					mkTag("py27", "none", "any"),
					mkTag("py27", "foo", "bar"),
				},
				map[string]string{
					"py_tag":     "py27-foo-bar",
					"py_version": "py27",
					"py_abi":     "foo",
					"py_arch":    "bar",
				},
			},

			{
				"exampleOS",
				[]*vpython.Environment_Pep425Tag{
					mkTag("py27", "none", "any"),
					mkTag("py27", "none", "linux_386"),
				},
				map[string]string{
					"py_tag":     "py27-none-linux_386",
					"py_version": "py27",
					"py_abi":     "none",
					"py_arch":    "linux_386",
				},
			},

			{
				"exampleOS",
				[]*vpython.Environment_Pep425Tag{
					mkTag("py27", "none", "any"),
					mkTag("py27", "cp27mu", "any"),
				},
				map[string]string{
					"py_tag":     "py27-cp27mu-any",
					"py_version": "py27",
					"py_abi":     "cp27mu",
					"py_arch":    "any",
				},
			},
		} {
			tagsStr := make([]string, len(tc.tags))
			for i, other := range tc.tags {
				tagsStr[i] = other.TagString()
			}

			Convey(fmt.Sprintf(`Generates template for [%s]`, strings.Join(tagsStr, ", ")), func() {
				So(getPEP425CIPDTemplates(tc.goOS, tc.tags), ShouldResemble, tc.template)
			})
		}
	})
}
