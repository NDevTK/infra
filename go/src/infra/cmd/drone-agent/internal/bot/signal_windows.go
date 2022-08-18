// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build windows
// +build windows

package bot

// TerminateOrKill implements Bot.
func (b realBot) TerminateOrKill() error {
	panic("windows not supported")
}
