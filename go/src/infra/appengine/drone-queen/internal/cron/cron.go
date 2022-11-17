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
	"context"
	"time"

	"go.chromium.org/luci/server/cron"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"infra/appengine/drone-queen/internal/config"
	"infra/appengine/drone-queen/internal/queries"
)

var tracer = otel.Tracer("infra/appengine/drone-queen/internal/cron")

// InstallHandlers installs global handlers for cron jobs that are part of this app.
func InstallHandlers() {
	install := func(id string, handler cron.Handler) {
		cron.RegisterHandler(id, wrap(handler, id))
	}
	install("import-service-config", importServiceConfig)
	install("free-invalid-duts", freeInvalidDUTs)
	install("prune-expired-drones", pruneExpiredDrones)
	install("prune-drained-duts", pruneDrainedDUTs)
}

// wrap wraps cron handlers (basically providing "middleware").
func wrap(f cron.Handler, name string) cron.Handler {
	return func(ctx context.Context) error {
		var span trace.Span
		ctx, span = tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		return f(ctx)
	}
}

func importServiceConfig(ctx context.Context) error {
	return config.Import(ctx)
}

func freeInvalidDUTs(ctx context.Context) (err error) {
	defer func() {
		freeInvalidDUTsTick.Add(ctx, 1, config.Instance(ctx), err == nil)
	}()
	return queries.FreeInvalidDUTs(ctx, time.Now())
}

func pruneExpiredDrones(ctx context.Context) (err error) {
	defer func() {
		pruneExpiredDronesTick.Add(ctx, 1, config.Instance(ctx), err == nil)
	}()
	return queries.PruneExpiredDrones(ctx, time.Now())
}

func pruneDrainedDUTs(ctx context.Context) (err error) {
	defer func() {
		pruneDrainedDUTsTick.Add(ctx, 1, config.Instance(ctx), err == nil)
	}()
	return queries.PruneDrainedDUTs(ctx)
}
