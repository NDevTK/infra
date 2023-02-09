// Copyright 2023 The ChromiumOS Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"encoding/base64"
	"log"

	"infra/cros/support/internal/cli"
	"infra/cros/support/internal/pubsub"

	luciPubsub "go.chromium.org/luci/common/gcloud/pubsub"
	"google.golang.org/api/option"
)

// Publish a message containing `data` to projects/`projectId`/topics/`topic-id`. `data` must be
// base64 encoded.
type input struct {
	ProjectID   string `json:"project_id"`
	TopicID     string `json:"topic_id"`
	Data        string `json:"data"`
	OrderingKey string `json:"ordering_key"`
	EndPoint    string `json:"endpoint"`
}

type output struct {
	MessageID string `json:"message_id"`
}

func main() {
	cli.SetAuthScopes(luciPubsub.SubscriberScopes...)
	cli.Init()

	var input input
	cli.MustUnmarshalInput(&input)
	if len(input.Data) == 0 {
		log.Fatalf("data must not be empty")
	}

	tokenSource, err := cli.AuthenticatedTokenSource()

	if err != nil {
		log.Fatalf("Failed to create TokenSource: %v", err)
	}

	decodedData, err := base64.StdEncoding.DecodeString(input.Data)

	if err != nil {
		log.Fatalf("Failed to decode data: %v", err)
	}

	id, err := pubsub.PublishMessage(
		input.ProjectID,
		input.TopicID,
		input.OrderingKey,
		input.EndPoint,
		decodedData,
		option.WithTokenSource(tokenSource),
	)

	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	cli.MustMarshalOutput(output{MessageID: id})
}
