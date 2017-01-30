// Copyright 2014 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gitiles

import (
	"io"
	"net/http"
	"strings"

	"github.com/luci/luci-go/common/lru"
)

const (
	defaultCommitCacheSize = 1000
)

// Types ///////////////////////////////////////////////////////////////////////

// Gitiles wraps one Gitiles server. Use NewGitiles to create one.
type Gitiles struct {
	url       string
	requests  chan<- request
	client    *http.Client
	commitLRU *lru.Cache
}

// TreeDiff represents the diff of a single file for a single commit.
// A value of /dev/null means the file was deleted.
type TreeDiff struct {
	Type    string `json:"type"`
	OldID   string `json:"old_id"`
	OldMode int    `json:"old_mode"`
	OldPath string `json:"old_path"`
	NewID   string `json:"new_id"`
	NewMode int    `json:"new_mode"`
	NewPath string `json:"new_path"`
}

// CommitUser is a user involved with this change.
type CommitUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Time  string `json:"time"`
}

// Commit is a git commit.
// Note that not all fields will be filled out from every API call.
type Commit struct {
	Commit    string     `json:"commit"`
	Tree      string     `json:"tree"`
	Parents   []string   `json:"parents"`
	Author    CommitUser `json:"author"`
	Committer CommitUser `json:"committer"`
	Message   string     `json:"message"`
	TreeDiff  []TreeDiff `json:"tree_diff"`
}

// Constructors  ///////////////////////////////////////////////////////////////

// NewGitiles creates a new Gitiles instance for the given |url|, and a maximum
// number of concurrent requests.
func NewGitiles(url string, maxConnections int, client *http.Client) *Gitiles {
	// TODO(iannucci): have a way to destroy the Gitiles instance?
	if client == nil {
		client = http.DefaultClient
	}

	requestChan := make(chan request, maxConnections)
	ret := &Gitiles{
		url:       strings.TrimRight(url, "/"),
		requests:  requestChan,
		client:    client,
		commitLRU: lru.New(defaultCommitCacheSize),
	}
	for i := 0; i < maxConnections; i++ {
		go ret.requestProcessor(requestChan)
	}
	return ret
}

// Member functions ////////////////////////////////////////////////////////////

// URL returns the base url for this Gitiles service wrapper
func (g *Gitiles) URL() string { return g.url }

// JSON Returns an instance of target or an error.
//
// Example:
//   data := map[string]int{}
//   result, err := g.JSON(data, "some", "url", "pieces")
//   if err != nil { panic(err) }
//   data = *result.(*map[string]int)
func (g *Gitiles) JSON(fact typeFactory, pieces ...string) (interface{}, error) {
	reply := make(chan jsonResult, 1)

	g.requests <- jsonRequest{
		strings.Join(pieces, "/"),
		fact,
		reply,
	}
	rslt := <-reply

	return rslt.data, rslt.err
}

type logDecoding struct {
	Logs []*Commit `json:"log"`
}

// GetLog returns the last X commits on the given branch.
// X is server specific. For chromium gitiles, X=100.
// If no branch is given, master is assumed.
func (g *Gitiles) GetLog(branch string) ([]*Commit, error) {
	if branch == "" {
		branch = "master"
	}

	rslt, err := g.JSON(func() interface{} {
		return &logDecoding{}
	}, "+log", branch)

	if err != nil {
		return nil, err
	}

	diff := rslt.(*logDecoding)
	return diff.Logs, nil
}

// GetCommit returns the Commit from the given committish.
func (g *Gitiles) GetCommit(committish string) (*Commit, error) {
	if val := g.commitLRU.Get(committish); val != nil {
		return val.(*Commit), nil
	}

	rslt, err := g.JSON(func() interface{} {
		return &Commit{}
	}, "+", committish)

	if err != nil {
		return nil, err
	}

	diff := rslt.(*Commit)
	g.commitLRU.Put(committish, diff)

	return diff, nil
}

// Private /////////////////////////////////////////////////////////////////////

type request interface {
	Process(rsp *http.Response, err error)
	Method() string
	URLPath() string
	Body() io.Reader
}

func (g *Gitiles) requestProcessor(queue <-chan request) {
	// Launched as a goroutine to avoid blocking the request processor.
	for r := range queue {
		func() {
			if r == nil {
				return
			}

			var req *http.Request
			req, err := http.NewRequest(r.Method(), g.url+"/"+r.URLPath(), r.Body())
			if err != nil {
				r.Process(nil, err)
				return
			}

			rsp, err := g.client.Do(req)
			if rsp.Body != nil {
				defer rsp.Body.Close()
			}

			if err != nil {
				r.Process(nil, err)
				return
			}

			e := StatusError(rsp.StatusCode)
			if e.Bad() {
				err = e
			}
			r.Process(rsp, err)
		}()
	}
}
