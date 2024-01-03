// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by mockery v2.36.0. DO NOT EDIT.

package mocks

import (
	context "context"
	datastorage "infra/appengine/chrome-test-health/datastorage"

	mock "github.com/stretchr/testify/mock"
)

// IDataClient is an autogenerated mock type for the IDataClient type
type IDataClient struct {
	mock.Mock
}

// BatchPut provides a mock function with given fields: ctx, entities, keys
func (_m *IDataClient) BatchPut(ctx context.Context, entities interface{}, keys interface{}) error {
	ret := _m.Called(ctx, entities, keys)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, interface{}) error); ok {
		r0 = rf(ctx, entities, keys)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, result, dataType, key, options
func (_m *IDataClient) Get(ctx context.Context, result interface{}, dataType string, key interface{}, options ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, ctx, result, dataType, key)
	_ca = append(_ca, options...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, string, interface{}, ...interface{}) error); ok {
		r0 = rf(ctx, result, dataType, key, options...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Query provides a mock function with given fields: ctx, result, dataType, queryFilters, order, limit, options
func (_m *IDataClient) Query(ctx context.Context, result interface{}, dataType string, queryFilters []datastorage.QueryFilter, order interface{}, limit int, options ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, ctx, result, dataType, queryFilters, order, limit)
	_ca = append(_ca, options...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, string, []datastorage.QueryFilter, interface{}, int, ...interface{}) error); ok {
		r0 = rf(ctx, result, dataType, queryFilters, order, limit, options...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewIDataClient creates a new instance of IDataClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIDataClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *IDataClient {
	mock := &IDataClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
