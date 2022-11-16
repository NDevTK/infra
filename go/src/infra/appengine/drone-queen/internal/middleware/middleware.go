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

	"go.chromium.org/luci/server/router"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

var tracer trace.Tracer

func init() {
	tracer = otel.Tracer("infra/appengine/drone-queen/internal/middleware")
}

var _ grpc.UnaryServerInterceptor = UnaryTrace

// UnaryTrace is a gRPC interceptor for adding trace instrumentation.
// This potentially allows access to gRPC specific metadata that is
// not exposed by the LUCI middleware API.
func UnaryTrace(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()
	return handler(ctx, req)
}

var _ router.Middleware = Trace

// Trace is LUCI middleware for adding trace instrumentation.
func Trace(c *router.Context, next router.Handler) {
	var span trace.Span
	c.Context, span = tracer.Start(c.Context, c.HandlerPath, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()
	next(c)
}
