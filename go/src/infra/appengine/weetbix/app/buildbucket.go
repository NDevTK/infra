// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	bbv1 "go.chromium.org/luci/common/api/buildbucket/buildbucket/v1"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
	"go.chromium.org/luci/server/router"
	"google.golang.org/protobuf/types/known/timestamppb"

	ctlpb "infra/appengine/weetbix/internal/ingestion/control/proto"
)

const (
	// userAgentTagKey is the key of the user agent tag.
	userAgentTagKey = "user_agent"
	// userAgentCQ is the value of the user agent tag, for builds started
	// by LUCI CV.
	userAgentCQ = "cq"
)

var (
	buildCounter = metric.NewCounter(
		"weetbix/ingestion/pubsub/buildbucket_builds",
		"The number of buildbucket builds received by Weetbix from PubSub.",
		nil,
		// The LUCI Project.
		field.String("project"),
		// "success", "ignored", "transient-failure" or "permanent-failure".
		field.String("status"))

	// chromiumMilestoneProjectPrefix is the LUCI project prefix
	// of chromium milestone projects, e.g. chromium-m100.
	chromiumMilestoneProjectRE = regexp.MustCompile(`^(chrome|chromium)-m[0-9]+$`)
)

// BuildbucketPubSubHandler accepts and process buildbucket Pub/Sub messages.
// As of Aug 2021, Weetbix subscribes to this Pub/Sub topic to get completed
// Chromium CI builds.
// For CQ builds, Weetbix uses CV Pub/Sub as the entrypoint.
func BuildbucketPubSubHandler(ctx *router.Context) {
	project := "unknown"
	status := "unknown"
	defer func() {
		// Closure for late binding.
		buildCounter.Add(ctx.Context, 1, project, status)
	}()

	project, processed, err := bbPubSubHandlerImpl(ctx.Context, ctx.Request)
	if err != nil {
		errors.Log(ctx.Context, errors.Annotate(err, "handling buildbucket pubsub event").Err())
		status = processErr(ctx, err)
		return
	}
	if processed {
		status = "success"
		// Use subtly different "success" response codes to surface in
		// standard GAE logs whether an ingestion was ignored or not,
		// while still acknowledging the pub/sub.
		// See https://cloud.google.com/pubsub/docs/push#receiving_messages.
		ctx.Writer.WriteHeader(http.StatusOK)
	} else {
		status = "ignored"
		ctx.Writer.WriteHeader(http.StatusNoContent) // 204
	}
}

func bbPubSubHandlerImpl(ctx context.Context, request *http.Request) (project string, processed bool, err error) {
	msg, err := parseBBMessage(ctx, request)
	if err != nil {
		return "unknown", false, errors.Annotate(err, "failed to parse buildbucket pub/sub message").Err()
	}
	processed, err = processBBMessage(ctx, msg)
	if err != nil {
		return msg.Build.Project, false, errors.Annotate(err, "processing build").Err()
	}
	return msg.Build.Project, processed, nil
}

type buildBucketMessage struct {
	Build    bbv1.LegacyApiCommonBuildMessage
	Hostname string
}

func parseBBMessage(ctx context.Context, r *http.Request) (*buildBucketMessage, error) {
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

func processBBMessage(ctx context.Context, message *buildBucketMessage) (processed bool, err error) {
	if message.Build.Status != bbv1.StatusCompleted {
		// Received build that hasn't completed yet, ignore it.
		return false, nil
	}
	if message.Build.CreatedTs == 0 {
		return false, errors.New("build did not have created timestamp specified")
	}

	if chromiumMilestoneProjectRE.MatchString(message.Build.Project) {
		// Chromium milestone projects are currently not supported.
		return false, nil
	}

	userAgents := extractTagValues(message.Build.Tags, userAgentTagKey)
	isPresubmit := len(userAgents) == 1 && userAgents[0] == userAgentCQ

	project := message.Build.Project
	id := buildID(message.Hostname, message.Build.Id)
	result := &ctlpb.BuildResult{
		CreationTime: timestamppb.New(bbv1.ParseTimestamp(message.Build.CreatedTs)),
		Id:           message.Build.Id,
		Host:         message.Hostname,
		Project:      project,
	}

	if err := JoinBuildResult(ctx, id, project, isPresubmit, result); err != nil {
		return false, errors.Annotate(err, "joining build result").Err()
	}
	return true, nil
}

func extractTagValues(tags []string, key string) []string {
	var values []string
	for _, tag := range tags {
		tagParts := strings.SplitN(tag, ":", 2)
		if len(tagParts) < 2 {
			// Invalid tag.
			continue
		}
		if tagParts[0] == key {
			values = append(values, tagParts[1])
		}
	}
	return values
}
