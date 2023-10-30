// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package bucket_services

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/collection"
)

type getDataFromObject = func(obj *storage.ObjectAttrs) (string, error)

// bucketClient is the client that control how to deal with the
// `storage.client`
type bucketClient struct {
	// a client for interacting with Google Cloud Storage
	client *storage.Client
	// a bucketName which bucket we want to get the information
	bucketName string
}

// BucketConnector is an object for connecting the GCS bucket storage.
type BucketConnector struct {
	client IBucketClient
}

// QueryObjects query objects from the bucket
func (b *bucketClient) QueryObjects(ctx context.Context, q *storage.Query) iObjectIterator {
	iter := b.client.Bucket(b.bucketName).Objects(ctx, q)
	return iter
}

// ReadObject read the object content by the given name
func (b *bucketClient) ReadObject(ctx context.Context, name string) (io.ReadCloser, error) {
	return b.client.Bucket(b.bucketName).Object(name).NewReader(ctx)
}

// GetAttrs get the bucket attributes
func (b *bucketClient) GetAttrs(ctx context.Context) (*storage.BucketAttrs, error) {
	return b.client.Bucket(b.bucketName).Attrs(ctx)
}

// Close to close the client connection.
func (b *bucketClient) Close() error {
	return b.client.Close()
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

	c := &bucketClient{
		client:     client,
		bucketName: bucketName,
	}

	return &BucketConnector{
		client: c,
	}, nil
}

func getPartialObjectPath(obj *storage.ObjectAttrs) (string, error) {
	return obj.Prefix, nil
}

// IsBucketInAsia returns boolean. Check the given bucket is in asia.
func (b *BucketConnector) IsBucketInAsia(ctx context.Context) (bool, error) {
	attrs, err := b.client.GetAttrs(ctx)
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
	q := &storage.Query{Prefix: prefix, Delimiter: "-"}
	// We don't need other fields here because
	// the field `Prefix` we need already included.
	q.SetAttrSelection([]string{"Name"})
	rawData := b.client.QueryObjects(ctx, q)
	data, err := collection.Collect(rawData.Next, getPartialObjectPath)
	if err != nil {
		return nil, err
	}

	res := make([]string, 0, len(data))
	for _, item := range data {
		res = append(res, item[len(prefix):len(item)-1])
	}

	return res, nil
}

// GetBuilds returns all build versions from the bucket by given board and milestone.
//
// string board the board name we want to use as a filter.
// string milestone the milestone we want to use as a filter.
func (b *BucketConnector) GetBuilds(ctx context.Context, board string, milestone int32) ([]string, error) {
	releasePrefix := fmt.Sprintf("%s-release/R%d-", board, milestone)
	q := &storage.Query{Prefix: releasePrefix, Delimiter: "/"}
	// We don't need other fields here because
	// the field `Prefix` we need already included.
	q.SetAttrSelection([]string{"Name"})
	releaseRawData := b.client.QueryObjects(ctx, q)
	releaseData, err := collection.Collect(releaseRawData.Next, getPartialObjectPath)
	if err != nil {
		return nil, err
	}

	localPrefix := fmt.Sprintf("%s-local/R%d-", board, milestone)
	q = &storage.Query{Prefix: localPrefix, Delimiter: "/"}
	// We don't need other fields here because
	// the field `Prefix` we need already included.
	q.SetAttrSelection([]string{"Name"})
	localRawData := b.client.QueryObjects(ctx, q)
	localData, err := collection.Collect(localRawData.Next, getPartialObjectPath)
	if err != nil {
		return nil, err
	}

	var res []string

	for _, item := range releaseData {
		res = append(res, item[len(releasePrefix):len(item)-1])
	}

	for _, item := range localData {
		res = append(res, item[len(localPrefix):len(item)-1])
	}

	return res, nil
}

var DefaultPageSize = 10

// ListTestplans list all testplan json in partner bucket under a `testplans` folder
func (b *BucketConnector) ListTestplans(ctx context.Context) ([]string, error) {
	d := "testplans/"
	q := &storage.Query{Prefix: d, Delimiter: "*.json"}
	q.SetAttrSelection([]string{"Name"})
	rawData := b.client.QueryObjects(ctx, q)

	res := []string{}
	for {
		item, err := rawData.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		if name, ok := strings.CutPrefix(item.Name, d); ok && strings.HasSuffix(name, ".json") {
			res = append(res, name)
		}
	}

	return res, nil
}
