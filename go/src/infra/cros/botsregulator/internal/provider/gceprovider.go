// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package provider

import (
	"context"

	"go.chromium.org/luci/common/logging"
	gcepAPI "go.chromium.org/luci/gce/api/config/v1"

	"infra/cros/botsregulator/internal/util"
)

// gcepProvider is the GCE Provider implementation of the Provider interface.
type gcepProvider struct {
	// GCE Provider configured PRPC client.
	ic gcepAPI.ConfigurationClient
	// The prefix of the config to update.
	cfID string
}

// NewGCEPClient returns a new gcepClient instance.
func NewGCEPClient(ctx context.Context, host string, cfID string) (*gcepProvider, error) {
	pc, err := util.RawPRPCClient(ctx, host)
	if err != nil {
		return nil, err
	}
	g := &gcepProvider{
		ic:   gcepAPI.NewConfigurationPRPCClient(pc),
		cfID: cfID,
	}
	return g, nil
}

// UpdateConfig is called as BPI.UpdateConfig and
// is responsible for orchestrating the config update.
func (g *gcepProvider) UpdateConfig(ctx context.Context, hns []string) error {
	logging.Infof(ctx, "hello from UpdateConfig!")
	return nil
}
