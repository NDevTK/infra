// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cron implements handlers for appengine cron targets in this app.
//
// All actual logic related to fleet management should be implemented in the
// main fleet API. These handlers should only encapsulate the following bits of
// logic:
//
//   - Calling other API as the appengine service account user.
//   - Translating luci-config driven admin task parameters.
package cron

import (
	"net/http"
	"time"

	"go.chromium.org/luci/appengine/gaemiddleware"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/router"

	"infra/appengine/drone-queen/internal/config"
	"infra/appengine/drone-queen/internal/middleware"
	"infra/appengine/drone-queen/internal/queries"
)

// InstallHandlers installs handlers for cron jobs that are part of this app.
//
// All handlers serve paths under /internal/cron/*
// These handlers can only be called by appengine's cron service.
func InstallHandlers(srv *server.Server) {
	mw := router.NewMiddlewareChain(gaemiddleware.RequireCron, middleware.Trace)
	srv.Routes.GET("/internal/cron/import-service-config", mw, errHandler(importServiceConfig))
	srv.Routes.GET("/internal/cron/free-invalid-duts", mw, errHandler(freeInvalidDUTs))
	srv.Routes.GET("/internal/cron/prune-expired-drones", mw, errHandler(pruneExpiredDrones))
	srv.Routes.GET("/internal/cron/prune-drained-duts", mw, errHandler(pruneDrainedDUTs))
}

// errHandler wraps a handler function that returns errors.
func errHandler(f func(*router.Context) error) router.Handler {
	return func(c *router.Context) {
		if err := f(c); err != nil {
			logging.Errorf(c.Context, "handler returned error: %s", err)
			http.Error(c.Writer, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func importServiceConfig(c *router.Context) error {
	return config.Import(c.Context)
}

func freeInvalidDUTs(c *router.Context) (err error) {
	defer func() {
		freeInvalidDUTsTick.Add(c.Context, 1, config.Instance(c.Context), err == nil)
	}()
	return queries.FreeInvalidDUTs(c.Context, time.Now())
}

func pruneExpiredDrones(c *router.Context) (err error) {
	defer func() {
		pruneExpiredDronesTick.Add(c.Context, 1, config.Instance(c.Context), err == nil)
	}()
	return queries.PruneExpiredDrones(c.Context, time.Now())
}

func pruneDrainedDUTs(c *router.Context) (err error) {
	defer func() {
		pruneDrainedDUTsTick.Add(c.Context, 1, config.Instance(c.Context), err == nil)
	}()
	return queries.PruneDrainedDUTs(c.Context)
}
