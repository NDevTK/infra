// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package pubsub handles pub/sub messages
package pubsub

import (
	"context"
	"encoding/json"
	"net/http"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry/transient"
	"go.chromium.org/luci/server/router"

	"infra/appengine/gofindit/compilefailuredetection"

	bbv1 "go.chromium.org/luci/common/api/buildbucket/buildbucket/v1"
)

type pubsubMessage struct {
	Message struct {
		Data []byte
	}
	Attributes map[string]interface{}
}

type buildBucketMessage struct {
	Build    bbv1.LegacyApiCommonBuildMessage
	Hostname string
}

// BuildbucketPubSubHandler handles pub/sub messages from buildbucket
func BuildbucketPubSubHandler(ctx *router.Context) {
	logging.Infof(ctx.Context, "Received buildbucket pubsub message")
	if err := buildbucketPubSubHandlerImpl(ctx.Context, ctx.Request); err != nil {
		logging.Errorf(ctx.Context, "Error processing buildbucket pubsub message: %s", err)
		processError(ctx, err)
		return
	}
	// Just returns OK here so pubsub does not resend the message
	ctx.Writer.WriteHeader(http.StatusOK)
}

func processError(ctx *router.Context, err error) {
	if transient.Tag.In(err) {
		// Pubsub will retry this
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
	} else {
		// Pubsub will not retry those errors
		ctx.Writer.WriteHeader(http.StatusAccepted)
	}
}

func buildbucketPubSubHandlerImpl(c context.Context, r *http.Request) error {
	bbmsg, err := parseBBMessage(r)
	if err != nil {
		return err
	}
	logging.Infof(c, "Received message for build id %s", bbmsg.Build.Id)

	// For now, we only handle chromium/ci builds
	// TODO (nqmtuan): Move this into config
	if !(bbmsg.Build.Project == "chromium" && bbmsg.Build.Bucket == "ci") {
		logging.Infof(c, "Unsupported build for bucket (%q, %q). Exiting early...", bbmsg.Build.Project, bbmsg.Build.Bucket)
		return nil
	}

	// Just ignore non-completed builds
	if bbmsg.Build.Status != bbv1.StatusCompleted {
		return nil
	}

	// We only care about failed builds
	if bbmsg.Build.Result != bbv1.ResultFailure {
		return nil
	}

	_, err = compilefailuredetection.AnalyzeBuild(c, bbmsg.Build.Id)
	return err
}

func parseBBMessage(r *http.Request) (*buildBucketMessage, error) {
	var psMsg pubsubMessage
	if err := json.NewDecoder(r.Body).Decode(&psMsg); err != nil {
		return nil, errors.Annotate(err, "could not decode buildbucket pubsub message").Err()
	}

	var bbMsg buildBucketMessage
	if err := json.Unmarshal(psMsg.Message.Data, &bbMsg); err != nil {
		return nil, errors.Annotate(err, "could not parse buildbucket pubsub message data").Err()
	}
	return &bbMsg, nil
}
