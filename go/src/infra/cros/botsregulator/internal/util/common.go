// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package util exports useful things.
package util

import (
	"context"
	"net/http"
	"os"
	"strings"

	"google.golang.org/protobuf/types/known/emptypb"

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

type Set = map[string]*emptypb.Empty

// NewStringSet creates a Set.
// We cannot use luci/common/data/stringset/stringset.go here
// since the types mismatch.
func NewStringSet(s []string) Set {
	m := make(Set, len(s))
	for _, k := range s {
		m[k] = &emptypb.Empty{}
	}
	return m
}

// Environments supported.
type environment = int

const (
	GCP environment = iota
	Satlab
)

// GetEnv determines where the server is running.
// TODO(b/329682593): Support more environments.
func GetEnv() environment {
	// K_SERVICE environment variable is set on Cloud Run instances.
	// 7446b7c04264699f6a1c4990a3126e0769081999:luci/luci-go/server/server.go;l=688
	if os.Getenv("K_SERVICE") == "" {
		return Satlab
	}
	return GCP
}
