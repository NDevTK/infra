// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/cros/cellular"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger/metrics"
	"infra/cros/recovery/tlw"
)

func init() {
	execs.Register("carrier_not_in", carrierNotInExec)
	execs.Register("cros_audit_cellular_modem", auditCellularModemExec)
	execs.Register("cros_update_cellular_modem_labels", updateCellularModemLabelsExec)
	execs.Register("cros_update_cellular_sim_labels", updateCellularSIMLabelsExec)
	execs.Register("cros_has_only_one_sim_profile", hasOnlyOneSIMProfileExec)
	execs.Register("cros_audit_cellular_connection", auditCellularConnectionExec)
	execs.Register("cros_has_mmcli", hasModemManagerCLIExec)
	execs.Register("cros_has_modemmanager_job", hasModemManagerJobExec)
	execs.Register("cros_modemmanager_running", modemManagerRunningExec)
	execs.Register("cros_modem_state_not_in", modemStateNotInExec)
	execs.Register("cros_restart_modemmanager", restartModemManagerExec)
	execs.Register("set_cellular_modem_state", setCellularModemStateExec)
	execs.Register("has_cellular_info", hasCellularInfoExec)
}

// hasModemManagerCLIExec validates that mmcli is present on the DUT
func hasModemManagerCLIExec(ctx context.Context, info *execs.ExecInfo) error {
	if !cellular.HasModemManagerCLI(ctx, info.DefaultRunner(), info.GetExecTimeout()) {
		return errors.Reason("has modem manager cli: mmcli is not found on device").Err()
	}
	return nil
}

// hasModemManagerJobExec validates that modemmanager job is known by upstart and present on the DUT.
func hasModemManagerJobExec(ctx context.Context, info *execs.ExecInfo) error {
	if !cellular.HasModemManagerJob(ctx, info.DefaultRunner(), info.GetExecTimeout()) {
		return errors.Reason("has modem manager job: modemmanager is not found on device").Err()
	}
	return nil
}

// modemManagerRunningExec ensures modemmanager is running on the DUT and starts it if it's not already.
func modemManagerRunningExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.DefaultRunner()
	argsMap := info.GetActionArgs(ctx)
	waitTimeout := argsMap.AsDuration(ctx, "wait_timeout", 10, time.Second)
	startTimeout := argsMap.AsDuration(ctx, "start_timeout", 30, time.Second)
	if cellular.WaitForModemManager(ctx, runner, waitTimeout) == nil {
		return nil
	}

	if err := cellular.StartModemManager(ctx, runner, startTimeout); err != nil {
		return errors.Annotate(err, "start modemmanager").Err()
	}

	if err := cellular.WaitForModemManager(ctx, runner, waitTimeout); err != nil {
		return errors.Annotate(err, "wait for modemmanager to start").Err()
	}
	return nil
}

// restartModemManagerExec restarts modemmanagr on the DUT.
func restartModemManagerExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.DefaultRunner()
	argsMap := info.GetActionArgs(ctx)
	waitTimeout := argsMap.AsDuration(ctx, "wait_timeout", 10, time.Second)
	restartTimeout := argsMap.AsDuration(ctx, "restart_timeout", 30, time.Second)
	if err := cellular.RestartModemManager(ctx, runner, restartTimeout); err != nil {
		return errors.Annotate(err, "restart modemmanager").Err()
	}

	if err := cellular.WaitForModemManager(ctx, runner, waitTimeout); err != nil {
		return errors.Annotate(err, "wait for modemmanager to start").Err()
	}
	return nil
}

// hasCellularInfoExec validates that cellular data is populated in the dut info.
func hasCellularInfoExec(ctx context.Context, info *execs.ExecInfo) error {
	if c := info.GetChromeos().GetCellular(); c == nil {
		return errors.Reason("has cellular info: cellular data is not present in dut info").Err()
	}
	return nil
}

