// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package git

import (
	"bufio"
	"encoding/binary"
	"io"
	"strconv"
	"strings"

	"go.chromium.org/luci/common/errors"
)

// Read reads the graph from r.
// It is the opposite of (*Graph).Write().
func (g *Graph) Read(r *bufio.Reader) error {
	g.ensureInitialized()
	return (&reader{r: r}).readGraph(g)
}

type reader struct {
	r *bufio.Reader

	// textMode means tokens are encoded as utf-8 strings and appear on separate
	// lines.
	textMode bool

	// ordered are the nodes in the order as they appear in the reader.
	ordered []*node
	buf     []byte

	// allEdges is a pre-allocated memory for all edges of all nodes.
	// The range allEdges[:cap(allEdges)] is the allocated memory,
	// whereas range allEdges[len(allEdges):] is the available one.
	allEdges []edge
}

func (r *reader) readGraph(g *Graph) error {
	// Verify header.
	switch header, err := r.readInt(); {
	case err != nil:
		return err
	case header != magicHeader:
		return errors.Reason("unexpected header").Err()
	}

	// Read version.
	switch ver, err := r.readInt(); {
	case err != nil:
		return err
	case ver != 0:
		return errors.Reason("unexpected version %d; expected 0", ver).Err()
	}

	// Read the commit.
	var err error
	if g.Commit, err = r.readString(); err != nil {
		return err
	}

	// Read the nodes.
	r.ordered = r.ordered[:0]
	if err := r.readNode(&g.root); err != nil {
		return errors.Annotate(err, "failed to read nodes").Err()
	}

	// Read the total number of edges.
	totalEdges, err := r.readInt()
	if err != nil {
		return errors.Annotate(err, "failed to read the total number of edges").Err()
	}

	// Allocate a giant slice for all edges that we are about to read.
	r.allEdges = make([]edge, 0, totalEdges)

	// Read the edges.
	for _, n := range r.ordered {
		if err := r.readEdges(n); err != nil {
			return errors.Annotate(err, "failed to read edges").Err()
		}
	}

	return nil
}

func (r *reader) readNode(n *node) error {
	r.ordered = append(r.ordered, n)

	// Read the denominator.
	var err error
	if n.probSumDenominator, err = r.readInt(); err != nil {
		return err
	}

	// Read the number of children.
	childCount, err := r.readInt()
	switch {
	case err != nil:
		return err
	case childCount == 0:
		return nil
	}

	// Read the children.
	n.children = make(map[string]*node, childCount)
	for i := 0; i < childCount; i++ {
		childBaseName, err := r.readString()
		if err != nil {
			return err
		}
		child := &node{parent: n}
		if n.name == "//" {
			child.name = n.name + childBaseName
		} else {
			child.name = n.name + "/" + childBaseName
		}
		n.children[childBaseName] = child
		if err := r.readNode(child); err != nil {
			return err
		}
	}

	return nil
}

func (r *reader) readEdges(n *node) error {
	// Read the number of edges.
	count, err := r.readInt()
	switch {
	case err != nil:
		return err
	case count == 0:
		return nil
	}

	// Allocate the edge slice and shift the cursor.
	n.edges = r.allEdges[len(r.allEdges) : len(r.allEdges)+count]
	n.copyEdgesOnAppend = true
	r.allEdges = r.allEdges[:len(r.allEdges)+count]

	// Read the edges.
	for i := range n.edges {
		switch index, err := r.readInt(); {
		case err != nil:
			return err
		case index < 0 || index >= len(r.ordered):
			return errors.Reason("node index %d is out of bounds", index).Err()
		default:
			n.edges[i].to = r.ordered[index]
		}

		p, err := r.readInt64()
		if err != nil {
			return err
		}
		n.edges[i].probSum = probability(p)
	}
	return nil
}

func (r *reader) readString() (string, error) {
	if r.textMode {
		return r.readLine()
	}

	length, err := r.readInt()
	if err != nil {
		return "", err
	}

	if cap(r.buf) < length {
		r.buf = make([]byte, length)
	}
	r.buf = r.buf[:length]
	if _, err := io.ReadFull(r.r, r.buf); err != nil {
		return "", err
	}
	return string(r.buf), nil
}

func (r *reader) readInt() (int, error) {
	n, err := r.readInt64()
	return int(n), err
}

func (r *reader) readInt64() (int64, error) {
	if r.textMode {
		s, err := r.readLine()
		if err != nil {
			return 0, err
		}
		return strconv.ParseInt(s, 10, 64)
	}

	n, err := binary.ReadVarint(r.r)
	return n, err
}

func (r *reader) readLine() (string, error) {
	if !r.textMode {
		panic("not text mode")
	}
	s, err := r.r.ReadString('\n')
	return strings.TrimSuffix(s, "\n"), err
}
