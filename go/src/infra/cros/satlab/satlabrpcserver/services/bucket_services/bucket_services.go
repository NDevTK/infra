// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package bucket_services

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

type iObjectIterator interface {
	Next() (*storage.ObjectAttrs, error)
}

// IBucketServices is the interface that provide the services
// It should not contain any `Business Logic` here, because it
// is to mock the interface for testing.
type IBucketServices interface {
	// IsBucketInAsia returns boolean. Check the given bucket is in asia.
	IsBucketInAsia(ctx context.Context) (bool, error)

	// GetMilestones returns all milestones from the bucket by given board.
	GetMilestones(ctx context.Context, board string) ([]string, error)

	// GetBuilds returns all build versions from the bucket by given board and milestone.
	GetBuilds(ctx context.Context, board string, milestone int32) ([]string, error)

	// ListTestplans list all testplan json in partner bucket under a `testplans` folder
	ListTestplans(ctx context.Context) ([]string, error)
}

type IBucketClient interface {
	// GetAttrs get the bucket attributes
	GetAttrs(ctx context.Context) (*storage.BucketAttrs, error)

	// QueryObjects query objects from the bucket
	QueryObjects(ctx context.Context, q *storage.Query) iObjectIterator

	// ReadObject read the object content by the given name
	ReadObject(ctx context.Context, name string) (io.ReadCloser, error)

	// Close clean up
	Close() error
}
