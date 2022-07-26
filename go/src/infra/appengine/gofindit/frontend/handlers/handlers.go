// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// packahe handlers contains the data structures and functions used for serving
// GoFindit HTTP routes, such as the GoFindit frontend
package handlers

import (
	"go.chromium.org/luci/server/router"
	"go.chromium.org/luci/server/templates"
)

// IndexPage serves a GET request for the index page for the frontend
func IndexPage(ctx *router.Context) {
	templates.MustRender(ctx.Context, ctx.Writer, "pages/index.html", templates.Args{})
}
