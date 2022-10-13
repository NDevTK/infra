// Copyright 2022 The ChromiumOS Authors. All rights reserved.
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

// ufsContainsCBIContents checks if any CBI Contents were previously stored for
// this DUT in UFS.
func ufsContainsCBIContents(ctx context.Context, info *execs.ExecInfo) error {
	if len(info.GetChromeos().GetCbi().GetRawContents()) == 0 {
		return errors.Reason("UFS contains CBI contents: no previous CBI contents were found in UFS").Err()
	}
	return nil
}

// cbiContentsDoNotMatch checks if the CBI contents on the DUT do not match what
// was previously stored in UFS. If they do not match, then the CBI contents
// have been corrupted. Throws an error if no previously stored CBI contents are
// found.
func cbiContentsDoNotMatch(ctx context.Context, info *execs.ExecInfo) error {
	cbiLocation, err := cbi.GetCBILocation(ctx, info.NewRunner(info.GetDut().Name))
	if err != nil {
		return errors.Annotate(err, "CBI contents do not match").Err()
	}

	dutCBI, err := cbi.ReadCBIContents(ctx, info.NewRunner(info.GetDut().Name), cbiLocation)
	if err != nil {
		return errors.Annotate(err, "CBI contents do not match").Err()
	}

	if info.GetChromeos().GetCbi().GetRawContents() != dutCBI.GetRawContents() {
		return nil
	}
	return errors.Reason("CBI contents do not match: CBI contents on DUT match the contents stored in UFS").Err()
}

// cbiIsPresent checks if CBI contents are found on the DUT.
func cbiIsPresent(ctx context.Context, info *execs.ExecInfo) error {
	cbiLocation, err := cbi.GetCBILocation(ctx, info.NewRunner(info.GetDut().Name))
	if err != nil {
		return errors.Annotate(err, "CBI is present").Err()
	}
	if cbiLocation == nil {
		return errors.Reason("CBI is present: no CBI contents were found on the DUT, but encountered no error - please submit a bug").Err()
	}
	return nil
}

func init() {
	execs.Register("cros_repair_cbi", repairCBI)
	execs.Register("cros_cbi_contents_do_not_match", cbiContentsDoNotMatch)
	execs.Register("cros_ufs_contains_cbi_contents", ufsContainsCBIContents)
	execs.Register("cros_cbi_is_present", cbiIsPresent)
}
