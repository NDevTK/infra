// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package site

import "flag"

// CommonFlags are the flags common to all commands.
type CommonFlags struct {
	// http forces the use of the http protocol rather than https.
	http bool
	dev  bool
}

// Register the common flags.
func (fl *CommonFlags) Register(f *flag.FlagSet) {
	f.BoolVar(&fl.http, "http", false, "force use of http")
	f.BoolVar(&fl.dev, "dev", false, "use the dev cloud run project")
}

// HTTP returns whether to use HTTP or HTTPS (default).
func (fl *CommonFlags) HTTP() bool {
	return fl.http
}

// Host returns the host to contact.
func (fl *CommonFlags) Host() (string, error) {
	switch {
	case fl.dev:
		return "fleet-cost-dev-see5vh56pa-uc.a.run.app", nil
	case !fl.dev && fl.http:
		return "localhost:8800", nil
	case !fl.dev && !fl.http:
		return "", nil
	}
	panic("internal error in .../internal/site/commonflags.go: impossible")
}
