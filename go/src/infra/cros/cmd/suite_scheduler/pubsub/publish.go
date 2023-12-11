// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package pubsub wraps all the pubsub API interactions that will be required by SuiteScheduler.
package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

// PublishClient defines the minimum requires that this project will need of a
// Pub/Sub API.
type PublishClient interface {
	InitClient(ctx context.Context, projectID string) error
	InitTopic(ctx context.Context, topicID string) error
	PublishMessage(ctx context.Context, data []byte) error
}

// Publish implements the PubsubClient interface.
type Publish struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

// InitPublishClient returns a newly created Pub/Sub Client interface.
func InitPublishClient(ctx context.Context, projectID, topicID string) (PublishClient, error) {
	psClient := &Publish{}
	err := psClient.InitClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	err = psClient.InitTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}

	return psClient, nil
}

// InitClient creates the client interface for the current Pub/Sub Client.
func (c *Publish) InitClient(ctx context.Context, projectID string) error {
	if c.client != nil {
		return fmt.Errorf("client is already initialized")
	}

	var err error
	c.client, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	return nil
}

// InitTopic creates the topic interface for the current Pub/Sub Client.
func (c *Publish) InitTopic(ctx context.Context, topicID string) error {
	if c.client == nil {
		return fmt.Errorf("client has not been initialized yet")
	}

	if c.topic != nil {
		return fmt.Errorf("topic is already initialized")
	}

	c.topic = c.client.Topic(topicID)

	exist, _ := c.topic.Exists(ctx)
	if !exist {
		return fmt.Errorf("topic %s does not exist on project %s", topicID, c.client.Project())
	}

	return nil
}

// PublishMessage sends the provided date to the clients pre-configured Pub/Sub
// topic.
func (c *Publish) PublishMessage(ctx context.Context, data []byte) error {
	if c.topic == nil {
		return fmt.Errorf("no topic is set for pubsub client")
	}

	message := pubsub.Message{
		Data: data,
	}

	result := c.topic.Publish(ctx, &message)

	_, err := result.Get(ctx)
	if err != nil {
		return err
	}

	return nil
}
