// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package crauditcommits implements cr-audit-commits.appspot.com services.
package crauditcommits

import (
	"net/http"

	"go.chromium.org/luci/appengine/gaeauth/server"
	"go.chromium.org/luci/appengine/gaemiddleware"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/router"
	"go.chromium.org/luci/server/templates"
)

func init() {
	r := router.New()

	// This does not require auth.
	templatesmw := gaemiddleware.BaseProd().Extend(getTemplatesMW())
	// This ensures that the route is only accessible to cron jobs.
	cronmw := gaemiddleware.BaseProd().Extend(gaemiddleware.RequireCron)
	// This requires authentication.
	authmw := getAuthMW()

	gaemiddleware.InstallHandlers(r)

	r.GET("/", templatesmw, index)
	r.GET("/_cron/commitscanner", cronmw, CommitScanner)
	r.GET("/admin/smoketest", authmw, SmokeTest)

	http.DefaultServeMux.Handle("/", r)
}

// Handler for the index page.
func index(rc *router.Context) {
	templates.MustRender(rc.Context, rc.Writer, "pages/index.html", templates.Args{})
}

// Requires authentication.
func getAuthMW() router.MiddlewareChain {
	a := auth.Authenticator{
		Methods: []auth.Method{
			&server.OAuth2Method{Scopes: []string{server.EmailScope}},
			&server.InboundAppIDAuthMethod{},
			server.CookieAuth,
		},
	}
	return gaemiddleware.BaseProd().Extend(a.GetMiddleware())
}

// Lets the templates lib know where to load templates from.
func getTemplatesMW() router.Middleware {
	return templates.WithTemplates(&templates.Bundle{
		Loader: templates.FileSystemLoader("templates"),
	})
}
