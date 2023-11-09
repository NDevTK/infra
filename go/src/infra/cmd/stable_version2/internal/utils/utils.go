// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"github.com/golang/protobuf/jsonpb"

	"go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	"go.chromium.org/chromiumos/infra/proto/go/device"
	sv "go.chromium.org/chromiumos/infra/proto/go/lab_platform"
)

// Unmarshaller is used to unmarshall all stable_version related data.
var (
	Unmarshaller = jsonpb.Unmarshaler{AllowUnknownFields: true}
)

// MakeCrOSSV makes a stable cros version object following format:
// https://chromium.googlesource.com/chromiumos/infra/proto/+/refs/heads/main/src/lab_platform/stable_cros_version.proto
func MakeCrOSSV(b, v string) *sv.StableCrosVersion {
	return &sv.StableCrosVersion{
		Key:     MakeStableVersionKey(b, ""),
		Version: v,
	}
}

// MakeStableVersionKey makes a key whose format follows:
// https://chromium.googlesource.com/chromiumos/infra/proto/+/refs/heads/main/src/lab_platform/stable_version.proto
func MakeStableVersionKey(buildTarget, model string) *sv.StableVersionKey {
	return &sv.StableVersionKey{
		ModelId: &device.ModelId{
			Value: model,
		},
		BuildTarget: &chromiumos.BuildTarget{
			Name: buildTarget,
		},
	}
}

// GetCrOSSVByBuildtarget find the cros stable version for a given build target.
func GetCrOSSVByBuildtarget(res []*sv.StableCrosVersion, b string) string {
	for _, c := range res {
		if c.GetKey().GetBuildTarget().GetName() == b {
			return c.GetVersion()
		}
	}
	return ""
}

// MakeSpecificCrOSSV makes a stable cros version object specific to the
// model.
func MakeSpecificCrOSSV(b, m, v string) *sv.StableCrosVersion {
	return &sv.StableCrosVersion{
		Key:     MakeStableVersionKey(b, m),
		Version: v,
	}
}

// MakeSpecificFirmwareVersion makes a firmware version specific to the model.
func MakeSpecificFirmwareVersion(b, m, v string) *sv.StableFirmwareVersion {
	return &sv.StableFirmwareVersion{
		Key:     MakeStableVersionKey(b, m),
		Version: v,
	}
}