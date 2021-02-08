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

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/logging"
)

// SelectionStrategy implements a selection strategy based on a git graph.
type SelectionStrategy struct {
	Graph *Graph
	EdgeReader

	// Threshold decides whether a test is to be selected: if it is closer or
	// equal than distance, then it is selected. Otherwise, skipped.
	Threshold rts.Affectedness
}

// Select calls skipTestFile for each test file that should be skipped.
func (s *SelectionStrategy) Select(changedFiles []string, skipFile func(name string) (keepGoing bool)) {
	runRTSQuery(s.Graph, &s.EdgeReader, changedFiles, func(name string, af rts.Affectedness) bool {
		if af.Distance <= s.Threshold.Distance {
			// This file too close to skip it.
			return true
		}
		return skipFile(name)
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
func (g *Graph) EvalStrategy(er *EdgeReader) eval.Strategy {
	return func(ctx context.Context, in eval.Input, out *eval.Output) error {
		changedFiles := make([]string, len(in.ChangedFiles))
		changedFileSet := stringset.New(len(in.ChangedFiles))
		for i, f := range in.ChangedFiles {
			changedFiles[i] = f.Path
			changedFileSet.Add(f.Path)
		}

		affectedness := make(map[string]rts.Affectedness, len(in.TestVariants))
		for _, tv := range in.TestVariants {
			// If the test file is in the graph, then by default it is not affected.
			if g.node(tv.FileName) != nil {
				affectedness[tv.FileName] = rts.Affectedness{Distance: math.Inf(1)}
			} else if tv.FileName != "" && !changedFileSet.Has(tv.FileName) {
				// This file is not new and yet the filegraph doesn't have it.
				// This might mean that the filegraph is incomplete/stale
				// or that the reported test file name is incorrect (data bug).
				logging.Warningf(ctx, "test file not found: %s", tv.FileName)
			}
		}

		found := 0
		runRTSQuery(g, er, changedFiles, func(name string, af rts.Affectedness) (keepGoing bool) {
			if _, ok := affectedness[name]; ok {
				affectedness[name] = af
				found++
			}
			return found < len(affectedness)
		})

		for i, tv := range in.TestVariants {
			// If tv.FileName is empty (not in the map), then zero value is used
			// which means very affected.
			out.TestVariantAffectedness[i] = affectedness[tv.FileName]
		}
		return nil
	}
}

type rtsCallback func(name string, af rts.Affectedness) (keepGoing bool)

// runRTSQuery walks the file graph from the changed files, along reversed edges
// and calls back for each found file.
// If a changed file is not in the graph, then it is treated as very affected.
func runRTSQuery(g *Graph, er *EdgeReader, changedFiles []string, callback rtsCallback) {
	q := &filegraph.Query{
		Sources:    make([]filegraph.Node, 0, len(changedFiles)),
		EdgeReader: er,
	}

	for _, f := range changedFiles {
		if n := g.Node(f); n != nil {
			// If the node exists, then include it in the Dijkstra walk.
			q.Sources = append(q.Sources, n)
		} else {
			// Otherwise assume the file is new and treat it as very affected.
			callback(f, rts.Affectedness{})
		}
	}

	q.Run(func(sp *filegraph.ShortestPath) (keepGoing bool) {
		return callback(sp.Node.Name(), rts.Affectedness{Distance: sp.Distance})
	})
}
