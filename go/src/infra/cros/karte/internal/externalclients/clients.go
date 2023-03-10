// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package externalclients

import (
	"context"

	"cloud.google.com/go/bigquery"
)

type key string

var BQClientKey key = "karte bigquery client"

// UseBQ installs the bigquery client on the context
func UseBQ(ctx context.Context, bqClient *bigquery.Client) context.Context {
	return context.WithValue(ctx, BQClientKey, bqClient)
}

// GetBQ returns the BQ client from the context
func GetBQ(ctx context.Context) *bigquery.Client {
	var zero *bigquery.Client
	out, ok := ctx.Value(BQClientKey).(*bigquery.Client)
	if ok {
		return out
	}
	return zero
}
