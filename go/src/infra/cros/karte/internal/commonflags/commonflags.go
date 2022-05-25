// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commonflags

import (
	"flag"
	"fmt"

	"go.chromium.org/luci/auth"

	"infra/cros/karte/client"
)

// Flags are the flags shared by most Karte commands.
type Flags struct {
	// Environment selection
	Env string
}

// Register registers the common flags.
func (c *Flags) Register(f *flag.FlagSet) {
	if c == nil {
		panic("common flags cannot be nil")
	}
	if f == nil {
		panic("flagset cannot be nil")
	}
	f.StringVar(&c.Env, "env", "local", `choose the Karte service, valid values are local|dev|prod`)
}

// MustSelectKarteConfig selects the a Karte config corresponding to the environment flag.
func (f *Flags) MustSelectKarteConfig(o auth.Options) *client.Config {
	switch f.Env {
	case "", "local":
		return client.LocalConfig(o)
	case "dev":
		return client.DevConfig(o)
	case "prod":
		return client.ProdConfig(o)
	default:
		panic(fmt.Sprintf("unrecognized environment %q", f.Env))
	}
}
