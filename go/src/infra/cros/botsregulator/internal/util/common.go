// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package util exports useful things.
package util

import (
	"context"
	"net/http"
	"strings"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/auth"

	ufspb "infra/unifiedfleet/api/v1/models"
)

const (
	// Default flag values.
	UFSDev   string = "staging.ufs.api.cr.dev"
	GCEPDev  string = "gce-provider-dev.appspot.com"
	ConfigID string = "cloudbots-dev"

	// Common prefix for machineLSE keys.
	machineLSEPrefix string = "machineLSEs/"
)

// RawPRPCClient returns a generic PRPC Client.
func RawPRPCClient(ctx context.Context, host string) (*prpc.Client, error) {
	t, err := auth.GetRPCTransport(ctx, auth.AsSelf, auth.WithScopes(auth.CloudOAuthScopes...))
	if err != nil {
		return nil, errors.Annotate(err, "could not create http.RoundTripper").Err()
	}
	c := &prpc.Client{
		C:    &http.Client{Transport: t},
		Host: host,
		Options: &prpc.Options{
			UserAgent: "bots-regulator/0.1.0",
		},
	}
	return c, nil
}

// CutHostnames cuts "machineLSEs/" prefix from DUT names.
func CutHostnames(lses []*ufspb.MachineLSE) ([]string, error) {
	hns := make([]string, len(lses))
	for i, lse := range lses {
		hn, ok := strings.CutPrefix(lse.GetName(), machineLSEPrefix)
		if !ok {
			return nil, errors.Reason("could not parse DUT hostname: %v", lse).Err()
		}
		hns[i] = hn
	}
	return hns, nil
}
