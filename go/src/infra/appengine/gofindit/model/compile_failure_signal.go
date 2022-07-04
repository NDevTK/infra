// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package model

import (
	"context"
	"infra/appengine/gofindit/util"
)

// Compile Failure Signal represents signal extracted from compile failure log.
type CompileFailureSignal struct {
	Nodes []string
	Edges []*CompileFailureEdge
	// A map of {<file_path>:[lines]} represents failure positions in source file
	Files map[string][]int
	// A map of {<dependency_file_name>:[<list of dependencies>]}. Used to improve
	// the speed when we do dependency analysis
	DependencyMap map[string][]string
}

// CompileFailureEdge represents a failed edge in ninja failure log
type CompileFailureEdge struct {
	Rule         string // Rule is like CXX, CC...
	OutputNodes  []string
	Dependencies []string
}

func (c *CompileFailureSignal) AddLine(filePath string, line int) {
	c.AddFilePath(filePath)
	for _, l := range c.Files[filePath] {
		if l == line {
			return
		}
	}
	c.Files[filePath] = append(c.Files[filePath], line)
}

func (c *CompileFailureSignal) AddFilePath(filePath string) {
	if c.Files == nil {
		c.Files = map[string][]int{}
	}
	_, exist := c.Files[filePath]
	if !exist {
		c.Files[filePath] = []int{}
	}
}

// Put all the dependencies in a map with the form
// {<dependency_file_name>:[<list of dependencies>]}
func (cfs *CompileFailureSignal) CalculateDependencyMap(c context.Context) {
	cfs.DependencyMap = map[string][]string{}
	for _, edge := range cfs.Edges {
		for _, dependency := range edge.Dependencies {
			fileName := util.GetCanonicalFileName(dependency)
			_, ok := cfs.DependencyMap[fileName]
			if !ok {
				cfs.DependencyMap[fileName] = []string{}
			}
			// Check if the dependency already exists
			// Do a for loop, the length is short. So it should be ok.
			exist := false
			for _, d := range cfs.DependencyMap[fileName] {
				if d == dependency {
					exist = true
					break
				}
			}
			if !exist {
				cfs.DependencyMap[fileName] = append(cfs.DependencyMap[fileName], dependency)
			}
		}
	}
}