// setCellularModemStateExec sets the DUT's modem state to the requested value.
func setCellularModemStateExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("set cellular modem state: cellular data is not present in dut info").Err()
	}

	actionMap := info.GetActionArgs(ctx)
	state := strings.ToUpper(actionMap.AsString(ctx, "state", ""))
	if state == "" {
		return errors.Reason("set cellular modem state: state is not provided").Err()
	}
	s, ok := tlw.HardwareState_value[state]
	if !ok {
		return errors.Reason("set cellular modem state state: state %q is invalid", state).Err()
	}

	c.ModemState = tlw.HardwareState(s)
	return nil
}

// carrierNotInExec validates that the DUT cellular network carrier is not in a provided list.
func carrierNotInExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("carrier not in: cellular data is not present in dut info").Err()
	}

	if c.GetCarrier() == "" {
		return errors.Reason("carrier not in: DUT carrier label is empty").Err()
	}

	argsMap := info.GetActionArgs(ctx)
	carriers := argsMap.AsStringSlice(ctx, "carriers", []string{})
	for _, carrier := range carriers {
		if strings.EqualFold(carrier, c.GetCarrier()) {
			return errors.Reason("carrier not in: carrier %q is in the provided list", c.Carrier).Err()
		}
	}

	return nil
}

// modemStateNotInExec verifies that the modem state exported by ModemManager is not in the provided list.
func modemStateNotInExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("modem state not in: cellular data is not present in dut info").Err()
	}

	argsMap := info.GetActionArgs(ctx)
	timeout := argsMap.AsDuration(ctx, "wait_modem_timeout", 15, time.Second)
	modemInfo, err := cellular.WaitForModemInfo(ctx, info.DefaultRunner(), timeout)
	if err != nil {
		return errors.Reason("modem state not in: no modem exported by ModemManager").Err()
	}

	if modemInfo.GetState() == "" {
		return errors.Reason("modem state not in: modem state is empty").Err()
	}

	modemState := modemInfo.GetState()
	states := argsMap.AsStringSlice(ctx, "states", []string{})
	for _, state := range states {
		if strings.EqualFold(state, modemState) {
			return errors.Reason("modem state not in: modem state %q is in the provided list", modemState).Err()
		}
	}

	return nil
}

// reportCellularObservations reports relevant observations for monitoring cellular DUT states during an action.
func reportCellularObservations(ctx context.Context, info *execs.ExecInfo, timeout time.Duration) {
	runner := info.DefaultRunner()
	modemInfo, err := cellular.WaitForModemInfo(ctx, runner, timeout)
	if err != nil {
		// Execs that call this function should have a dependency on the modem being
		// in a good state, if some how the modem crashed in the middle of an exec
		// we should just report it and quit rather than erroring out.
		info.AddObservation(metrics.NewStringObservation("cellularModemHWState", "MISSING"))
		log.Errorf(ctx, err.Error())
		return
	}

	connectionState := "UNKNOWN"
	if modemInfo.GetState() != "" {
		connectionState = strings.ToUpper(modemInfo.GetState())
	}
	info.AddObservation(metrics.NewStringObservation("cellularModemHWState", "AVAILABLE"))
	info.AddObservation(metrics.NewStringObservation("cellularConnectionState", connectionState))

	// Signal strength may not always be available by the modem so only report if there's no error.
	if signalStrength, err := cellular.GetSignalStrength(ctx, runner, timeout); err == nil {
		for _, strength := range signalStrength {
			prefix := fmt.Sprintf("cellular%vSignal", strength.Technology)
			if strength.RSRP != nil {
				info.AddObservation(metrics.NewFloat64Observation(prefix+"RSRP", *strength.RSRP))
			}
			if strength.RSSI != nil {
				info.AddObservation(metrics.NewFloat64Observation(prefix+"RSSI", *strength.RSSI))
			}
			if strength.SNR != nil {
				info.AddObservation(metrics.NewFloat64Observation(prefix+"SNR", *strength.SNR))
			}
		}
	}
}

