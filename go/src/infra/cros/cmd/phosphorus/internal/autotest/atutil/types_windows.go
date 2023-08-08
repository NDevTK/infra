// Copyright 2018 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build windows
// +build windows

package atutil

func (r *Result) Signaled() bool {
	panic("not supported on windows")
}
