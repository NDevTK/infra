// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go.chromium.org/luci/auth/identity"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/encryptedcookies"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"
	"go.chromium.org/luci/server/router"
	"go.chromium.org/luci/server/secrets"
	spanmodule "go.chromium.org/luci/server/span"
	"go.chromium.org/luci/server/templates"
	"go.chromium.org/luci/server/tq"
	"google.golang.org/appengine/log"

	// Store auth sessions in the datastore.
	_ "go.chromium.org/luci/server/encryptedcookies/session/datastore"

	"infra/appengine/weetbix/app"
	"infra/appengine/weetbix/internal/bugclusters"
	"infra/appengine/weetbix/internal/bugs"
	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/config"
	"infra/appengine/weetbix/internal/services/resultingester"
)

// authGroup is the name of the LUCI Auth group that controls whether the user
// should have access to Weetbix.
const authGroup = "weetbix-access"

func init() {
	// TODO (crbug.com/1242998): Remove when this becomes the default (~Jan 2022).
	datastore.EnableSafeGet()
}

// prepareTemplates configures templates.Bundle used by all UI handlers.
func prepareTemplates(opts *server.Options) *templates.Bundle {
	return &templates.Bundle{
		Loader: templates.FileSystemLoader("templates"),
		// Controls whether templates are cached.
		DebugMode: func(context.Context) bool { return !opts.Prod },
		DefaultArgs: func(ctx context.Context, e *templates.Extra) (templates.Args, error) {
			logoutURL, err := auth.LogoutURL(ctx, e.Request.URL.RequestURI())
			if err != nil {
				return nil, err
			}

			return templates.Args{
				"AuthGroup":       authGroup,
				"AuthServiceHost": opts.AuthServiceHost,
				"User":            auth.CurrentUser(ctx).Email,
				"LogoutURL":       logoutURL,
			}, nil
		},
	}
}

// requireAuth is middleware that forces the user to login and checks the
// user is authorised to use Weetbix before handling any request.
// If the user is not authorised, a standard "access is denied" page is
// displayed that allows the user to logout and login again with new
// credentials.
func requireAuth(ctx *router.Context, next router.Handler) {
	user := auth.CurrentIdentity(ctx.Context)
	if user.Kind() == identity.Anonymous {
		// User is not logged in.
		url, err := auth.LoginURL(ctx.Context, ctx.Request.URL.RequestURI())
		if err != nil {
			logging.Errorf(ctx.Context, "Fetching LoginURL: %s", err.Error())
			http.Error(ctx.Writer, "Internal server error while fetching Login URL.", http.StatusInternalServerError)
		} else {
			http.Redirect(ctx.Writer, ctx.Request, url, http.StatusFound)
		}
		return
	}

	isAuthorised, err := auth.IsMember(ctx.Context, authGroup)
	switch {
	case err != nil:
		logging.Errorf(ctx.Context, "Checking Auth Membership: %s", err.Error())
		http.Error(ctx.Writer, "Internal server error while checking authorisation.", http.StatusInternalServerError)
	case !isAuthorised:
		ctx.Writer.WriteHeader(http.StatusForbidden)
		templates.MustRender(ctx.Context, ctx.Writer, "pages/access-denied.html", nil)
	default:
		next(ctx)
	}
}

func pageBase(srv *server.Server) router.MiddlewareChain {
	return router.NewMiddlewareChain(
		auth.Authenticate(srv.CookieAuth),
		templates.WithTemplates(prepareTemplates(&srv.Options)),
		requireAuth,
	)
}

type handlers struct {
	CloudProject string
}

func (hc *handlers) indexPage(ctx *router.Context) {
	templates.MustRender(ctx.Context, ctx.Writer, "pages/index.html", templates.Args{})
}

func (hc *handlers) monorailTest(ctx *router.Context) {
	// TODO(crbug.com/1243174): Replace as part of MVP development.
	cfg, err := config.Get(ctx.Context)
	if err != nil {
		http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
		return
	}
	mc, err := bugs.NewMonorailClient(ctx.Context, cfg.GetMonorailHostname())
	if err != nil {
		logging.Errorf(ctx.Context, "Getting Monorail client: %s", err)
		http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
		return
	}
	issue, err := mc.GetIssue(ctx.Context, "projects/chromium/issues/2")
	if err != nil {
		logging.Errorf(ctx.Context, "Getting Monorail issue: %s", err)
		http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
		return
	}
	respondWithJSON(ctx, issue)
}

func (hc *handlers) listBugClusters(ctx *router.Context) {
	transctx, cancel := spanmodule.ReadOnlyTransaction(ctx.Context)
	defer cancel()

	var bcs []*bugclusters.BugCluster
	var err error
	clusterID := ctx.Request.URL.Query().Get("cluster")
	if clusterID != "" {
		bcs, err = bugclusters.ReadBugsForCluster(transctx, clusterID)
		if err != nil {
			logging.Errorf(ctx.Context, "Reading bugs for cluster %s: %s", clusterID, err)
			http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
			return
		}
	} else {
		bcs, err = bugclusters.ReadActive(transctx)
		if err != nil {
			logging.Errorf(ctx.Context, "Reading bugs: %s", err)
			http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
			return
		}
	}

	respondWithJSON(ctx, bcs)
}

