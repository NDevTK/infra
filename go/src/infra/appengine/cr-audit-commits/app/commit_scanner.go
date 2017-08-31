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
// It expects the 'repo' parameter containing the name of a configured repo
// e.g. "chromium-src-master"
//
// The handler uses this url as a key to retrieve the state of the last run
// from the datastore and resume the git log from the last known commit.
//
// Returns 200 http status if no errors occur.
func CommitScanner(rc *router.Context) {
	ctx, resp, req := rc.Context, rc.Writer, rc.Request
	repo := req.FormValue("repo")
	rev := ""
	// Supported repositories are those present as keys in RuleMap.
	// see rules_config.go.
	repoRuleSets, hasRuleSets := RuleMap[repo]
	if !hasRuleSets {
		http.Error(resp, fmt.Sprintf("No audit rules defined for %s", repo), 400)
		return
	}
	c := &RepoConfig{Name: repo}
	switch err := datastore.Get(ctx, c); err {
	case datastore.ErrNoSuchEntity:
		http.Error(resp, fmt.Sprintf("The specified repository %s is not configured", repo), 400)
		return
	case nil:
		rev = c.LastKnownCommit
		if rev == "" {
			rev = c.StartingCommit
		}
		if rev == "" {
			http.Error(resp, fmt.Sprintf("The specified repository %s is missing a starting revision", repo), 400)
			return
		}
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
	// TODO(robertocn): Make sure that we break out of this for loop if we
	// reach a deadline of ~5 mins (Since cron job have a 10 minute
	// deadline). Use the context for this.
	for _, commit := range fl {
		for _, ruleSet := range repoRuleSets {
			if ruleSet.Matches(commit) {
				n, err := saveNewRelevantCommit(ctx, c, commit)
				if err != nil {
					http.Error(resp, err.Error(), 500)
					return
				}
				c.LastRelevantCommit = n.CommitHash
				// If the commit matches one ruleSet that's
				// enough. Break to move on to the next commit.
				break
			}
		}
		c.LastKnownCommit = commit.Commit
	}
	if err := datastore.Put(ctx, c); err != nil {
		logging.WithError(err).Errorf(ctx, "Could not save last known/interesting commits")
	}
}

func saveNewRelevantCommit(ctx context.Context, cfg *RepoConfig, commit gitiles.Commit) (*RelevantCommit, error) {
	rk, err := datastore.KeyForObjErr(ctx, cfg)

	if err != nil {
		return nil, err
	}

	rc := &RelevantCommit{
		RepoConfigKey:          rk,
		CommitHash:             commit.Commit,
		PreviousRelevantCommit: cfg.LastRelevantCommit,
		Status:                 auditScheduled,
	}

	if err = datastore.Put(ctx, rc, cfg); err != nil {
		return nil, err
	}

	return rc, nil
}

// getGitilesClient creates a new gitiles client bound to a new http client
// that is bound to an authenticated transport.
func getGitilesClient(ctx context.Context) (*gitiles.Client, error) {
	httpClient, err := getAuthenticatedHTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	return &gitiles.Client{Client: httpClient}, nil
}

func getAuthenticatedHTTPClient(ctx context.Context) (*http.Client, error) {
	t, err := auth.GetRPCTransport(ctx, auth.AsSelf, auth.WithScopes(gitilesScope))
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: t}, nil
}
