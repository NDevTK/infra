// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package provider provides a generic template
// for new Bots Provider Interfaces.
package provider

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/botsregulator/internal/util"
)

// BPI is a generic Provider interface.
// Future Providers need to satisfy this interface.
type BPI interface {
	UpdateConfig(ctx context.Context, hostnames []string, cfID string) error
}

// NewProviderFromEnv creates a provider.BPI based on the server running environment.
// The provider is responsible for the actual implementation.
// Providers currently supported are GCE Provider and Satlab(WIP).
func NewProviderFromEnv(ctx context.Context, host string) (BPI, error) {
	var bc BPI
	var err error
	switch util.GetEnv() {
	case util.GCP:
		bc, err = NewGCEPClient(ctx, host)
	case util.Satlab:
		err = errors.New("Satlab flow not implemented")
	default:
		panic("unrecognized running environment")
	}
	if err != nil {
		return nil, err
	}
	return bc, nil
}
