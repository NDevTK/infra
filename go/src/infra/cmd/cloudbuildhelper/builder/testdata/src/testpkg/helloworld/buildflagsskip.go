// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build windows
// +build windows

package main

import "testpkg/nope"

// B3 exists to make golint happy.
const B3 = nope.A
