// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package site

import (
	"errors"
	"flag"
)

// CommonFlags are the flags common to all commands.
type CommonFlags struct {
	// local forces the use of the local protocol rather than https,
	// which makes sense because you're talking to a local service.
	local bool
	dev   bool
}

// Register the common flags.
func (fl *CommonFlags) Register(f *flag.FlagSet) {
	f.BoolVar(&fl.local, "local", false, "talk to the local project.")
	f.BoolVar(&fl.dev, "dev", false, "use the dev cloud run project")
}

// HTTP returns whether to use HTTP or HTTPS (default).
func (fl *CommonFlags) HTTP() bool {
	return fl.local
}

// Host returns the host to contact.
func (fl *CommonFlags) Host() (string, error) {
	switch {
	case fl.dev && !fl.local:
		return "fleet-cost-dev-see5vh56pa-uc.a.run.app", nil
	case !fl.dev && fl.local:
		return "localhost:8800", nil
	case !fl.dev && !fl.local:
		return "", errors.New("prod not deployed yet")
	default:
		return "", errors.New("-dev and -local are alternatives")
	}
}
