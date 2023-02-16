// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"infra/libs/vmlab/api"
)

var ConfigMapping = map[string]*BuildtinConfig{
	"cts-prototype": {
		ProviderId: api.ProviderId_GCLOUD,
		GcloudConfig: api.Config_GCloudBackend{
			Project:        "betty-cloud-prototype",
			Zone:           "us-west2-a",
			MachineType:    "n2-standard-4",
			InstancePrefix: "ctsprototype-",
			PublicIp:       true,
		},
	},
}
