// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handlers

import (
	"go.chromium.org/luci/server/router"
	"go.chromium.org/luci/server/templates"
)

// IndexPage serves a GET request for the index page.
func (h *Handlers) IndexPage(ctx *router.Context) {
	templates.MustRender(ctx.Request.Context(), ctx.Writer, "pages/index.html", templates.Args{})
}
