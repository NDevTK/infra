// Copyright 2022 All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"infra/unifiedfleet/app/external"

	"cloud.google.com/go/pubsub"
	"go.chromium.org/luci/common/logging"
)

// createPubSubTopicClient returns back an instance of a Pub/Sub client for the
// given target.
func createPubSubTopicClient(ctx context.Context, topicID string) (*pubsub.Topic, error) {
	// Create client for Pub/Sub publishing.
	client := external.GetPubSub(ctx)
	if client == nil {
		return nil, fmt.Errorf("Pubsub client is nil.")
	}

	// Associate the Pub/Sub client with the correct topicID.
	topic := client.Topic(topicID)

	// Check if the topic exists on the project.go test.
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error when checking if topic %s exist error: %s", topicID, err.Error())
	}

	// Attempt to create the topic if it doesn't exist.
	if !exists {
		topic, err = client.CreateTopic(ctx, topicID)
		if err != nil {
			return nil, fmt.Errorf("Topic %s doesn't exist and failed to be created. error: %s", topicID, err.Error())
		}
	}

	return topic, nil
}

// publish wraps all the steps required to send a message to a Pub/Sub topic.
func publish(ctx context.Context, topicId string, msgs [][]byte) error {
	logging.Debugf(ctx, "Pubsub: attempting to publish %d messages", len(msgs))
	topic, err := createPubSubTopicClient(ctx, topicId)
	if err != nil {
		return err
	}

	var results []*pubsub.PublishResult
	for _, msg := range msgs {
		// Asynchronously publish the message.
		// NOTE: By default publish waits for up-to 1000 messages for batching.
		// NOTE: By default publish spins up a maximum of GOMAXPROCS goroutines.
		result := topic.Publish(ctx, &pubsub.Message{Data: msg})
		results = append(results, result)
	}

	// Block and wait to check for publishing errors.
	for _, result := range results {
		_, err = result.Get(ctx)
		// Rather than flood errors, fail fast.
		if err != nil {
			return err
		}
	}

	logging.Debugf(ctx, "Pubsub: published %d messages", len(results))
	return nil
}
