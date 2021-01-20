// Copyright 2020 The Chromium Authors. All rights reserved.
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
	runRTSQuery(s.Graph, changedFiles, func(nodeName string, af rts.Affectedness) bool {
		if af.Rank <= s.Threshold.Rank || af.Distance <= s.Threshold.Distance {
			// This file too close to skip it.
			return true
		}
		return skipFile(nodeName)
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

	affectedness := make(map[string]rts.Affectedness, len(in.TestVariants))
	for _, tv := range in.TestVariants {
		if tv.FileName != "" {
			affectedness[tv.FileName] = rts.Affectedness{Distance: math.Inf(1), Rank: math.MaxInt32}
		}
	}

	found := 0
	runRTSQuery(g, changedFiles, func(nodeName string, af rts.Affectedness) (keepGoing bool) {
		if _, ok := affectedness[nodeName]; ok {
			affectedness[nodeName] = af
			found++
		}
		return found < len(affectedness)
	})

	for i, tv := range in.TestVariants {
		out.TestVariantAffectedness[i] = affectedness[tv.FileName]
	}
	return nil
}

type rtsCallback func(nodeName string, af rts.Affectedness) (keepGoing bool)

// runRTSQuery walks the file graph from the changed files, along reversed edges
// and calls back for each found file.
func runRTSQuery(g *Graph, changedFiles []string, callback rtsCallback) {
	q := &filegraph.Query{
		Sources: make([]filegraph.Node, 0, len(changedFiles)),
		EdgeReader: &EdgeReader{
			// We run the query from changed files, but we need distance
			// from test files to changed files, and not the other way around.
			Reversed: true,
		},
	}

	for _, f := range changedFiles {
		n := g.Node(f)
		if n != nil {
			q.Sources = append(q.Sources, n)
		} else {
			callback(f, rts.Affectedness{}) // Very affected
		}
	}

	rank := 0
	q.Run(func(sp *filegraph.ShortestPath) (keepGoing bool) {
		// Note: the files are enumerated in the order of distance.
		rank++
		return callback(sp.Node.Name(), rts.Affectedness{Distance: sp.Distance, Rank: rank})
	})
	return
}
