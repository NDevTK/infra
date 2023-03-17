// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// CBI corruption detection and repair logic. go/cbi-auto-recovery-dd
package cros

import (
	"context"
	"infra/cros/recovery/internal/components/cros/cbi"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"

	"go.chromium.org/luci/common/errors"
)

// restoreCBIContentsFromUFS restores CBI contents on the DUT by writing the CBI contents stored
// in UFS to CBI EEPROM on the DUT.
func restoreCBIContentsFromUFS(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.NewRunner(info.GetDut().Name)
	if info.GetChromeos().GetCbi() == nil {
		return errors.Reason("restore CBI contents from UFS: no previous CBI contents were found in UFS").Err()
	}

	cbiLocation, err := cbi.GetCBILocation(ctx, runner)
	if err != nil {
		return errors.Annotate(err, "restore CBI contents from UFS").Err()
	}

	err = cbi.WriteCBIContents(ctx, runner, cbiLocation, info.GetChromeos().GetCbi())
	return errors.Annotate(err, "restore CBI contents from UFS").Err()
}

// invalidateCBICache clears the current CBI cache to ensure that any existing
// CBI contents are up to date. Will throw an error if something unexpected occurs,
// but should otherwise always return nil.
func invalidateCBICache(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.NewRunner(info.GetDut().Name)
	err := cbi.InvalidateCBICache(ctx, runner)
	return errors.Annotate(err, "invalidate CBI cache").Err()
}

// ufsContainsCBIContents returns nil if CBI Contents were previously stored for
// this DUT in UFS.
func ufsContainsCBIContents(ctx context.Context, info *execs.ExecInfo) error {
	if len(info.GetChromeos().GetCbi().GetRawContents()) == 0 {
		return errors.Reason("UFS contains CBI contents: no previous CBI contents were found in UFS").Err()
	}
	return nil
}

// ufsDoesNotContainCBIContents returns nil if CBI Contents were NOT previously
// stored for this DUT in UFS.
func ufsDoesNotContainCBIContents(ctx context.Context, info *execs.ExecInfo) error {
	if len(info.GetChromeos().GetCbi().GetRawContents()) != 0 {
		return errors.Reason("UFS does not contain CBI contents: previous CBI contents were found in UFS").Err()
	}
	return nil
}

// cbiContentsMatch checks if the CBI contents on the DUT match what
// was previously stored in UFS.
func cbiContentsMatch(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.NewRunner(info.GetDut().Name)
	dutCBI, err := cbi.GetCBIContents(ctx, runner)
	if err != nil {
		return errors.Annotate(err, "CBI contents match").Err()
	}

	if info.GetChromeos().GetCbi().GetRawContents() == dutCBI.GetRawContents() {
		log.Debugf(ctx, "CBI contents match: CBI contents on the DUT match the CBI contents stored in UFS.\nCBI contents: %s", dutCBI.GetRawContents())
		return nil
	}
	log.Debugf(ctx, "CBI contents match: CBI contents on DUT: %s\nCBI contents in UFS: %s", dutCBI.GetRawContents(), info.GetChromeos().GetCbi().GetRawContents())
	return errors.Reason("CBI contents match: CBI contents on the DUT do not match the CBI contents stored in UFS").Err()
}

// cbiContentsDoNotMatch checks if the CBI contents on the DUT do not match what
// was previously stored in UFS. If they do not match, then the CBI contents
// have been corrupted.
func cbiContentsDoNotMatch(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.NewRunner(info.GetDut().Name)
	dutCBI, err := cbi.GetCBIContents(ctx, runner)
	if err != nil {
		return errors.Annotate(err, "CBI contents do not match").Err()
	}

	if info.GetChromeos().GetCbi().GetRawContents() != dutCBI.GetRawContents() {
		log.Debugf(ctx, "CBI contents do not match: CBI contents on the DUT do not match the CBI contents stored in UFS.\nCBI contents on DUT: %s\nCBI contents in UFS: %s", dutCBI.GetRawContents(), info.GetChromeos().GetCbi().GetRawContents())
		return nil
	}
	return errors.Reason("CBI contents do not match: CBI contents on DUT match the contents stored in UFS").Err()
}

// cbiIsPresent checks if CBI contents are found on the DUT.
func cbiIsPresent(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.NewRunner(info.GetDut().Name)
	cbiLocation, err := cbi.GetCBILocation(ctx, runner)
	if err != nil {
		return errors.Annotate(err, "CBI is present").Err()
	}
	if cbiLocation == nil {
		return errors.Reason("CBI is present: no CBI contents were found on the DUT, but encountered no error - please submit a bug").Err()
	}
	return nil
}

// backupCBI reads the CBI contents on the DUT and stores them in UFS.
func backupCBI(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.NewRunner(info.GetDut().Name)
	dutCBI, err := cbi.GetCBIContents(ctx, runner)
	if err != nil {
		return errors.Annotate(err, "backup CBI").Err()
	}

	info.GetChromeos().Cbi = dutCBI
	return err
}

// TODO(b/268499406): Replace all references of this function with cbiContentsAreValidExec
// cbiContentsContainValidMagic reads the CBI contents on the DUT  and returns an
// error if they do not contain valid CBI magic.
func cbiContentsContainValidMagic(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.NewRunner(info.GetDut().Name)
	dutCBI, err := cbi.GetCBIContents(ctx, runner)
	if err != nil {
		return errors.Annotate(err, "CBI contents contain valid magic").Err()
	}

	if !cbi.ContainsCBIMagic(dutCBI) {
		return errors.Reason("CBI contents contain valid magic: the CBI contents on the DUT do not contain valid magic:  %s", dutCBI.GetRawContents()).Err()
	}

	return nil
}

// cbiContentsAreValidExec reads the CBI contents on the DUT and returns an
// error if they do not contain valid CBI magic or are missing any of the
// required fields.
func cbiContentsAreValidExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.NewRunner(info.GetDut().Name)
	dutCBI, err := cbi.GetCBIContents(ctx, runner)
	if err != nil {
		return errors.Annotate(err, "CBI contents are valid").Err()
	}

	if !cbi.ContainsCBIMagic(dutCBI) {
		log.Debugf(ctx, "CBI contents are valid: CBI contents found on the DUT: %s", dutCBI.GetRawContents())
		return errors.Reason("CBI contents are valid: the CBI contents on the DUT do not contain valid magic").Err()
	}

	if err := cbi.VerifyRequiredFields(ctx, runner); err != nil {
		return errors.Annotate(err, "CBI contents are valid").Err()
	}

	return nil
}

func init() {
	execs.Register("cros_restore_cbi_contents_from_ufs", restoreCBIContentsFromUFS)
	execs.Register("cros_cbi_contents_do_not_match", cbiContentsDoNotMatch)
	execs.Register("cros_cbi_contents_match", cbiContentsMatch)
	execs.Register("cros_ufs_contains_cbi_contents", ufsContainsCBIContents)
	execs.Register("cros_ufs_does_not_contain_cbi_contents", ufsDoesNotContainCBIContents)
	execs.Register("cros_cbi_is_present", cbiIsPresent)
	execs.Register("cros_backup_cbi", backupCBI)
	execs.Register("cros_invalidate_cbi_cache", invalidateCBICache)
	execs.Register("cros_cbi_contents_contain_valid_magic", cbiContentsContainValidMagic)
	execs.Register("cros_cbi_contents_are_valid", cbiContentsAreValidExec)
}
