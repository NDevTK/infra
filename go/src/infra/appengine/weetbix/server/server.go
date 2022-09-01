// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package server contains shared server initialisation logic for
// Weetbix services.
package server

import (
	"infra/appengine/weetbix/app"
	"infra/appengine/weetbix/internal/admin"
	adminpb "infra/appengine/weetbix/internal/admin/proto"
	"infra/appengine/weetbix/internal/analysis"
	"infra/appengine/weetbix/internal/analyzedtestvariants"
	"infra/appengine/weetbix/internal/bugs/updater"
	"infra/appengine/weetbix/internal/clustering/reclustering/orchestrator"
	"infra/appengine/weetbix/internal/config"
	"infra/appengine/weetbix/internal/metrics"
	"infra/appengine/weetbix/internal/services/reclustering"
	"infra/appengine/weetbix/internal/services/resultcollector"
	"infra/appengine/weetbix/internal/services/resultingester"
	"infra/appengine/weetbix/internal/services/testvariantbqexporter"
	"infra/appengine/weetbix/internal/services/testvariantupdator"
	"infra/appengine/weetbix/internal/span"
	analysispb "infra/appengine/weetbix/proto/v1"
	"infra/appengine/weetbix/rpc"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/grpc/prpc"
	luciserver "go.chromium.org/luci/server"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/encryptedcookies"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"
	"go.chromium.org/luci/server/secrets"
	spanmodule "go.chromium.org/luci/server/span"
	"go.chromium.org/luci/server/tq"
)

// Main implements the common entrypoint for all LUCI Analysis GAE services.
// All LUCI Analysis GAE services have the code necessary to serve all pRPCs,
// crons and task queues. The only thing that is not shared is frontend
// handling, due to the fact this service requires other assets (javascript,
// files) to be deployed.
//
// Allowing all services to serve everything (except frontend) minimises
// the need to keep server code in sync with changes with dispatch.yaml.
// Moreover, dispatch.yaml changes are not deployed atomically
// with service changes, so this avoids transient traffic rejection during
// rollout of new LUCI Analysis versions that switch handling of endpoints
// between services.
func Main(init func(srv *luciserver.Server) error) {
	// Use the same modules for all LUCI Analysis services.
	modules := []module.Module{
		cfgmodule.NewModuleFromFlags(),
		cron.NewModuleFromFlags(),
		encryptedcookies.NewModuleFromFlags(), // Required for auth sessions.
		gaeemulation.NewModuleFromFlags(),     // Needed by cfgmodule.
		secrets.NewModuleFromFlags(),          // Needed by encryptedcookies.
		spanmodule.NewModuleFromFlags(),
		tq.NewModuleFromFlags(),
	}
	luciserver.Main(nil, modules, func(srv *luciserver.Server) error {
		// Register pPRC servers.
		srv.PRPC.AccessControl = prpc.AllowOriginAll
		srv.PRPC.Authenticator = &auth.Authenticator{
			Methods: []auth.Method{
				&auth.GoogleOAuth2Method{
					Scopes: []string{"https://www.googleapis.com/auth/userinfo.email"},
				},
			},
		}
		// TODO(crbug/1082369): Remove this workaround once field masks can be decoded.
		srv.PRPC.HackFixFieldMasksForJSON = true
		srv.RegisterUnaryServerInterceptor(span.SpannerDefaultsInterceptor())

		ac, err := analysis.NewClient(srv.Context, srv.Options.CloudProject)
		if err != nil {
			return errors.Annotate(err, "creating analysis client").Err()
		}

		analysispb.RegisterClustersServer(srv.PRPC, rpc.NewClustersServer(ac))
		analysispb.RegisterRulesServer(srv.PRPC, rpc.NewRulesSever())
		analysispb.RegisterProjectsServer(srv.PRPC, rpc.NewProjectsServer())
		analysispb.RegisterInitDataGeneratorServer(srv.PRPC, rpc.NewInitDataGeneratorServer())
		analysispb.RegisterTestVariantsServer(srv.PRPC, rpc.NewTestVariantsServer())
		adminpb.RegisterAdminServer(srv.PRPC, admin.CreateServer())
		analysispb.RegisterTestHistoryServer(srv.PRPC, rpc.NewTestHistoryServer())

		// GAE crons.
		updateAnalysisAndBugsHandler := updater.NewHandler(srv.Options.CloudProject, srv.Options.Prod)
		cron.RegisterHandler("update-analysis-and-bugs", updateAnalysisAndBugsHandler.CronHandler)
		cron.RegisterHandler("read-config", config.Update)
		cron.RegisterHandler("export-test-variants", testvariantbqexporter.ScheduleTasks)
		cron.RegisterHandler("purge-test-variants", analyzedtestvariants.Purge)
		cron.RegisterHandler("reclustering", orchestrator.CronHandler)
		cron.RegisterHandler("global-metrics", metrics.GlobalMetrics)

		// Pub/Sub subscription endpoints.
		srv.Routes.POST("/_ah/push-handlers/buildbucket", nil, app.BuildbucketPubSubHandler)
		srv.Routes.POST("/_ah/push-handlers/cvrun", nil, app.CVRunPubSubHandler)

		// Register task queue tasks.
		if err := reclustering.RegisterTaskHandler(srv); err != nil {
			return errors.Annotate(err, "register reclustering").Err()
		}
		if err := resultingester.RegisterTaskHandler(srv); err != nil {
			return errors.Annotate(err, "register result ingester").Err()
		}
		resultcollector.RegisterTaskClass()
		testvariantbqexporter.RegisterTaskClass()
		testvariantupdator.RegisterTaskClass()

		return init(srv)
	})
}
