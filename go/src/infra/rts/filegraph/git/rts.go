// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package git

import (
	"context"
	"math"

	"infra/rts"
	"infra/rts/filegraph"
	"infra/rts/presubmit/eval"
)

// SelectionStrategy implements a selection strategy based on a git graph.
type SelectionStrategy struct {
	Graph *Graph

	// Threshold decides whether a test is to be selected: if it is closer or
	// equal than distance OR rank, then it is selected. Otherwise, skipped.
	Threshold rts.Affectedness
}

// Select calls skipTestFile for each test file that should be skipped.
func (s *SelectionStrategy) Select(changedFiles []string, skipFile func(name string) (keepGoing bool)) {
	runRTSQuery(s.Graph, changedFiles, func(sp *filegraph.ShortestPath, rank int) bool {
		if rank <= s.Threshold.Rank || sp.Distance <= s.Threshold.Distance {
			// This file too close to skip it.
			return true
		}
		return skipFile(sp.Node.Name())
	})
}

// EvalStrategy implements eval.Strategy. It can be used to evaluate data
// quality of the graph.
//
// This function has minimal input validation. It is not designed to be called
// by the evaluation framework directly. Instead it should be wrapped with
// another strategy function that does the validation. In particular, this
// function does not check in.ChangedFiles[i].Repo and does not check for file
// patterns that must be exempted from RTS.
func (g *Graph) EvalStrategy(ctx context.Context, in eval.Input, out *eval.Output) error {
	changedFiles := make([]string, len(in.ChangedFiles))
	for i, f := range in.ChangedFiles {
		changedFiles[i] = f.Path
	}

	affectedness := make(map[filegraph.Node]rts.Affectedness, len(in.TestVariants))
	testNodes := make([]filegraph.Node, len(in.TestVariants))
	for i, tv := range in.TestVariants {
		if n := g.Node(tv.FileName); n != nil {
			// Too far away by default.
			affectedness[n] = rts.Affectedness{Distance: math.Inf(1), Rank: math.MaxInt32}
			testNodes[i] = n
		}
	}

	found := 0
	runAllTests := runRTSQuery(g, changedFiles, func(sp *filegraph.ShortestPath, rank int) (keepGoing bool) {
		if _, ok := affectedness[sp.Node]; ok {
			affectedness[sp.Node] = rts.Affectedness{Distance: sp.Distance, Rank: rank}
			found++
		}
		return found < len(affectedness)
	})
	if runAllTests {
		return nil
	}

	for i, n := range testNodes {
		if n != nil {
			out.TestVariantAffectedness[i] = affectedness[n]
		}
	}
	return nil
}

type rtsCallback func(sp *filegraph.ShortestPath, rank int) (keepGoing bool)

// runRTSQuery walks the file graph from the changed files, along reversed edges
// and calls back for each found file.
func runRTSQuery(g *Graph, changedFiles []string, callback rtsCallback) (runAllTests bool) {
	q := &filegraph.Query{
		Sources: make([]filegraph.Node, len(changedFiles)),
		EdgeReader: &EdgeReader{
			// We run the query from changed files, but we need distance
			// from test files to changed files, and not the other way around.
			Reversed: true,
		},
	}

	for i, f := range changedFiles {
		if q.Sources[i] = g.Node(f); q.Sources[i] == nil {
			// TODO(nodir): consider not bailing.
			// We are bailing on all CLs where at least one file is new.
			runAllTests = true
			return
		}
	}

	rank := 0
	q.Run(func(sp *filegraph.ShortestPath) (keepGoing bool) {
		// Note: the files are enumerated in the order of distance.
		rank++
		return callback(sp, rank)
	})
	return
}
