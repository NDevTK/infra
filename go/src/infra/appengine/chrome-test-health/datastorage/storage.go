// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This file contains a list of interfaces that our data clients want to implement.
package datastorage

import "context"

type QueryFilter struct {
	Field    string
	Operator string
	Value    interface{}
}

// IDataClient interface has necessary data related functions that need
// to be implemented by the struct using it. For example, a datastore
// client struct implementing this interface would need to have datastore
// specific implementation of the functions mentioned in this interface.
type IDataClient interface {
	// Gets a single data item of the specified type, matching the specified key.
	// Optionally takes additional arguments that can be used for specific
	// implementations of this interface.
	//
	// Returns an error if the data item cannot be found.
	Get(
		ctx context.Context,
		result interface{},
		dataType string,
		key interface{},
		options ...interface{},
	) error

	// QueryOne function performs a data storage query with the given filters,
	// order, and options. Returns the first data item from the query result,
	// or an error if no data items are found.
	//
	// Filters should be tuples of property name, operator, and value.
	//
	// The order argument is used to specify the order of the query results.
	//
	// The options argument can be used for specific implementations of this
	// function.
	//
	// The result is reflected in the "result" argument.
	QueryOne(
		ctx context.Context,
		result interface{},
		dataType string,
		queryFilters []QueryFilter,
		order interface{},
		options ...interface{},
	) error
}
