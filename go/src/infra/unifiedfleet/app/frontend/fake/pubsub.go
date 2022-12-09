// Copyright 2022 All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package fake

import (
	"context"

	"cloud.google.com/go/pubsub"
)

// unique key used to store and retrieve a pubsub client.
var pubsubKey = "ufs pubsub key"

// FakePubsubClientInterface creates a mock client to be used during testing.
func FakePubsubClientInterface(ctx context.Context) context.Context {
	client, err := pubsub.NewClient(ctx, "testing-project")
	if err != nil {
		return nil
	}
	return context.WithValue(ctx, &pubsubKey, client)
}