// auditCellularConnectionExec verifies that the device is able to connect to the provided cellular network.
func auditCellularConnectionExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.DefaultRunner()
	argsMap := info.GetActionArgs(ctx)
	waitTimeout := argsMap.AsDuration(ctx, "wait_connected_timeout", 120, time.Second)

	// Action requires at least 1 minute more than wait_connected_timeout to successfully complete the action.
	if waitTimeout+time.Minute < info.GetExecTimeout() {
		return errors.Reason("audit cellular connection: exec timeout must be >= wait_connected_timeout + 60s").Err()
	}

	// Report cellular state observations after connection attempt, don't perform this as a dedicated
	// action since we want the observations to be linked to the audit connection action.
	defer reportCellularObservations(ctx, info, 15*time.Second)

	if err := cellular.ConnectToDefaultService(ctx, runner, waitTimeout); err != nil {
		return errors.Annotate(err, "audit cellular connection").Err()
	}
	return nil
}

// updateCellularModemLabelsExec sets the cellular modem labels in swarming to match those available on the DUT.
func updateCellularModemLabelsExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("audit cellular modem labels: cellular data is not present in dut info").Err()
	}

	// Get labels directly from modem hardware.
	argsMap := info.GetActionArgs(ctx)
	timeout := argsMap.AsDuration(ctx, "wait_modem_timeout", 15, time.Second)
	modemInfo, err := cellular.WaitForModemInfo(ctx, info.DefaultRunner(), timeout)
	if err != nil {
		return errors.Reason("audit cellular modem labels: no modem exported by ModemManager").Err()
	}

	if modemInfo.GetImei() == "" {
		return errors.Reason("audit cellular modem labels: failed to get modem imei").Err()
	}

	// Get labels from cros_config.
	variant := cellular.GetModelVariant(ctx, info.DefaultRunner())
	if variant == "" {
		return errors.Reason("audit cellular modem labels: cellular variant not present on device").Err()
	}

	modemType := cellular.GetModemTypeFromVariant(variant)
	if modemType == tlw.Cellular_MODEM_TYPE_UNSUPPORTED && (c.ModemInfo.Type == tlw.Cellular_MODEM_TYPE_UNSPECIFIED || c.ModemInfo.Type == tlw.Cellular_MODEM_TYPE_UNSUPPORTED) {
		// If unknown modem type and no modem was previously specified then just log as its a new device.
		log.Errorf(ctx, "audit cellular modem labels: unknown modem type for variant: %q", variant)
	} else if modemType == tlw.Cellular_MODEM_TYPE_UNSUPPORTED {
		// If unknown modem type and modem was previously specified, then we should error out
		// without updating anything as something has gone wrong and the device can't be trusted.
		return errors.Reason("audit cellular modem labels: unknown modem type for variant: %q", variant).Err()
	}

	// Update properties at end once everything has been verified.
	c.ModemInfo.Type = modemType
	c.ModelVariant = variant
	c.ModemInfo.Imei = modemInfo.GetImei()
	return nil
}

// Ensures that the SIM labels only contain references to at most one profile.
func hasOnlyOneSIMProfileExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("audit cellular sim labels: cellular data is not present in dut info").Err()
	}

	// It is possible that some SIMs have more than one profile installed on them.
	// If this happens, then we should fail and require the SIM labels to be manually
	// added or wiped first. As updating these would require activating/deactivating
	// SIM profiles, which could have unintended consequences if there is an inactive
	// profile installed on the DUT.
	for _, s := range c.GetSimInfos() {
		if len(s.GetProfileInfos()) > 1 {
			return errors.Reason("audit cellular sim labels: expected <= 1 profiles for each SIM, got %d", len(s.ProfileInfos)).Err()
		}
	}
	return nil
}

