// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build windows

package main

import (
	"context"
)

func notifySIGTERM(ctx context.Context) (_ context.Context, stop context.CancelFunc) {
	panic("windows not supported")
}
