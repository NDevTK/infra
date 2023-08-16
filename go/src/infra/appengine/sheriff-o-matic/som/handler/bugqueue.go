// Copyright 2016 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/caching"
	"go.chromium.org/luci/server/router"
	"google.golang.org/grpc"

	"infra/appengine/sheriff-o-matic/som/client"
	"infra/appengine/sheriff-o-matic/som/model"
	monorailv3 "infra/monorailv2/api/v3/api_proto"
)

const (
	bugQueueCacheFormat = "bugqueue-%s"
)

var (
	bugQueueLength = metric.NewInt("bug_queue_length", "Number of bugs in queue.",
		nil, field.String("label"))
)

var bugCache = caching.RegisterLRUCache[string, []byte](20)

// IssueClient is for testing purpose
type IssueClient interface {
	SearchIssues(context.Context, *monorailv3.SearchIssuesRequest, ...grpc.CallOption) (*monorailv3.SearchIssuesResponse, error)
}

// SearchIssueResponseExtras wraps around SearchIssuesResponse
// but adds some information
type SearchIssueResponseExtras struct {
	*monorailv3.SearchIssuesResponse
	Extras map[string]interface{} `json:"extras,omitempty"`
}

// BugQueueHandler handles bug queue-related requests.
type BugQueueHandler struct {
	MonorailIssueClient    IssueClient
	DefaultMonorailProject string
}

func (bqh *BugQueueHandler) getBugsFromMonorailV3(c context.Context, q string, projectID string) (*SearchIssueResponseExtras, error) {
	// TODO (nqmtuan): Implement pagination if necessary
	projects := []string{"projects/" + projectID}
	req := monorailv3.SearchIssuesRequest{
		Projects: projects,
		Query:    q,
	}
	before := clock.Now(c)
	resp, err := bqh.MonorailIssueClient.SearchIssues(c, &req)
	if err != nil {
		logging.Errorf(c, "error searching issues: %v", err)
		return nil, err
	}
	logging.Debugf(c, "Fetch to monorail took %v. Got %d bugs.", clock.Now(c).Sub(before), len(resp.Issues))

	// Add extra priority field, since Monorail response does not indicate
	// which field is priority field
	respExtras := &SearchIssueResponseExtras{
		SearchIssuesResponse: resp,
	}
	priorityField, err := client.GetMonorailPriorityField(c, projectID)
	if err == nil {
		respExtras.Extras = make(map[string]interface{})
		respExtras.Extras["priority_field"] = priorityField
	}
	return respExtras, nil
}

// Switches chromium.org emails for google.com emails and vice versa.
// Note that chromium.org emails may be different from google.com emails.
func getAlternateEmail(email string) string {
	s := strings.Split(email, "@")
	if len(s) != 2 {
		return email
	}

	user, domain := s[0], s[1]
	if domain == "chromium.org" {
		return fmt.Sprintf("%s@google.com", user)
	}
	return fmt.Sprintf("%s@chromium.org", user)
}

// GetBugQueueHandler returns a set of bugs for the current user and tree.
func (bqh *BugQueueHandler) GetBugQueueHandler(ctx *router.Context) {
	var err error
	c, w, p := ctx.Request.Context(), ctx.Writer, ctx.Params
	label := p.ByName("label")
	key := fmt.Sprintf(bugQueueCacheFormat, label)

	item, found := bugCache.LRU(c).Get(c, key)

	if !found {
		logging.Debugf(c, "No bug queue data for %s in cache, refreshing...", label)
		item, err = bqh.refreshBugQueue(c, label, bqh.GetMonorailProjectNameFromLabel(c, label))
		if err != nil {
			errStatus(c, w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(item)
}

// GetUncachedBugsHandler bypasses the cache to return the bug queue for current user and tree.
// TODO (nqmtuan): This is not used. We should remove it.
func (bqh *BugQueueHandler) GetUncachedBugsHandler(ctx *router.Context) {
	c, w, p := ctx.Request.Context(), ctx.Writer, ctx.Params

	label := p.ByName("label")

	user := auth.CurrentIdentity(c)
	email := getAlternateEmail(user.Email())
	q := fmt.Sprintf("is:open (label:%[1]s -has:owner OR label:%[1]s owner:%s OR owner:%s label:%[1]s)",
		label, user.Email(), email)

	bugs, err := bqh.getBugsFromMonorailV3(c, q, bqh.GetMonorailProjectNameFromLabel(c, label))
	if err != nil && bugs != nil {
		bugQueueLength.Set(c, int64(len(bugs.Issues)), label)
	}

	out, err := json.Marshal(bugs)
	if err != nil {
		errStatus(c, w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Makes a request to Monorail for bugs in a label and caches the results.
func (bqh *BugQueueHandler) refreshBugQueue(c context.Context, label string, projectID string) ([]byte, error) {
	q := fmt.Sprintf("is:open (label=%s)", label)
	res, err := bqh.getBugsFromMonorailV3(c, q, projectID)

	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	bugCache.LRU(c).Put(c, fmt.Sprintf(bugQueueCacheFormat, label), bytes, 15*time.Minute)

	return bytes, nil
}

// RefreshBugQueueHandler updates the cached bug queue for current tree.
func (bqh *BugQueueHandler) RefreshBugQueueHandler(ctx context.Context) error {
	labels, err := queryAllMonorailLabels(ctx)
	if err != nil {
		return errors.Annotate(err, "getting monorail project names").Err()
	}
	errs := errors.NewMultiError()
	for label, project := range labels {
		_, err := bqh.refreshBugQueue(ctx, label, project)
		errs.MaybeAdd(err)
	}
	return errs.AsError()
}

// GetMonorailProjectNameFromLabel returns the default monorail project name
// configured in project settings by comparing the bugqueue label.
func (bqh *BugQueueHandler) GetMonorailProjectNameFromLabel(c context.Context, label string) string {

	if bqh.DefaultMonorailProject == "" {
		bqh.DefaultMonorailProject = bqh.queryTreeForLabel(c, label)
	}

	return bqh.DefaultMonorailProject
}

func (bqh *BugQueueHandler) queryTreeForLabel(c context.Context, label string) string {
	q := datastore.NewQuery("Tree")
	trees := []*model.Tree{}
	if err := datastore.GetAll(c, q, &trees); err == nil {
		for _, tree := range trees {
			if tree.BugQueueLabel == label && tree.DefaultMonorailProjectName != "" {
				return tree.DefaultMonorailProjectName
			}
		}
	}
	return "chromium"
}

func queryAllMonorailLabels(ctx context.Context) (map[string]string, error) {
	q := datastore.NewQuery("Tree")
	trees := []*model.Tree{}
	if err := datastore.GetAll(ctx, q, &trees); err != nil {
		return nil, err
	}
	labels := map[string]string{}
	for _, tree := range trees {
		if tree.BugQueueLabel != "" {
			labels[tree.BugQueueLabel] = tree.DefaultMonorailProjectName
		}
	}
	return labels, nil
}
