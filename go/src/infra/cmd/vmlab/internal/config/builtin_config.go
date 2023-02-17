// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"infra/libs/vmlab/api"
)

type BuiltinConfig struct {
	ProviderId api.ProviderId
	// TODO(fqj): replace to a different type outside of api.
	GcloudConfig api.Config_GCloudBackend
}
