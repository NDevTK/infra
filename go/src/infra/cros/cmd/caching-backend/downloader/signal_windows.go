// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build windows
// +build windows

package main

import (
	"context"
	"net/http"
	"time"
)

func cancelOnSignals(ctx context.Context, idleConns chan struct{}, svr *http.Server, gracePeriod time.Duration) context.Context {
	panic("windows not supported")
}
