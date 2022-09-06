// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/logging"
	"sync"
)

// publishToTopic publishToTopic upload objects to a given pub/sub topic.
//
// It retrieves rows in the form of a list of protos and uploads
// them to the given pub/sub topic.
func publishToTopic(ctx context.Context, msgs []proto.Message, projectID, topicID string) error {
	// Send the messages to Pub/Sub in parallel.
	errChan := make(chan error)
	var wg sync.WaitGroup
	logging.Infof(ctx, "dumping %d messages to pubsub", len(msgs))
	for _, msg := range msgs {
		wg.Add(1)
		go func(msg proto.Message) {
			publish(ctx, projectID, topicID, msg, errChan)
			defer wg.Done()
		}(msg)
	}

	// Wait for publishing attempts to finish.
	wg.Wait()
	close(errChan)

	// Generate an error for any and all failed publishing attempts.
	failedPublishCount := 0
	runningError := ""
	for errMsg := range errChan {
		// Append the error message to running error.
		if errMsg != nil {
			failedPublishCount++
			runningError = fmt.Sprintf("%s\n%s", runningError, errMsg.Error())
		}
	}

	// If any errors in publishing occurred, return them all at once.
	if failedPublishCount > 0 {
		logging.Warningf(ctx, "pubsub: %d rows failed with the following errors:", failedPublishCount)
		return fmt.Errorf("pubsub: %d rows failed with the following errors:\n%s", failedPublishCount, runningError)
	}
	logging.Infof(ctx, "dumped all %d messages to pubsub topic projects/%s/topic/%s", len(msgs), projectID, topicID)
	// All messages published properly.
	return nil
}

// publish wraps all the steps required to send a message to a Pub/Sub topic.
func publish(ctx context.Context, projectID, topicID string, msg proto.Message, retChan chan error) {
	// Create client for Pub/Sub publishing.
	client, err := pubsub.NewClient(ctx, projectID)
	defer client.Close()
	if err != nil {
		retChan <- fmt.Errorf("Failed to create Pub/Sub client for projects/%s/topic/%s", projectID, topicID)
		return
	}

	// Attempt to create the topic if it doesn't exist.
	// The alternative to this would be manually creating the Pub/Sub topic
	// beforehand.
	_, _ = client.CreateTopic(ctx, topicID)

	// Associate the Pub/Sub client with the correct topicID.
	topic := client.Topic(topicID)

	// Convert the proto representation of the row into a JSON.
	data, err := proto.Marshal(msg)
	if err != nil {
		retChan <- err
		return
	}

	// Asynchronously publish the message.
	result := topic.Publish(ctx, &pubsub.Message{Data: data})

	// Block and wait to check for publishing errors.
	_, err = result.Get(ctx)
	if err != nil {
		retChan <- fmt.Errorf("Failed to get publishing ID from server\nget: %s", err.Error())
		return
	}
}
