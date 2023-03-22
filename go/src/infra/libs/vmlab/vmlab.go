// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmlab

import (
	"fmt"

	"infra/libs/vmlab/api"
	"infra/libs/vmlab/internal/instance/gcloud"
	vmleaser "infra/libs/vmlab/internal/instance/vm_leaser"
)

// NewInstanceApi serves as the entry point to the vmlab library by returning an
// api.InstanceApi for the given provider.
func NewInstanceApi(pid api.ProviderId) (api.InstanceApi, error) {
	switch pid {
	case api.ProviderId_GCLOUD:
		return gcloud.New()
	case api.ProviderId_VM_LEASER:
		return vmleaser.New()
	default:
		return nil, fmt.Errorf("provider %v is not implemented", pid)
	}
}
