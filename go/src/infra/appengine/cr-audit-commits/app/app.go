// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package crauditcommits implements cr-audit-commits.appspot.com services.
package crauditcommits

import (
	"net/http"

	"go.chromium.org/luci/appengine/gaemiddleware"
	"go.chromium.org/luci/server/router"
)

func init() {
	r := router.New()

	// This ensures that the route is only accessible to cron jobs.
	cronmw := gaemiddleware.BaseProd().Extend(gaemiddleware.RequireCron)

	gaemiddleware.InstallHandlers(r)

	r.GET("/_cron/commitscanner", cronmw, CommitScanner)

	http.DefaultServeMux.Handle("/", r)
}
