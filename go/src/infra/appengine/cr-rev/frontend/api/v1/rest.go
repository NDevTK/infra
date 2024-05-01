package api

import (
	"net/http"
	"strconv"

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc/status"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/router"
)

var marshaler = &jsonpb.Marshaler{}

type restAPIServer struct {
	grpcServer CrrevServer
}

// crCPOldReferences holds references to repositories that changed
// Cr-Commit-Position reference. The number indicates the latest commit
// position that uses old reference.
var crCPOldReferences = map[string]map[string]int{
	"chromium": {
		"chromium/src": 913133,
		"v8/v8":        76350,
		"infra/infra":  42976,
	},
	"webrtc": {
		"src": 34825,
	},
}

func (s *restAPIServer) handleRedirect(c *router.Context) {
	// gRPC expects a leading slash. However, router doesn't include it in
	// named parameter.
	q := "/" + c.Params.ByName("query")
	req := &RedirectRequest{
		Query: q,
	}
	resp, err := s.grpcServer.Redirect(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
	}
	marshaler.Marshal(c.Writer, resp)
}

func (s *restAPIServer) handleNumbering(c *router.Context) {
	queryValues := c.Request.URL.Query()
	n, err := strconv.Atoi(queryValues.Get("number"))
	if err != nil {
		http.Error(c.Writer, "Parameter number is not an integer", http.StatusBadRequest)
		return
	}
	host := queryValues.Get("project")
	repository := queryValues.Get("repo")
	ref := queryValues.Get("numbering_identifier")
	// Handle Cr-Commit-Position change in all repositories
	// See: https://crbug.com/1241484
	var lastCommitOldReference int
	if repoMapping, ok := crCPOldReferences[host]; ok {
		if n, ok := repoMapping[repository]; ok {
			lastCommitOldReference = n
		}
	}
	if lastCommitOldReference > 0 {
		if ref == "refs/heads/master" && n > lastCommitOldReference {
			ref = "refs/heads/main"
		} else if ref == "refs/heads/main" && n <= lastCommitOldReference {
			ref = "refs/heads/master"
		}
	}

	req := &NumberingRequest{
		Host:           host,
		Repository:     repository,
		PositionRef:    ref,
		PositionNumber: int64(n),
	}
	resp, err := s.grpcServer.Numbering(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}
	marshaler.Marshal(c.Writer, resp)
}

func (s *restAPIServer) handleCommit(c *router.Context) {
	req := &CommitRequest{
		GitHash: c.Params.ByName("hash"),
	}
	resp, err := s.grpcServer.Commit(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}
	marshaler.Marshal(c.Writer, resp)
}

func handleError(c *router.Context, err error) {
	if err, ok := status.FromError(err); ok {
		http.NotFound(c.Writer, c.Request)
	} else {
		logging.Errorf(c.Request.Context(), "Error in API while handling redirect: %w", err)
		http.Error(c.Writer, "Internal server errror", http.StatusInternalServerError)
	}
}

// NewRESTServer installs REST handlers to provided router.
func NewRESTServer(r *router.Router, grpcServer CrrevServer) {
	s := &restAPIServer{
		grpcServer: grpcServer,
	}
	mw := router.MiddlewareChain{}
	mw = mw.Extend(func(c *router.Context, next router.Handler) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Add(
			"Access-Control-Allow-Headers",
			"Origin, Authorization, Content-Type, Accept, User-Agent")
		c.Writer.Header().Add(
			"Access-Control-Allow-Methods",
			"DELETE, GET, OPTIONS, POST, PUT")
		next(c)
	})

	r.GET("/redirect/:query", mw, s.handleRedirect)
	r.GET("/get_numbering", mw, s.handleNumbering)
	r.GET("/commit/:hash", mw, s.handleCommit)
}
