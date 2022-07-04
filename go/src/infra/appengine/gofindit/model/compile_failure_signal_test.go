// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package model

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAddLine(t *testing.T) {
	Convey("Add line or file path", t, func() {
		signal := &CompileFailureSignal{}
		signal.AddLine("a/b", 12)
		So(signal.Files, ShouldResemble, map[string][]int{"a/b": {12}})
		signal.AddLine("a/b", 14)
		So(signal.Files, ShouldResemble, map[string][]int{"a/b": {12, 14}})
		signal.AddLine("c/d", 8)
		So(signal.Files, ShouldResemble, map[string][]int{"a/b": {12, 14}, "c/d": {8}})
		signal.AddLine("a/b", 14)
		So(signal.Files, ShouldResemble, map[string][]int{"a/b": {12, 14}, "c/d": {8}})
		signal.AddFilePath("x/y")
		So(signal.Files, ShouldResemble, map[string][]int{"a/b": {12, 14}, "c/d": {8}, "x/y": {}})
		signal.AddFilePath("x/y")
		So(signal.Files, ShouldResemble, map[string][]int{"a/b": {12, 14}, "c/d": {8}, "x/y": {}})
	})
}

func TestCalculateDependencyMap(t *testing.T) {
	Convey("Calculate dependency map", t, func() {
		signal := &CompileFailureSignal{
			Edges: []*CompileFailureEdge{
				{
					Dependencies: []string{
						"x/y/a.h",
						"xx/yy/b.h",
					},
				},
				{
					Dependencies: []string{
						"y/z/a.cc",
						"zz/y/c.yy",
						"x/y/a.h",
					},
				},
			},
		}
		signal.CalculateDependencyMap(context.Background())
		So(signal.DependencyMap, ShouldResemble, map[string][]string{
			"a": {"x/y/a.h", "y/z/a.cc"},
			"b": {"xx/yy/b.h"},
			"c": {"zz/y/c.yy"},
		})
	})
}
