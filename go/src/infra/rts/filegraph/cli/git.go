// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cli

import (
	"context"
	"flag"
	"path/filepath"
	"strings"

	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"

	"infra/rts/filegraph"
	"infra/rts/filegraph/git"
	"infra/rts/internal/gitutil"
)

// gitGraph loads a file graph from a git log.
type gitGraph struct {
	opt         git.LoadOptions
	edgeReader  git.EdgeReader
	maxDistance float64
	*git.Graph
}

func (g *gitGraph) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&g.opt.Ref, "ref", "refs/heads/main", text.Doc(`
		Load the file graph for this git ref.
		For refs/heads/main, refs/heads/master is read if main doesn't exist.
	`))
	fs.IntVar(&g.opt.MaxCommitSize, "max-commit-size", 100, text.Doc(`
		Maximum number of files touched by a commit.
		Commits that exceed this limit are ignored.
		The rationale is that large commits provide a weak signal of file
		relatedness and are expensive to process, O(N^2).
	`))
	fs.Float64Var(&g.maxDistance, "max-distance", 0, text.Doc(`
		If positive, the distance threshold. Nodes further than this are considered
		unreachable.
	`))
}

func (g *gitGraph) query(sources ...filegraph.Node) *filegraph.Query {
	return &filegraph.Query{
		Sources:     sources,
		EdgeReader:  &g.edgeReader,
		MaxDistance: g.maxDistance,
	}
}

func (g *gitGraph) Validate() error {
	if !strings.HasPrefix(g.opt.Ref, "refs/") {
		return errors.Reason("-ref %q doesn't start with refs/", g.opt.Ref).Err()
	}
	if g.opt.MaxCommitSize < 0 {
		return errors.Reason("-max-commit-size must be non-negative").Err()
	}
	return nil
}

// loadSyncedNodes calls loadSyncedGraph for filePaths' repo, and then loads a
// node for each of the files.
func (g *gitGraph) loadSyncedNodes(ctx context.Context, filePaths ...string) ([]filegraph.Node, error) {
	repoDir, err := gitutil.EnsureSameRepo(filePaths...)
	if err != nil {
		return nil, err
	}

	// Load the graph.
	if g.Graph, err = git.Load(ctx, repoDir, g.opt); err != nil {
		return nil, err
	}

	// Load the nodes.
	nodes := make([]filegraph.Node, len(filePaths))
	for i, f := range filePaths {
		// Convert the filename to a node name.
		if f, err = filepath.Abs(f); err != nil {
			return nil, err
		}
		name, err := filepath.Rel(repoDir, f)
		if err != nil {
			return nil, err
		}
		name = filepath.ToSlash(name)
		switch {
		case name == ".":
			name = "//" // the root
		case strings.HasPrefix(name, "/") || strings.HasPrefix(name, "../") || strings.HasPrefix(name, "./"):
			return nil, errors.Reason("unexpected path %q", name).Err()
		default:
			name = "//" + name
		}

		// Load the node.
		node := g.Node(name)
		if node == nil {
			return nil, errors.Reason("node %q not found", name).Err()
		}
		nodes[i] = node
	}

	return nodes, nil
}
