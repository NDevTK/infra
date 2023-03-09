// Copyright 2022 The LUCI Authors.
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

// Package middleware implements shared LUCI middleware and gRPC interceptors.
package middleware

import (
	"context"

	"go.chromium.org/luci/server/cron"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Name used for OpenTelemetry tracers.
const tname = "infra/appengine/drone-queen/internal/middleware"

// A CronWrapper is a function that wraps a cron handler to provide
// middleware functionality.
type CronWrapper func(cron.Handler) cron.Handler

// CronTrace is a wrapper to add tracing to cron handlers.
// name is the name of the trace span.
func CronTrace(name string) CronWrapper {
	return func(h cron.Handler) cron.Handler {
		return func(ctx context.Context) error {
			var span trace.Span
			ctx, span = otel.Tracer(tname).Start(ctx, name, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()
			err := h(ctx)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return err
		}
	}
}
