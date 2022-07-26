// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/router"
)

func respondWithJSON(ctx *router.Context, data interface{}) {
	bytes, err := json.Marshal(data)
	if err != nil {
		logging.Errorf(ctx.Context, "Marshalling JSON for response: %s", err)
		http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
		return
	}

	ctx.Writer.Header().Add("Content-Type", "application/json")
	if _, err := ctx.Writer.Write(bytes); err != nil {
		logging.Errorf(ctx.Context, "Writing JSON response: %s", err)
	}
}
