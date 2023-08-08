// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package btpeer

import (
	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/tlw"
)

// activeHost finds active host related to the executed plan.
func activeHost(resource string, chromeos *tlw.ChromeOS) (*tlw.BluetoothPeer, error) {
	for _, btp := range chromeos.GetBluetoothPeers() {
		if btp.GetName() == resource {
			return btp, nil
		}
	}
	return nil, errors.Reason("active host: host not found").Err()
}
