// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ethernethook

import (
	"context"
	"strings"

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
			if err == iterator.Done {
				return nil, done, err
			}
			if strings.Contains(err.Error(), "no more items") {
				return nil, done, err
			}
			return nil, invalid, err
		}
		return objectAttrs, keepGoing, nil
	}
	return res
}

// LsSync synchronously gets objects.
func (e *extendedGSClient) LsSync(ctx context.Context, bucket string, prefix string) ([]storage.ObjectAttrs, error) {
	it := e.Ls(ctx, bucket, prefix)
	var out []storage.ObjectAttrs
	for {
		objectAttrs, status, err := it()
		switch status {
		case invalid:
			return nil, err
		case done:
			return out, nil
		}
		out = append(out, *objectAttrs)
	}
}
