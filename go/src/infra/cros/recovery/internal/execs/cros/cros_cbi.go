// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// CBI corruption detection and repair logic. go/cbi-auto-recovery-dd
package cros

import (
	"context"
	"infra/cros/recovery/internal/components/cros/cbi"
	"infra/cros/recovery/internal/execs"

	"go.chromium.org/luci/common/errors"
)

// repairCbi repairs CBI contents on the DUT by writing the CBI contents stored
// in UFS to CBI EEPROM on the DUT.
// TODO(b/235000813) Implement
func repairCBI(ctx context.Context, info *execs.ExecInfo) error {
	return nil
}

// cbiIsCorrupt checks if the CBI contents on the DUT match what was previously
// stored in UFS. If not, then the CBI contents have been corrupted. Throws an
// error if no previously stored CBI contents are found.
func cbiIsCorrupt(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetChromeos().GetCbi() == nil {
		return errors.Reason("no previous CBI contents were found in UFS. Unable to determine if CBI is corrupt.").Err()
	}

	cbiLocation, err := cbi.GetCBILocation(ctx, info.NewRunner(info.GetDut().Name))
	if err != nil {
		return errors.Annotate(err, "CBI is corrupt").Err()
	}

	dutCBI, err := cbi.ReadCBIContents(ctx, info.NewRunner(info.GetDut().Name), cbiLocation)
	if err != nil {
		return errors.Annotate(err, "CBI is corrupt").Err()
	}

	if info.GetChromeos().GetCbi().RawContents != dutCBI.RawContents {
		return nil
	}
	return errors.Reason("CBI is valid").Err()
}

// cbiIsPresent checks if CBI contents are found on the DUT.
func cbiIsPresent(ctx context.Context, info *execs.ExecInfo) error {
	cbiLocation, err := cbi.GetCBILocation(ctx, info.NewRunner(info.GetDut().Name))
	if err != nil {
		return errors.Annotate(err, "CBI is present").Err()
	}
	if cbiLocation == nil {
		return errors.Reason("CBI is present: no CBI contents were found on the DUT, but encountered no error. This shouldn't ever happen. Please submit a bug.").Err()
	}
	return nil
}

func init() {
	execs.Register("cros_repair_cbi", repairCBI)
	execs.Register("cros_cbi_is_corrupt", cbiIsCorrupt)
	execs.Register("cros_cbi_is_present", cbiIsPresent)
}
