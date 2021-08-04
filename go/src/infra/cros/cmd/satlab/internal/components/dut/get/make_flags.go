// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package get

// Flagmap is a map from flags to their values.
type flagmap = map[string][]string

// MakeShivasFlags creates a map for the flags inherited from shivas.
func makeShivasFlags(c *getDUT) flagmap {
	out := make(flagmap)

	if len(c.zones) != 0 {
		out["zone"] = c.zones
	}
	if len(c.racks) != 0 {
		out["rack"] = c.racks
	}
	if len(c.machines) != 0 {
		out["machine"] = c.machines
	}
	if len(c.prototypes) != 0 {
		out["prototype"] = c.prototypes
	}
	if len(c.servos) != 0 {
		out["servo"] = c.servos
	}
	if len(c.servotypes) != 0 {
		out["servotype"] = c.servotypes
	}
	if len(c.switches) != 0 {
		out["switch"] = c.switches
	}
	if len(c.rpms) != 0 {
		out["rpms"] = c.rpms
	}
	if len(c.pools) != 0 {
		out["pools"] = c.pools
	}
	if c.wantHostInfoStore {
		out["host-info-store"] = []string{}
	}
	return out
}
