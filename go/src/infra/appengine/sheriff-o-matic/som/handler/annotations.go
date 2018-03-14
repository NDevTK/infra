// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handler

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"google.golang.org/appengine"

	"infra/appengine/sheriff-o-matic/som/client"
	"infra/appengine/sheriff-o-matic/som/model"
	"infra/monorail"

	"golang.org/x/net/context"

	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/gae/service/info"
	"go.chromium.org/gae/service/memcache"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/xsrf"
	"go.chromium.org/luci/server/router"
)

const (
	annotationsCacheKey = "annotation-metadata"
	// annotations will expire after this amount of time
	annotationExpiration = time.Hour * 24 * 10
)

// AnnotationResponse ... The Annotation object extended with cached bug data.
type AnnotationResponse struct {
	model.Annotation
	BugData map[string]monorail.Issue `json:"bug_data"`
}

func makeAnnotationResponse(a *model.Annotation, meta map[string]monorail.Issue) *AnnotationResponse {
	bugs := make(map[string]monorail.Issue)
	for _, b := range a.Bugs {
		if bugData, ok := meta[b]; ok {
			bugs[b] = bugData
		}
	}
	return &AnnotationResponse{*a, bugs}
}

// GetAnnotationsHandler retrieves a set of annotations.
func GetAnnotationsHandler(ctx *router.Context) {
	c, w, p := ctx.Context, ctx.Writer, ctx.Params

	tree := p.ByName("tree")

	q := datastore.NewQuery("Annotation")

	if tree != "" {
		q = q.Ancestor(datastore.MakeKey(c, "Tree", tree))
	}

	annotations := []*model.Annotation{}
	datastore.GetAll(c, q, &annotations)

	meta, err := getAnnotationsMetaData(c)

	if err != nil {
		logging.Errorf(c, "while fetching annotation metadata")
	}

	output := make([]*AnnotationResponse, len(annotations))
	for i, a := range annotations {
		output[i] = makeAnnotationResponse(a, meta)
	}

	data, err := json.Marshal(output)
	if err != nil {
		ErrStatus(c, w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func getAnnotationsMetaData(c context.Context) (map[string]monorail.Issue, error) {
	item, err := memcache.GetKey(c, annotationsCacheKey)
	val := make(map[string]monorail.Issue)

	if err == memcache.ErrCacheMiss {
		logging.Warningf(c, "No annotation metadata in memcache, refreshing...")
		val, err = refreshAnnotations(c, nil)

		if err != nil {
			return nil, err
		}
	} else {
		if err = json.Unmarshal(item.Value(), &val); err != nil {
			logging.Errorf(c, "while unmarshaling metadata in getAnnotationsMetaData")
			return nil, err
		}
	}

	return val, nil
}

// RefreshAnnotationsHandler refreshes the set of annotations.
func RefreshAnnotationsHandler(ctx *router.Context) {
	c, w := ctx.Context, ctx.Writer

	bugMap, err := refreshAnnotations(c, nil)
	if err != nil {
		ErrStatus(c, w, http.StatusInternalServerError, err.Error())
		return
	}

	data, err := json.Marshal(bugMap)
	if err != nil {
		ErrStatus(c, w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Update the cache for annotation bug data.
func refreshAnnotations(c context.Context, a *model.Annotation) (map[string]monorail.Issue, error) {
	q := datastore.NewQuery("Annotation")
	results := []*model.Annotation{}
	datastore.GetAll(c, q, &results)

	// Monorail takes queries of the format id:1,2,3 (gets bugs with those ids).
	mq := "id:"

	if a != nil {
		results = append(results, a)
	}

	allBugs := stringset.New(len(results))
	for _, ann := range results {
		for _, b := range ann.Bugs {
			allBugs.Add(b)
		}
	}

	bugsSlice := allBugs.ToSlice()
	// Sort so that tests are consistent.
	sort.Strings(bugsSlice)
	mq = fmt.Sprintf("%s%s", mq, strings.Join(bugsSlice, ","))

	issues, err := getBugsFromMonorail(c, mq, monorail.IssuesListRequest_ALL)
	if err != nil {
		return nil, err
	}

	// Turn the bug data into a map with the bug id as a key for easier searching.
	m := make(map[string]monorail.Issue)

	for _, b := range issues.Items {
		key := fmt.Sprintf("%d", b.Id)
		m[key] = *b
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	item := memcache.NewItem(c, annotationsCacheKey).SetValue(bytes)

	err = memcache.Set(c, item)

	if err != nil {
		return nil, err
	}

	return m, nil
}

type postRequest struct {
	XSRFToken string           `json:"xsrf_token"`
	Data      *json.RawMessage `json:"data"`
}

// PostAnnotationsHandler handles updates to annotations.
func PostAnnotationsHandler(ctx *router.Context) {
	c, w, r, p := ctx.Context, ctx.Writer, ctx.Request, ctx.Params

	tree := p.ByName("tree")
	action := p.ByName("action")
	if action != "add" && action != "remove" {
		ErrStatus(c, w, http.StatusBadRequest, "unrecognized annotation action")
		return
	}

	req := &postRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		ErrStatus(c, w, http.StatusBadRequest, fmt.Sprintf("while decoding request: %s", err))
		return
	}

	if err := xsrf.Check(c, req.XSRFToken); err != nil {
		ErrStatus(c, w, http.StatusForbidden, err.Error())
		return
	}

	// Extract the annotation key from the otherwise unparsed body.
	rawJSON := struct{ Key string }{}
	if err := json.Unmarshal([]byte(*req.Data), &rawJSON); err != nil {
		ErrStatus(c, w, http.StatusBadRequest, fmt.Sprintf("while decoding request: %s", err))
	}

	key := rawJSON.Key

	annotation := &model.Annotation{
		Tree:      datastore.MakeKey(c, "Tree", tree),
		KeyDigest: fmt.Sprintf("%x", sha1.Sum([]byte(key))),
		Key:       key,
	}

	err = datastore.Get(c, annotation)
	if action == "remove" && err != nil {
		logging.Errorf(c, "while getting %s: %s", key, err)
		ErrStatus(c, w, http.StatusNotFound, fmt.Sprintf("Annotation %s not found", key))
		return
	}

	needRefresh := false
	if info.AppID(c) != "" && info.AppID(c) != "app" {
		c = appengine.WithContext(c, r)
	}
	// The annotation probably doesn't exist if we're adding something.
	data := bytes.NewReader([]byte(*req.Data))
	if action == "add" {
		needRefresh, err = annotation.Add(c, data)
	} else if action == "remove" {
		needRefresh, err = annotation.Remove(c, data)
	}

	if err != nil {
		ErrStatus(c, w, http.StatusBadRequest, err.Error())
		return
	}

	err = r.Body.Close()
	if err != nil {
		ErrStatus(c, w, http.StatusInternalServerError, err.Error())
		return
	}

	err = datastore.Put(c, annotation)
	if err != nil {
		ErrStatus(c, w, http.StatusInternalServerError, err.Error())
		return
	}

	var m map[string]monorail.Issue
	// Refresh the annotation cache on a write. Note that we want the rest of the
	// code to still run even if this fails.
	if needRefresh {
		logging.Infof(c, "Refreshing annotation metadata, due to a stateful modification.")
		m, err = refreshAnnotations(c, annotation)
		if err != nil {
			logging.Errorf(c, "while refreshing annotation cache on post: %s", err)
		}
	} else {
		m, err = getAnnotationsMetaData(c)
		if err != nil {
			logging.Errorf(c, "while getting annotation metadata: %s", err)
		}

	}

	resp, err := json.Marshal(makeAnnotationResponse(annotation, m))
	if err != nil {
		ErrStatus(c, w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// FlushOldAnnotationsHandler culls obsolute annotations from the datastore.
func FlushOldAnnotationsHandler(ctx *router.Context) {
	c, w := ctx.Context, ctx.Writer

	numDeleted, err := flushOldAnnotations(c)
	if err != nil {
		ErrStatus(c, w, http.StatusInternalServerError, err.Error())
		return
	}

	s := fmt.Sprintf("deleted %d annotations", numDeleted)
	logging.Debugf(c, s)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}

func flushOldAnnotations(c context.Context) (int, error) {
	q := datastore.NewQuery("Annotation")
	q = q.Lt("ModificationTime", clock.Get(c).Now().Add(-annotationExpiration))
	q = q.KeysOnly(true)

	results := []*model.Annotation{}
	err := datastore.GetAll(c, q, &results)
	if err != nil {
		return 0, fmt.Errorf("while fetching annotations to delete: %s", err)
	}

	for _, ann := range results {
		logging.Debugf(c, "Deleting %#v\n", ann)
	}

	err = datastore.Delete(c, results)
	if err != nil {
		return 0, fmt.Errorf("while deleting annotations: %s", err)
	}

	return len(results), nil
}

// FileBugHandler files a new bug in monorail.
func FileBugHandler(ctx *router.Context) {
	c, w, r := ctx.Context, ctx.Writer, ctx.Request

	req := &postRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		ErrStatus(c, w, http.StatusBadRequest, fmt.Sprintf("while decoding request: %s", err))
		return
	}

	if err := xsrf.Check(c, req.XSRFToken); err != nil {
		ErrStatus(c, w, http.StatusForbidden, err.Error())
		return
	}

	rawJSON := struct {
		Summary     string
		Description string
		Cc          []string
		Priority    string
		Labels      []string
	}{}
	if err := json.Unmarshal([]byte(*req.Data), &rawJSON); err != nil {
		ErrStatus(c, w, http.StatusBadRequest, fmt.Sprintf("while decoding request: %s", err))
	}

	ccList := make([]*monorail.AtomPerson, len(rawJSON.Cc))
	for i, cc := range rawJSON.Cc {
		ccList[i] = &monorail.AtomPerson{cc}
	}

	sa, err := info.ServiceAccount(c)
	if err != nil {
		logging.Errorf(c, "failed to get service account: %v", err)
		ErrStatus(c, w, http.StatusInternalServerError, err.Error())
		return
	}

	user := auth.CurrentIdentity(c)
	description := fmt.Sprintf("Filed by %s on behalf of %s\n\n%s", sa, user.Email(),
		rawJSON.Description)

	fileBugReq := &monorail.InsertIssueRequest{
		ProjectId: "chromium",
		Issue: &monorail.Issue{
			Cc:          ccList,
			Summary:     rawJSON.Summary,
			Description: description,
			Status:      "Available",
			Labels:      rawJSON.Labels,
		},
	}

	mr := client.GetMonorail(c)

	res, err := mr.InsertIssue(c, fileBugReq)
	if err != nil {
		logging.Errorf(c, "error inserting new Issue: %v", err)
		ErrStatus(c, w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := json.Marshal(res)
	if err != nil {
		ErrStatus(c, w, http.StatusInternalServerError, err.Error())
		return
	}

	logging.Infof(c, "%v", out)
	w.Header().Set("Content-Type", "applications/json")
	w.Write(out)
}
