// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ethernethook

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/api/iterator"
)

// IteratorStatus is an enum that makes it easy to distinguish between iterators that
// hit the end of their iteration successfully and genuine errors.
type IteratorStatus int32

const (
	invalid   = IteratorStatus(0)
	keepGoing = IteratorStatus(1)
	done      = IteratorStatus(2)
)

// extendedGSClient is an extended storage client.
type extendedGSClient struct {
	*storage.Client
}

// NewExtendedGSClient takes a Google Storage client and returns a wrapped version.
//
// A wrapped Google Storage client, much like the raw client it wraps, is intended to be
// a long-lived object. For that reason, we return an error value as well to make error handling
// more obvious at the call site.
func NewExtendedGSClient(client *storage.Client) (*extendedGSClient, error) {
	if client == nil {
		return nil, errors.New("new extended gs client: wrapped client cannot be nil")
	}
	return &extendedGSClient{client}, nil
}

// LsResult is an iterator over objects.
type LsResult func() (*storage.ObjectAttrs, IteratorStatus, error)

// Ls iterates over items in Google Storage beginning with a prefix.
func (e *extendedGSClient) Ls(ctx context.Context, bucket string, prefix string) LsResult {
	query := &storage.Query{
		Delimiter: "/",
		Prefix:    prefix,
	}
	objectIterator := e.Bucket(bucket).Objects(ctx, query)
	res := func() (*storage.ObjectAttrs, IteratorStatus, error) {
		objectAttrs, err := objectIterator.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				return nil, done, err
			}
			return nil, invalid, errors.Annotate(err, `looking at path %q`, fmt.Sprintf("gs://%s/%s", bucket, prefix)).Err()
		}
		return objectAttrs, keepGoing, nil
	}
	return res
}

// LsSync synchronously gets objects.
func (e *extendedGSClient) LsSync(ctx context.Context, bucket string, prefix string) ([]*storage.ObjectAttrs, error) {
	it := e.Ls(ctx, bucket, prefix)
	var out []*storage.ObjectAttrs
	for {
		objectAttrs, status, err := it()
		if err != nil {
			if status == done {
				return out, nil
			}
			return nil, err
		}
		out = append(out, objectAttrs)
	}
}

// ToGSURL converts a storage object to a Google Storage URL.
func ToGSURL(bucket string, attrs *storage.ObjectAttrs) (string, error) {
	if err := validateToGSUrl(bucket, attrs); err != nil {
		return "", err
	}
	if bucket == "" {
		bucket = attrs.Bucket
	}
	if attrs.Prefix != "" {
		return fmt.Sprintf("gs://%s/%s", bucket, attrs.Prefix), nil
	}
	if attrs.Name != "" {
		return fmt.Sprintf("gs://%s/%s", bucket, attrs.Name), nil
	}
	return "", errors.New("object has no name and no prefix")
}

func validateToGSUrl(bucket string, attrs *storage.ObjectAttrs) error {
	if attrs == nil {
		return errors.New("attrs cannot be nil")
	}
	if bucket != "" && attrs.Bucket != "" {
		return errors.New("bucket %q and attrs.Bucket %q cannot both be set")
	}
	if bucket == "" && attrs.Bucket == "" {
		return errors.New("bucket and attrs.Bucket cannot both be empty")
	}
	return nil
}
