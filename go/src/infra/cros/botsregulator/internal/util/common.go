// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package util exports useful things.
package util

import (
	"os"

	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// Default flag values.
	UFSDev   string = "staging.ufs.api.cr.dev"
	GCEPDev  string = "gce-provider-dev.appspot.com"
	ConfigID string = "cloudbots-dev"

	// Common prefix for machineLSE keys.
	MachineLSEPrefix string = "machineLSEs/"
	// Common prefix for schedulingUnits keys.
	SchedulingUnitsPrefix string = "schedulingunits/"
)

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
