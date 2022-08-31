// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// CBI corruption detection and repair logic. go/cbi-auto-recovery-dd
package cros

import (
	"context"

	"infra/cros/recovery/internal/execs"
)

// repairCbi repairs CBI contents on the DUT by writing the CBI contents stored in UFS to
// CBI EEPROM on the DUT.
// TODO(b/235000813) Implement
func repairCbi(ctx context.Context, info *execs.ExecInfo) error {
	return nil
}

// cbiIsCorrupt checks if the CBI contents on the DUT match what was previously
// stored in UFS. If not, then the CBI contents have been corrupted.
// TODO(b/235000813) Implement
func cbiIsCorrupt(ctx context.Context, info *execs.ExecInfo) error {
	return nil
}

// cbiIsPresent checks if CBI is present on the DUT using
// the ectool locatechip utility.
// TODO(b/235000813) Implement
func cbiIsPresent(ctx context.Context, info *execs.ExecInfo) error {
	return nil
}

func init() {
	execs.Register("cros_repair_cbi", repairCbi)
	execs.Register("cros_cbi_is_corrupt", cbiIsCorrupt)
	execs.Register("cros_cbi_is_present", cbiIsPresent)
}
