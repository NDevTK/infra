// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package crauditcommits implements cr-audit-commits.appspot.com services.
package crauditcommits

import (
	"fmt"
	"net/http"

	"go.chromium.org/luci/common/api/gerrit"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/router"

	"infra/appengine/cr-audit-commits/buildstatus"
)

// SmokeTest is a handler that makes sure that the application can talk to the
// external services it needs.
func SmokeTest(rc *router.Context) {
	ctx, resp := rc.Context, rc.Writer

	// Gitiles
	g, err := getGitilesClient(ctx)
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Could not get a gitiles client")
		http.Error(resp, err.Error(), 500)
		return
	}
	base := "https://chromium.googlesource.com/chromium/src.git"
	branch := "master"
	c, err := g.Log(ctx, base, branch, 1)
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Failed to get log for branch %s in repo %s", branch, base)
		http.Error(resp, err.Error(), 500)
		return
	}
	logging.Infof(ctx, "Successfully obtained details about commit %s from gitiles\n\n", c[0].Commit)

	// Milo
	m, err := buildstatus.GetBuildbotClient(ctx, auth.AsSelf)
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Failed to get BuildbotClient for querying Milo")
		http.Error(resp, err.Error(), 500)
		return
	}
	buildURL := "https://luci-milo.appspot.com/buildbot/chromium.linux/Android%20Builder/86716"
	b, err := buildstatus.GetBuildInfo(ctx, m, buildURL)
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Failed to get details of build %q", buildURL)
		http.Error(resp, err.Error(), 500)
		return
	}
	logging.Infof(ctx, "Successfully obtained %d steps from milo build %s/%s/%d\n\n", len(b.Steps), b.Master, b.BuilderName, b.Number)

	// Gerrit
	ge, err := gerrit.NewClient(g.Client, "https://chromium-review.googlesource.com/")
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Could not create a new gerrit client")
		http.Error(resp, err.Error(), 500)
		return
	}
	clNum := "630300"
	cls, _, err := ge.Query(ctx, gerrit.QueryRequest{Query: clNum})
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Failed to get details of CL %s", clNum)
		http.Error(resp, err.Error(), 500)
		return
	}
	logging.Infof(ctx, "Successfully obtained change %s with subject \"%s\" from gerrit\n\n", cls[0].ChangeID, cls[0].Subject)
	fmt.Fprintf(resp, "Smoke test successful. Examine log for details.")

}
