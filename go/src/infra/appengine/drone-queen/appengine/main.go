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

package main

import (
	"context"
	"os"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	icron "infra/appengine/drone-queen/internal/cron"
	"infra/appengine/drone-queen/internal/frontend"
	"infra/appengine/drone-queen/internal/middleware"
)

func main() {
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if exp, err := texporter.New(texporter.WithProjectID(projectID)); err == nil {
		tp := trace.NewTracerProvider(
			trace.WithBatcher(exp),
			trace.WithResource(newResource(ctx)),
		)
		// The TracerProvider is shut down as a server cleanup
		// action below rather than deferred here, because
		// server.Main calls exit() directly (meaning defers
		// don't run).
		otel.SetTracerProvider(tp)
	} else {
		logging.Infof(ctx, "Error setting up trace exporter: %s", err)
	}

	modules := []module.Module{
		gaeemulation.NewModuleFromFlags(),
		cron.NewModuleFromFlags(),
		cfgmodule.NewModuleFromFlags(),
	}
	server.Main(nil, modules, func(srv *server.Server) error {
		icron.InstallHandlers()
		srv.RegisterUnaryServerInterceptor(middleware.UnaryTrace)
		frontend.RegisterServers(srv)
		srv.RegisterCleanup(func(ctx context.Context) {
			tp := otel.GetTracerProvider()
			if tp2, ok := tp.(*trace.TracerProvider); ok {
				if err := tp2.Shutdown(ctx); err != nil {
					logging.Infof(ctx, "Error shutting down TracerProvider: %s", err)
				}
			} else {
				logging.Infof(ctx, "Unexpected type when shutting down TracerProvider: %T", tp)
			}
		})
		return nil
	})
}

func newResource(ctx context.Context) *resource.Resource {
	// This should never error.
	// Even if it does, try to keep running normally.
	r, _ := resource.New(
		ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("drone-queen"),
		),
	)
	return r
}
