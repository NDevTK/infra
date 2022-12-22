// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmlab

import (
	"fmt"

	"infra/libs/vmlab/api"
	"infra/libs/vmlab/internal/instance/gcloud"
)

// NewInstanceApi serves as the entry point to the vmlab library by returning an
// api.InstanceApi for the given provider.
func NewInstanceApi(pid api.ProviderId) (api.InstanceApi, error) {
	if pid == api.ProviderId_GCLOUD {
		return gcloud.New()
	}
	return nil, fmt.Errorf("provider %v is not implemented", pid)
}