// updateCellularSIMLabelsExec queries the available SIM cards on the DUT and updates their information.
func updateCellularSIMLabelsExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("audit cellular sim labels: cellular data is not present in dut info").Err()
	}

	// Fetch available SIMs on device, if we fail to fetch any required information for a SIM then
	// we should fail without updating as we would still want to know which SIM are not being properly
	// populated.
	simInfos, err := cellular.GetAllSIMInfo(ctx, info.DefaultRunner())
	if err != nil {
		return errors.Annotate(err, "audit cellular sim labels: failed to query sim info").Err()
	}

	simInfosBySlot := make(map[int32]*tlw.Cellular_SIMInfo)
	for _, si := range c.SimInfos {
		if _, ok := simInfosBySlot[si.SlotId]; ok {
			return errors.Reason("audit cellular sim labels: duplicate SIM slot ID found: %d", si.SlotId).Err()
		}
		simInfosBySlot[si.SlotId] = si
	}

	for _, newSI := range simInfos {
		if oldSI, ok := simInfosBySlot[newSI.SlotId]; ok {
			oldSI.Type = newSI.Type
			oldSI.Eid = newSI.Eid

			// Technically it's possible that old SimInfo has no profiles.
			var oldProfile *tlw.Cellular_SIMProfileInfo
			if len(oldSI.ProfileInfos) > 0 {
				oldProfile = oldSI.ProfileInfos[0]
			} else {
				oldProfile = &tlw.Cellular_SIMProfileInfo{}
				oldSI.ProfileInfos = append(oldSI.ProfileInfos, oldProfile)
			}

			// newSI always has exactly 1 profile since we just created it.
			newProfile := newSI.ProfileInfos[0]
			oldProfile.Iccid = newProfile.Iccid
			oldProfile.CarrierName = newProfile.CarrierName

			// OwnNumber may not be available from the DUT directly, and may instead
			// have been added directly to swarming via shivas, so don't overwrite it if it's missing.
			if newProfile.OwnNumber != "" {
				oldProfile.OwnNumber = newProfile.OwnNumber
			}
			delete(simInfosBySlot, newSI.SlotId)
		} else {
			c.SimInfos = append(c.SimInfos, newSI)
		}
	}

	// Check that we are not failing to detect any previously detected SIMs, if
	// we are, then fail the action after updating the SIMs that we did manage to detect.
	// We don't want to clear the missing SIMs since it may be helpful to know which SIMs
	// we are failing to find.
	if len(simInfosBySlot) != 0 {
		return errors.Reason("audit cellular sim labels: failed to find SIM info for %d slots", len(simInfosBySlot)).Err()
	}

	return nil
}

// auditCellularModem will validate cellular modem hardware state.
func auditCellularModemExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("audit cellular modem: cellular data is not present in dut info").Err()
	}
	expected := cellular.IsExpected(ctx, info.DefaultRunner())

	// if no cellular is expected then set total timeout to be much lower otherwise we will add
	// ~2 minutes to every repair even ones that don't require a modem.
	argsMap := info.GetActionArgs(ctx)
	timeout := argsMap.AsDuration(ctx, "wait_manager_when_not_expected", 120, time.Second)
	if !expected {
		timeout = argsMap.AsDuration(ctx, "wait_manager_when_expected", 15, time.Second)
	}

	modemInfo, err := cellular.WaitForModemInfo(ctx, info.DefaultRunner(), timeout)
	if err == nil {
		// found modem, try to get connection status.
		connectionState := "UNKNOWN"
		if modemInfo.Modem.Generic != nil && modemInfo.Modem.Generic.State != "" {
			connectionState = strings.ToUpper(modemInfo.Modem.Generic.State)
		}

		// only report connection state for devices where modem was found.
		info.AddObservation(metrics.NewStringObservation("cellularConnectionState", connectionState))
		c.ModemState = tlw.HardwareState_HARDWARE_NORMAL
		return nil
	}

	err = errors.Annotate(err, "audit cellular modem").Err()
	if execs.SSHErrorInternal.In(err) {
		c.ModemState = tlw.HardwareState_HARDWARE_UNSPECIFIED
		return err
	}

	if expected {
		// no modem detected but was expected by setup info
		// then we set needs_replacement as it is probably a hardware issue.
		c.ModemState = tlw.HardwareState_HARDWARE_NEED_REPLACEMENT
		return err
	}

	// not found and not expected, don't report an error, instead just log it
	log.Errorf(ctx, err.Error())
	c.ModemState = tlw.HardwareState_HARDWARE_NOT_DETECTED
	return nil
}
