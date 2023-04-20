// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/satlab/satlab/internal/site"
)

// ShivasDeleteDUT holds the delete DUT flags inherited from shivas.
type shivasDeleteDUT struct {
	subcommands.CommandRunBase

	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags
}

// RegisterShivasFlags registers the shivas flags.
func registerShivasFlags(c *deleteDUT) {
	c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
	c.envFlags.Register(&c.Flags)
	c.commonFlags.Register(&c.Flags)
}
