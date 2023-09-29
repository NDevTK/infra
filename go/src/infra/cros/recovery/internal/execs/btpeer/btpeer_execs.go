// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package btpeer

import (
	"context"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/btpeer"
	"infra/cros/recovery/internal/components/btpeer/chameleond"
	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// setStateBrokenExec sets state as BROKEN.
func setStateBrokenExec(ctx context.Context, info *execs.ExecInfo) error {
	if h, err := activeHost(info.GetActiveResource(), info.GetChromeos()); err != nil {
		return errors.Annotate(err, "set state broken").Err()
	} else {
		h.State = tlw.BluetoothPeer_BROKEN
	}
	return nil
}

// setStateWorkingExec sets state as WORKING.
func setStateWorkingExec(ctx context.Context, info *execs.ExecInfo) error {
	if h, err := activeHost(info.GetActiveResource(), info.GetChromeos()); err != nil {
		return errors.Annotate(err, "set state working").Err()
	} else {
		h.State = tlw.BluetoothPeer_WORKING
	}
	return nil
}

// getDetectedStatusesExec verifies communication with XMLRPC service running on bluetooth-peer and send one request to verify that service is responsive and initialized.
func getDetectedStatusesExec(ctx context.Context, info *execs.ExecInfo) error {
	h, err := activeHost(info.GetActiveResource(), info.GetChromeos())
	if err != nil {
		return errors.Annotate(err, "get detected statuses").Err()
	}
	res, err := Call(ctx, info.GetAccess(), h, "GetDetectedStatus")
	if err != nil {
		return errors.Annotate(err, "get detected statuses").Err()
	}
	count := len(res.GetArray().GetValues())
	if count == 0 {
		return errors.Reason("get detected statuses: list is empty").Err()
	}
	log.Debugf(ctx, "Detected statuses count: %v", count)
	return nil
}

// fetchInstalledChameleondBundleCommitExec retrieves the chameleond commit of
// the currently installed chameleond version from a log file on the btpeer and
// stores it in the btpeer scope state for later reference.
func fetchInstalledChameleondBundleCommitExec(ctx context.Context, info *execs.ExecInfo) error {
	btpeerScopeState, err := getBtpeerScopeState(ctx, info)
	if err != nil {
		return errors.Annotate(err, "failed to get btpeer scope state").Err()
	}
	sshRunner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	installedCommit, err := chameleond.FetchInstalledChameleondBundleCommit(ctx, sshRunner)
	if err != nil {
		return errors.Annotate(err, "failed to fetch installed chameleond bundle commit from btpeer").Err()
	}
	btpeerScopeState.Chameleond.InstalledCommit = installedCommit
	return nil
}

// fetchBtpeerChameleondReleaseConfigExec retrieves the production btpeer
// chameleond config from GCS and stores it in the scope state for later reference.
func fetchBtpeerChameleondReleaseConfigExec(ctx context.Context, info *execs.ExecInfo) error {
	btpeerScopeState, err := getBtpeerScopeState(ctx, info)
	if err != nil {
		return errors.Annotate(err, "failed to get btpeer scope state").Err()
	}
	sshRunner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	config, err := chameleond.FetchBtpeerChameleondReleaseConfig(ctx, sshRunner, info.GetAccess(), info.GetActiveResource())
	if err != nil {
		return errors.Annotate(err, "failed to fetch btpeer chameleond release config").Err()
	}
	configJSON, err := chameleond.MarshalBtpeerChameleondReleaseConfig(config)
	if err != nil {
		return errors.Annotate(err, "failed to marshal successfully fetched btpeer chameleond release config").Err()
	}
	log.Debugf(ctx, "Successfully retrieved btpeer chameleond release config:\n%s", configJSON)
	btpeerScopeState.Chameleond.ReleaseConfig = config
	return nil
}

// identifyExpectedChameleondReleaseBundleExec Identifies the expected
// chameleond release bundle based off of the chameleond config and DUT host.
// The config of the expected bundle is stored in the scope state for later
// reference.
//
// Note: For now this step ignores the DUT host and always selects the latest,
// non-next bundle. This can be adjusted in the config using the "cros_version"
// action arg to use a specific version (defaults to "999999999", which would
// always be higher than every release number to make sure the latest is chosen).
func identifyExpectedChameleondReleaseBundleExec(ctx context.Context, info *execs.ExecInfo) error {
	btpeerScopeState, err := getBtpeerScopeState(ctx, info)
	if err != nil {
		return errors.Annotate(err, "failed to get btpeer scope state").Err()
	}
	if btpeerScopeState.GetChameleond().GetReleaseConfig() == nil {
		return errors.Reason("invalid scope state: BluetoothPeerScopeState.Chameleond.ReleaseConfig is nil").Err()
	}
	// For now, we always want repair to select the highest non-next version, so
	// a very high version is used for selection to force this behavior.
	actionArgs := info.GetActionArgs(ctx)
	const crosVersionActionArgKey = "cros_version"
	crosVersion := actionArgs.AsString(ctx, crosVersionActionArgKey, "999999999")
	expectedBundleConfig, err := chameleond.SelectChameleondBundleByCrosReleaseVersion(btpeerScopeState.GetChameleond().GetReleaseConfig(), crosVersion)
	if err != nil {
		return errors.Annotate(err, "failed to select highest non-next chameleond bundle for btpeer").Err()
	}
	expectedBundleConfigJSON, err := protojson.Marshal(expectedBundleConfig)
	if err != nil {
		return errors.Annotate(err, "failed to marshall successfully identified expected bundle config").Err()
	}
	log.Debugf(ctx, "Successfully identified expected chameleond release bundle for cros_version %q: %s", crosVersion, expectedBundleConfigJSON)
	btpeerScopeState.Chameleond.ExpectedBundleConfig = expectedBundleConfig
	return nil
}

// assertBtpeerHasExpectedChameleondReleaseBundleInstalledExec checks if the
// installed chameleond commit matches the expected chameleond bundle commit and
// returns a non-nil error if it does not match.
func assertBtpeerHasExpectedChameleondReleaseBundleInstalledExec(ctx context.Context, info *execs.ExecInfo) error {
	btpeerScopeState, err := getBtpeerScopeState(ctx, info)
	if err != nil {
		return errors.Annotate(err, "failed to get btpeer scope state").Err()
	}
	installedCommit := btpeerScopeState.GetChameleond().GetInstalledCommit()
	if installedCommit == "" {
		return errors.Reason("invalid scope state: BluetoothPeerScopeState.Chameleond.InstalledCommit is empty").Err()
	}
	expectedBundleConfig := btpeerScopeState.GetChameleond().GetExpectedBundleConfig()
	if expectedBundleConfig == nil {
		return errors.Reason("invalid scope state: BluetoothPeerScopeState.Chameleond.ExpectedBundleConfig is nil").Err()
	}
	if !strings.EqualFold(installedCommit, expectedBundleConfig.GetChameleondCommit()) {
		return errors.Reason(
			"chameleond bundle installed on btpeer (commit %q) does not match expected bundle (commit %q)",
			installedCommit,
			expectedBundleConfig.GetChameleondCommit(),
		).Err()
	}
	log.Debugf(ctx, "Chameleond bundle installed on btpeer (commit %q) is the same as the expected bundle, assuming installation is as expected", installedCommit)
	return nil
}

// installExpectedChameleondReleaseBundleExec Installs/updates chameleond on the
// btpeer with the expected chameleond bundle.
//
// The expected bundle archive is downloaded from GCS to the btpeer through the
// cache, extracted, and installed via make.
func installExpectedChameleondReleaseBundleExec(ctx context.Context, info *execs.ExecInfo) error {
	btpeerScopeState, err := getBtpeerScopeState(ctx, info)
	if err != nil {
		return errors.Annotate(err, "failed to get btpeer scope state").Err()
	}
	expectedBundleConfig := btpeerScopeState.GetChameleond().GetExpectedBundleConfig()
	if expectedBundleConfig == nil {
		return errors.Reason("invalid scope state: BluetoothPeerScopeState.Chameleond.ExpectedBundleConfig is nil").Err()
	}
	expectedCommit := expectedBundleConfig.GetChameleondCommit()
	sshRunner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	// Prepare install dir.
	if err := chameleond.PrepareEmptyInstallDir(ctx, sshRunner); err != nil {
		return errors.Annotate(err, "failed to prepare empty install dir on btpeer").Err()
	}
	// Download bundle to btpeer.
	localBundleLocation, err := chameleond.DownloadChameleondBundle(ctx, sshRunner, info.GetAccess(), info.GetActiveResource(), expectedBundleConfig)
	if err != nil {
		return errors.Annotate(err, "failed to download expected chameleond bundle (commit %q) to btpeer", expectedCommit).Err()
	}
	// Install bundle.
	if err := chameleond.InstallChameleondBundle(ctx, sshRunner, localBundleLocation); err != nil {
		return errors.Annotate(err, "failed to install expected chameleond bundle (commit %q) on btpeer", expectedCommit).Err()
	}
	// Clean install dir.
	if err := chameleond.RemoveInstallDir(ctx, sshRunner); err != nil {
		return errors.Annotate(err, "failed to remove install dir on btpeer").Err()
	}
	// Validate installed bundle matches expected commit.
	installedCommit, err := chameleond.FetchInstalledChameleondBundleCommit(ctx, sshRunner)
	if err != nil {
		return errors.Annotate(err, "failed to fetch installed chameleond bundle commit from btpeer").Err()
	}
	btpeerScopeState.Chameleond.InstalledCommit = installedCommit
	if !strings.EqualFold(installedCommit, expectedCommit) {
		return errors.Annotate(err, "newly installed bundle (commit %q) does not match expected bundle (commit %q)", installedCommit, expectedCommit).Err()
	}
	log.Debugf(ctx, "Successfully installed expected chameleond bundle (commit %q) to btpeer", expectedCommit)
	return nil
}

// rebootExec reboots the device over ssh and waits for the device to become
// ssh-able again.
func rebootExec(ctx context.Context, info *execs.ExecInfo) error {
	sshRunner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	if err := ssh.Reboot(ctx, sshRunner, 10*time.Second, 10*time.Second, 3*time.Minute); err != nil {
		return errors.Annotate(err, "failed to reboot btpeer").Err()
	}
	return nil
}

// assertChameleondServiceIsRunningExec checks the status of the chameleond
// service on the device to see if it is running. Returns a non-nil error if the
// service is not running.
func assertChameleondServiceIsRunningExec(ctx context.Context, info *execs.ExecInfo) error {
	sshRunner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	if err := chameleond.AssertChameleondServiceIsRunning(ctx, sshRunner); err != nil {
		return errors.Annotate(err, "failed to assert that chameleond is running").Err()
	}
	return nil
}

// assertUptimeIsLessThanDurationExec checks the uptime of the device and
// fails if the uptime is not less than the duration in minutes provided in
// the "duration_min" action arg.
func assertUptimeIsLessThanDurationExec(ctx context.Context, info *execs.ExecInfo) error {
	// Parse duration arg.
	actionArgs := info.GetActionArgs(ctx)
	const durationMinArgKey = "duration_min"
	durationArg := actionArgs.AsDuration(ctx, durationMinArgKey, 24*60, time.Minute)

	// Get uptime from device.
	sshRunner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	uptime, err := cros.Uptime(ctx, sshRunner.Run)
	if err != nil {
		return errors.Annotate(err, "assert uptime is less than duration: failed to get uptime from device").Err()
	}

	// Evaluate assertion.
	if !(*uptime < durationArg) {
		return errors.Reason("assert uptime is less than duration: device uptime of %s is not less than %s", *uptime, durationArg).Err()
	}
	log.Debugf(ctx, "Device uptime of %s is less than %s", *uptime, durationArg)
	return nil
}

func init() {
	execs.Register("btpeer_state_broken", setStateBrokenExec)
	execs.Register("btpeer_state_working", setStateWorkingExec)
	execs.Register("btpeer_get_detected_statuses", getDetectedStatusesExec)
	execs.Register("btpeer_fetch_installed_chameleond_bundle_commit", fetchInstalledChameleondBundleCommitExec)
	execs.Register("btpeer_fetch_btpeer_chameleond_release_config", fetchBtpeerChameleondReleaseConfigExec)
	execs.Register("btpeer_identify_expected_chameleond_release_bundle", identifyExpectedChameleondReleaseBundleExec)
	execs.Register("btpeer_assert_btpeer_has_expected_chameleond_release_bundle_installed", assertBtpeerHasExpectedChameleondReleaseBundleInstalledExec)
	execs.Register("btpeer_install_expected_chameleond_release_bundle", installExpectedChameleondReleaseBundleExec)
	execs.Register("btpeer_reboot", rebootExec)
	execs.Register("btpeer_assert_chameleond_service_is_running", assertChameleondServiceIsRunningExec)
	execs.Register("btpeer_assert_uptime_is_less_than_duration", assertUptimeIsLessThanDurationExec)
}
