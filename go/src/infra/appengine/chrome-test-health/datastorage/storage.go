// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This file contains a list of interfaces that our data clients want to implement.
package datastorage

import "context"

// IDataClient interface has necessary data related functions that need
// to be implemented by the struct using it. For example, a datastore
// client struct implementing this interface would need to have datastore
// specific implementation of the functions mentioned in this interface.
type IDataClient interface {
	Get(ctx context.Context, result interface{}, dataType string, key interface{}, options ...interface{}) error
}
