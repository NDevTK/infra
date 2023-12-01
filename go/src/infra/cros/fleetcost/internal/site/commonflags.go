// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package site

import "flag"

// CommonFlags are the flags common to all commands.
type CommonFlags struct {
	// http forces the use of the http protocol rather than https.
	http bool
}

func (fl *CommonFlags) Register(f *flag.FlagSet) {
	f.BoolVar(&fl.http, "http", false, "force use of http")
}

func (fl *CommonFlags) HTTP() bool {
	return fl.http
}
