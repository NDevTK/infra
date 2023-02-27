// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmlab

import (
	"fmt"

	"infra/libs/vmlab/api"
	"infra/libs/vmlab/internal/image/cloudsdk"
)

// NewImageApi serves as the entry point to the vmlab library for api.ImageApi
func NewImageApi(pid api.ProviderId) (api.ImageApi, error) {
	if pid == api.ProviderId_CLOUDSDK {
		return cloudsdk.New()
	}
	return nil, fmt.Errorf("provider %v is not implemented", pid)
}
