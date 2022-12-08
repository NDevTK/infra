// Copyright 2022 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package provides a wrapper to specify local dev options.
package dev

import (
	"context"
)

// localDevKeyType is a unique type for a context key.
type localDevKeyType string

const (
	localDevOptionsKey localDevKeyType = "local_developing_options"
)

// WithDevOptions sets DevOptions to the context.
func WithDevOptions(ctx context.Context, devOptions any) context.Context {
	if devOptions == nil {
		panic("logger is not provided")
	}
	return context.WithValue(ctx, localDevOptionsKey, devOptions)
}

// ActiveLocalDevOption represent interface to tell if local development is active.
type ActiveLocalDevOption interface {
	// Specify if client is active.
	IsActive() bool
}

// IsActive specifies if local dev option is active.
func IsActive(ctx context.Context) bool {
	if o, ok := ctx.Value(localDevOptionsKey).(ActiveLocalDevOption); ok {
		return o.IsActive()
	}
	return false
}
