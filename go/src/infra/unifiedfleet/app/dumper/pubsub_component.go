// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"
	"fmt"
	"runtime"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/logging"
	"golang.org/x/sync/semaphore"
)

func createPubSubTopicClient(ctx context.Context, projectID, topicID string) (*pubsub.Topic, *pubsub.Client, error) {
	// Create client for Pub/Sub publishing.
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create Pub/Sub client for project %s", projectID)
	}

	// Associate the Pub/Sub client with the correct topicID.
	topic := client.Topic(topicID)

	// Check if the topic exists on the project.
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("Error when checking if topic %s on project %s. error: %s", topicID, projectID, err.Error())
	}

	// Attempt to create the topic if it doesn't exist.
	if !exists {
		topic, err = client.CreateTopic(ctx, topicID)
		if err != nil {
			return nil, nil, fmt.Errorf("Topic %s doesn't exist and failed to be created. error: %s", topicID, err.Error())
		}
	}

	return topic, client, nil
}

// publishToTopic publishToTopic upload objects to a given pub/sub topic.
//
// It retrieves rows in the form of a list of protos and uploads
// them to the given pub/sub topic.
func publishToTopic(ctx context.Context, msgs []proto.Message, projectID, topicID string) error {
	// Send the messages to Pub/Sub in parallel.
	errChan := make(chan error, len(msgs))

	logging.Infof(ctx, "dumping %d messages to pubsub", len(msgs))

	var (
		// Sets to max number of CPUs that can can run simultaneously. https://pkg.go.dev/runtime#GOMAXPROCS
		maxWorkers = runtime.GOMAXPROCS(0)
		sem        = semaphore.NewWeighted(int64(maxWorkers))
	)

	for i := 0; i < len(msgs); i++ {
		// Acquire a lock to run another goroutine.
		if err := sem.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("Failed to acquire semaphore: %v", err)
		}

		go func(msg proto.Message) {
			publish(ctx, projectID, topicID, msg, errChan)
			defer sem.Release(1)
		}(msgs[i])
	}

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
func publish(ctx context.Context, projectID, topicId string, msg proto.Message, retChan chan error) {
	// Convert the proto representation of the row into a JSON.
	data, err := proto.Marshal(msg)
	if err != nil {
		retChan <- err
		return
	}
	topic, client, err := createPubSubTopicClient(ctx, projectID, topicId)
	defer client.Close()
	if err != nil {
		retChan <- err
		return
	}

	// Asynchronously publish the message.
	result := topic.Publish(ctx, &pubsub.Message{Data: data})

	// Block and wait to check for publishing errors.
	_, err = result.Get(ctx)
	if err != nil {
		retChan <- err
		return
	}
}
