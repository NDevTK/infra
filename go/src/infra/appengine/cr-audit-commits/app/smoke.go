// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package crauditcommits implements cr-audit-commits.appspot.com services.
package crauditcommits

import (
	"fmt"
	"net/http"

	"go.chromium.org/luci/common/api/gerrit"
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
		http.Error(resp, err.Error(), 500)
		return
	}
	base := "https://chromium.googlesource.com/chromium/src.git"
	branch := "master"
	c, err := g.Log(ctx, base, branch, 1)
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
	fmt.Fprintf(resp, "Successfully obtained details about commit %s from gitiles\n\n", c[0].Commit)

	// Milo
	m, err := buildstatus.GetBuildbotClient(ctx, auth.AsSelf)
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
	b, err := buildstatus.GetBuildInfo(ctx, m,
		"https://build.chromium.org/p/chromium.webkit/builders/WebKit%20Mac10.12/builds/5696")
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
	fmt.Fprintf(resp, "Successfully obtained %d steps from milo build %s/%s/%d\n\n", len(b.Steps), b.Master, b.BuilderName, b.Number)

	// Gerrit
	ge, err := gerrit.NewClient(g.Client, "https://chromium-review.googlesource.com/")
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
	cls, _, err := ge.Query(ctx, gerrit.QueryRequest{Query: "630300"})
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
	fmt.Fprintf(resp, "Successfully obtained change %s with subject \"%s\" from gerrit\n\n", cls[0].ChangeID, cls[0].Subject)

}
