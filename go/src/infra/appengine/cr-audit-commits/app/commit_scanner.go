// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package crauditcommits implements cr-audit-commits.appspot.com services.
package crauditcommits

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/context"

	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/router"
)

const (
	gitilesScope = "https://www.googleapis.com/auth/gerritcodereview"
)

// CommitScanner is a handler function that gets the list of new commits and
// if they are eithered authored or committed by an account defined in RuleMap
// (see rules.go), records details about them and schedules detailed audits.
//
// It expects the 'repo' parameter containing a url to a valid gitiles-enabled
// git repository and branch.
// e.g. "https://chromium.googlesource.com/infra/infra/+/master"
//
// The handler uses this url as a key to retrieve the state of the last run
// from the datastore and resume the git log from the last known commit.
//
// Returns 200 http status if no errors occur.
func CommitScanner(rc *router.Context) {
	ctx, resp, req := rc.Context, rc.Writer, rc.Request
	repo := req.FormValue("repo")
	// Supported repositories are those present as keys in RuleMap.
	// see rules.go.
	repoRules, hasRules := RuleMap[repo]
	if !hasRules {
		http.Error(resp, fmt.Sprintf("No audit rules defined for %s", repo), 400)
	}
	rev := req.FormValue("starting_revision")
	c := &RepoConfig{RepoURL: repo}
	switch err := datastore.Get(ctx, c); err {
	case datastore.ErrNoSuchEntity:
		if rev != "" {
			c.LastKnownCommit = rev

		} else {
			http.Error(resp, "Missing starting_revision for new repo", 400)
			return
		}
	case nil:
		rev = c.LastKnownCommit
	default:
		http.Error(resp, err.Error(), 500)
		return
	}
	g, err := getGitilesClient(ctx)
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
	parts := strings.Split(repo, "/+/")
	base, branch := parts[0], parts[1]
	fl, err := g.LogForward(ctx, base, rev, branch)
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
	// Defer persisting c.LastKnownCommit and c.LastRelevantCommit in case
	// ScheduleAudit panics.
	defer func() {
		if err := datastore.Put(ctx, c); err != nil {
			logging.WithError(err).Errorf(ctx, "Could not save last known/interesting commits")
		}
	}()
	// TODO(robertocn): Make sure that we break out of this for loop if we
	// reach a deadline of ~5 mins (Since cron job have a 10 minute
	// deadline). Use the context for this.
	for _, commit := range fl {
		for _, rule := range repoRules {
			switch rule.Account {
			case commit.Author.Email, commit.Committer.Email:
				if c.LastRelevantCommit != commit.Commit {
					n, err := saveNewRelevantCommit(ctx, c, commit)
					if err != nil {
						http.Error(resp, err.Error(), 500)
						return
					}
					c.LastRelevantCommit = n.CommitHash
					// If one rule matches, that's
					// enough, move on to the next
					// commit.
					break
				}
			}
		}
		c.LastKnownCommit = commit.Commit
	}
}

func saveNewRelevantCommit(ctx context.Context, cfg *RepoConfig, commit gitiles.Commit) (*RelevantCommit, error) {
	rk, err := datastore.KeyForObjErr(ctx, cfg)

	if err != nil {
		return nil, err
	}

	rc := &RelevantCommit{
		ForRepoConfig:          rk,
		CommitHash:             commit.Commit,
		PreviousRelevantCommit: cfg.LastRelevantCommit,
		Status:                 auditScheduled,
	}

	if err = datastore.Put(ctx, rc); err != nil {
		return nil, err
	}

	return rc, nil
}

// getGitilesClient creates a new gitiles client bound to a new http client
// that is bound to an authenticated transport.
func getGitilesClient(ctx context.Context) (*gitiles.Client, error) {
	t, err := auth.GetRPCTransport(ctx, auth.AsSelf, auth.WithScopes(gitilesScope))
	if err != nil {
		return nil, err
	}
	return &gitiles.Client{Client: &http.Client{Transport: t}}, nil
}
