// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build windows
// +build windows

package main

import (
	"context"
)

func cancelOnSignals(ctx context.Context) context.Context {
	panic("windows not supported")
}
