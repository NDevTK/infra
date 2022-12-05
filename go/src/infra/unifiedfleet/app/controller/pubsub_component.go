// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"infra/unifiedfleet/app/external"

	"cloud.google.com/go/pubsub"
	"go.chromium.org/luci/common/logging"
)

// MarshalAndPublish converts the given struct into json and sends it to the
// provided Pub/Sub destination.
func marshalAndPublish(ctx context.Context, item any, topicID string) {
	// Marshal the data into json for the Pub/Sub message.
	data, err := json.Marshal(item)

	// If we receive an error while marshalling the data log the error then
	// return with a passing result.
	if err != nil {
		logging.Warningf(ctx, err.Error())
		return
	}

	// Similar to above, if we encounter an error on the upload then log it
	// and return wth a passing result.
	err = publishToTopic(ctx, data, topicID)
	if err != nil {
		logging.Warningf(ctx, err.Error())
	}
}

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
func publish(ctx context.Context, topicId string, data []byte) error {
	topic, err := createPubSubTopicClient(ctx, topicId)
	if err != nil {
		return err
	}
	// Asynchronously publish the message.
	result := topic.Publish(ctx, &pubsub.Message{Data: data})

	// Block and wait to check for publishing errors.
	_, err = result.Get(ctx)
	if err != nil {
		return err
	}
	return nil
}

// publishToTopic publishToTopic uploads JSON messages to a given pub/sub topic
// on the context defined client.
func publishToTopic(ctx context.Context, msg []byte, topicID string) error {
	logging.Infof(ctx, "dumping message to pubsub\n")

	err := publish(ctx, topicID, msg)
	if err != nil {
		return err
	}

	logging.Infof(ctx, "Successfully dumped message to pubsub\n")

	// Message published properly.
	return nil
}
