// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package bucket_services

import (
	"context"
	"fmt"
	"strings"

	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/collection"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// BucketConnector is an object for connecting the GCS bucket storage.
type BucketConnector struct {
	// a client for interacting with Google Cloud Storage
	client *storage.Client
	// a bucketName which bucket we want to get the information
	bucketName string
}

// New sets up the storage client and returns a BucketConnector.
// The service account is set in the global environment.
//
// string bucketName config which bucket we want to connect with in later functions.
func New(ctx context.Context, bucketName string) (IBucketServices, error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(site.GetServiceAccountPath()))
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %w", err)
	}

	return &BucketConnector{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// getPartialObjectPath returns the partial blobs for given prefix and delimiter.
//
// string prefix is the prefix filter to query objects.
// string delimiter Objects whose names, aside from the prefix, contain delimiter will have their name,
// truncated after the delimiter.
func (b *BucketConnector) getPartialObjectPath(ctx context.Context, prefix string, delimiter string) ([]string, error) {
	bucket := b.client.Bucket(b.bucketName)
	iter := bucket.Objects(ctx, &storage.Query{Prefix: prefix, Delimiter: delimiter})

	return collection.Collect(iter.Next, func(obj *storage.ObjectAttrs) (string, error) {
		return obj.Prefix, nil
	})
}

// IsBucketInAsia returns boolean. Check the given bucket is in asia.
func (b *BucketConnector) IsBucketInAsia(ctx context.Context) (bool, error) {
	bucket := b.client.Bucket(b.bucketName)
	attrs, err := bucket.Attrs(ctx)
	if err != nil {
		return false, err
	}

	return strings.Index(strings.ToLower(attrs.Location), "asia") != -1, nil
}

// GetMilestones returns all milestones from the bucket by given board.
//
// string board the board name we want to use as a filter.
func (b *BucketConnector) GetMilestones(ctx context.Context, board string) ([]string, error) {
	prefix := fmt.Sprintf("%s-release/R", board)
	rawData, err := b.getPartialObjectPath(ctx, prefix, "-")
	if err != nil {
		return nil, err
	}

	res := make([]string, len(rawData))
	for idx, item := range rawData {
		res[idx] = item[len(prefix) : len(item)-1]
	}

	return res, nil
}

// GetBuilds returns all build versions from the bucket by given board and milestone.
//
// string board the board name we want to use as a filter.
// string milestone the milestone we want to use as a filter.
func (b *BucketConnector) GetBuilds(ctx context.Context, board string, milestone int32) ([]string, error) {
	releasePrefix := fmt.Sprintf("%s-release/R%d-", board, milestone)
	releaseRawData, err := b.getPartialObjectPath(ctx, releasePrefix, "/")
	if err != nil {
		return nil, err
	}

	localPrefix := fmt.Sprintf("%s-local/R%d-", board, milestone)
	localRawData, err := b.getPartialObjectPath(ctx, localPrefix, "/")

	var res []string

	for _, item := range releaseRawData {
		res = append(res, item[len(releasePrefix):len(item)-1])
	}

	for _, item := range localRawData {
		res = append(res, item[len(localPrefix):len(item)-1])
	}

	return res, nil
}

// Close to close the client connection.
func (b *BucketConnector) Close() error {
	return b.client.Close()
}
