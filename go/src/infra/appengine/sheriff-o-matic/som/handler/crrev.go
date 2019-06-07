package handler

import (
	"fmt"
	"net/http"
	"strings"

	"infra/appengine/sheriff-o-matic/som/client"

	"golang.org/x/net/context"

	"go.chromium.org/gae/service/memcache"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/router"
)

// getOAuthClient returns a client capable of making HTTP requests authenticated
// with OAuth access token for userinfo.email scope.
var getOAuthClient = func(c context.Context) (*http.Client, error) {
	// Note: "https://www.googleapis.com/auth/userinfo.email" is the default
	// scope used by GetRPCTransport(AsSelf). Use auth.WithScopes(...) option to
	// override.
	t, err := auth.GetRPCTransport(c, auth.AsSelf)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: t}, nil
}

// GetRevRangeHandler returns a revision range queury for gitiles, given two
// git hashes.
func GetRevRangeHandler(ctx *router.Context) {
	crRev := client.NewCrRev("https://cr-rev.appspot.com")
	getRevRangeHandler(ctx, crRev)
}

func getRevRangeHandler(ctx *router.Context, crRev client.CrRev) {
	c, w, r, p := ctx.Context, ctx.Writer, ctx.Request, ctx.Params

	start := p.ByName("start")
	end := p.ByName("end")
	host := p.ByName("host")
	project := p.ByName("project")
	if start == "" || end == "" {
		errStatus(c, w, http.StatusBadRequest, "Start and end parameters must be set.")
		return
	}

	itm := memcache.NewItem(c, fmt.Sprintf("revrange:%s..%s", start, end))
	err := memcache.Get(c, itm)

	// TODO: nix this double layer of caching..
	if true { //err == memcache.ErrCacheMiss {
		// TODO(seanmccullough): some sanity checking of the rev json (same repo etc)

		if host == "" {
			host = "chromium"
		}
		if project == "" {
			project = "chromium/src"
		} else {
			project = strings.Replace(project, "^", "/", -1)
		}

		gitilesURL := fmt.Sprintf("https://%s.googlesource.com/%s/+log/%s^..%s?format=JSON",
			host, project, start, end)

		itm.SetValue([]byte(gitilesURL))
		if err = memcache.Set(c, itm); err != nil {
			errStatus(c, w, http.StatusInternalServerError, fmt.Sprintf("while setting memcache: %s", err))
			return
		}
	} else if err != nil {
		errStatus(c, w, http.StatusInternalServerError, fmt.Sprintf("while getting memcache: %s", err))
		return
	}
	logging.Infof(c, "%+v", string(itm.Value()))
	http.Redirect(w, r, string(itm.Value()), 301)
}