func (hc *handlers) listClusters(ctx *router.Context) {
	cc, err := clustering.NewClient(ctx.Context, hc.CloudProject)
	if err != nil {
		log.Errorf(ctx.Context, "Creating new clustering client: %v", err)
		http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
	}
	defer func() {
		if err := cc.Close(); err != nil {
			log.Warningf(ctx.Context, "Closing clustering client: %v", err)
		}
	}()

	opts := clustering.ImpactfulClusterReadOptions{
		Project: "chromium",
		Thresholds: clustering.ImpactThresholds{
			UnexpectedFailures1d: 1000,
			UnexpectedFailures3d: 3000,
			UnexpectedFailures7d: 7000,
		},
	}
	clusters, err := cc.ReadImpactfulClusters(ctx.Context, opts)
	if err != nil {
		logging.Errorf(ctx.Context, "Reading Clusters from BigQuery: %s", err)
		http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
		return
	}

	respondWithJSON(ctx, clusters)
}

func (hc *handlers) getCluster(ctx *router.Context) {
	clusterID := strings.TrimPrefix(ctx.Params.ByName("id"), "/")
	if clusterID == "" {
		http.Error(ctx.Writer, "Please supply a cluster ID.", http.StatusBadRequest)
		return
	}
	cc, err := clustering.NewClient(ctx.Context, hc.CloudProject)
	if err != nil {
		log.Errorf(ctx.Context, "Creating new clustering client: %v", err)
		http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := cc.Close(); err != nil {
			log.Warningf(ctx.Context, "Closing clustering client: %v", err)
		}
	}()

	clusters, err := cc.ReadCluster(ctx.Context, clusterID)
	if err != nil {
		logging.Errorf(ctx.Context, "Reading Cluster from BigQuery: %s", err)
		http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
		return
	}

	respondWithJSON(ctx, clusters)
}

func (hc *handlers) updateBugs(ctx context.Context) error {
	cfg, err := config.Get(ctx)
	if err != nil {
		return errors.Annotate(err, "get config").Err()
	}
	// TODO(crbug.com/1243174): Replace with (possibly project-specific) configuration.
	reporter := "chops-weetbix-dev@appspot.gserviceaccount.com"
	t := clustering.ImpactThresholds{
		UnexpectedFailures1d: 1000,
		UnexpectedFailures3d: 3000,
		UnexpectedFailures7d: 7000,
	}
	err = bugclusters.UpdateBugs(ctx, cfg.MonorailHostname, hc.CloudProject, reporter, t)
	if err != nil {
		return errors.Annotate(err, "update bugs").Err()
	}
	return nil
}

func respondWithJSON(ctx *router.Context, data interface{}) {
	bytes, err := json.Marshal(data)
	if err != nil {
		logging.Errorf(ctx.Context, "Marshalling JSON for response: %s", err)
		http.Error(ctx.Writer, "Internal server error.", http.StatusInternalServerError)
	}
	ctx.Writer.Header().Add("Content-Type", "application/json")
	if _, err := ctx.Writer.Write(bytes); err != nil {
		logging.Errorf(ctx.Context, "Writing JSON response: %s", err)
	}
}

func main() {
	modules := []module.Module{
		cfgmodule.NewModuleFromFlags(),
		cron.NewModuleFromFlags(),
		encryptedcookies.NewModuleFromFlags(), // Required for auth sessions.
		gaeemulation.NewModuleFromFlags(),     // Needed by cfgmodule.
		secrets.NewModuleFromFlags(),          // Needed by encryptedcookies.
		spanmodule.NewModuleFromFlags(),
		tq.NewModuleFromFlags(),
	}
	server.Main(nil, modules, func(srv *server.Server) error {
		mw := pageBase(srv)

		handlers := &handlers{CloudProject: srv.Options.CloudProject}
		srv.Routes.GET("/api/monorailtest", mw, handlers.monorailTest)
		srv.Routes.GET("/api/cluster/*id", mw, handlers.getCluster)
		srv.Routes.GET("/api/cluster", mw, handlers.listClusters)
		srv.Routes.GET("/api/bugcluster", mw, handlers.listBugClusters)
		srv.Routes.Static("/static/", mw, http.Dir("./ui/dist"))
		// Anything that is not found, serve app html and let the client side router handle it.
		srv.Routes.NotFound(mw, handlers.indexPage)

		// GAE crons.
		cron.RegisterHandler("read-config", config.Update)
		cron.RegisterHandler("update-bugs", handlers.updateBugs)

		// Pub/Sub subscription endpoints.
		srv.Routes.POST("/_ah/push-handlers/buildbucket", nil, app.BuildbucketPubSubHandler)
		resultingester.RegisterResultIngestionTasksClass()

		return nil
	})
}
