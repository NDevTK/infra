// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ethernethook

import (
	"context"

	"cloud.google.com/go/storage"
	"go.chromium.org/luci/common/errors"
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

// LsState stores the current iteration state as we're traversing the results of Ls.
type LsState struct {
	Attrs *storage.ObjectAttrs
	Err   error
}

// Reset resets the values of all the fields to zero.
func (s *LsState) Reset() {
	if s != nil {
		var zero LsState
		*s = zero
	}
}

// LsResult is an iterator over objects.
type LsResult func(state *LsState) bool

// Ls iterates over items in Google Storage beginning with a prefix.
func (e *extendedGSClient) Ls(ctx context.Context, bucket string, query *storage.Query) LsResult {
	objectIterator := e.Bucket(bucket).Objects(ctx, query)
	res := func(state *LsState) bool {
		state.Reset()
		objectAttrs, err := objectIterator.Next()
		if err != nil {
			state.Err = err
			return false
		}
		state.Attrs = objectAttrs
		return true
	}
	return res
}
