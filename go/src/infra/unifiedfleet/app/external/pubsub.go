// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package external

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

// unique key used to store and retrieve a pubsub client.
var pubsubKey = "ufs pubsub key"

// GetPubSub returns the client in c if it exists, or nil.
func GetPubSub(c context.Context) *pubsub.Client {
	client := c.Value(&pubsubKey)
	if client == nil {
		return nil
	}

	return client.(*pubsub.Client)
}

// UsePubSub installs a pubsub client into the context for later use.
func UsePubSub(c context.Context, projectID string) (context.Context, error) {
	// Create client for Pub/Sub publishing.
	client, err := pubsub.NewClient(c, projectID)
	if err != nil {
		// The precise value of the error that we got back from the pubsub client is important for local debugging.
		// See b:267829708 for details.
		return nil, fmt.Errorf("failed to create Pub/Sub client for project %q: %w", projectID, err)
	}
	return context.WithValue(c, &pubsubKey, client), nil
}
