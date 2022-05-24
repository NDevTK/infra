// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/dutstate"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
	ufsProto "infra/unifiedfleet/api/v1/models"
)

// hasDutNameActionExec verifies that DUT provides name.
func hasDutNameActionExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.RunArgs.DUT != nil && info.RunArgs.DUT.Name != "" {
		log.Debugf(ctx, "DUT name: %q", info.RunArgs.DUT.Name)
		return nil
	}
	return errors.Reason("dut name is empty").Err()
}

// hasDutBoardActionExec verifies that DUT provides board name.
func hasDutBoardActionExec(ctx context.Context, info *execs.ExecInfo) error {
	if d := info.RunArgs.DUT; d != nil && d.Board != "" {
		log.Debugf(ctx, "DUT board name: %q", d.Board)
		return nil
	}
	return errors.Reason("dut board name is empty").Err()
}

// hasDutModelActionExec verifies that DUT provides model name.
func hasDutModelActionExec(ctx context.Context, info *execs.ExecInfo) error {
	if d := info.RunArgs.DUT; d != nil && d.Model != "" {
		log.Debugf(ctx, "DUT model name: %q", d.Model)
		return nil
	}
	return errors.Reason("dut model name is empty").Err()
}

// dutServolessExec verifies that setup is servoless.
func dutServolessExec(ctx context.Context, info *execs.ExecInfo) error {
	if sh := info.RunArgs.DUT.ServoHost; sh == nil || (sh.Name == "" && sh.ContainerName == "") {
		log.Debugf(ctx, "DUT servoless confirmed!")
		return nil
	}
	return errors.Reason("dut is servoless").Err()
}

// hasDutDeviceSkuActionExec verifies that DUT has the device sku label.
func hasDutDeviceSkuActionExec(ctx context.Context, info *execs.ExecInfo) error {
	deviceSku := info.GetChromeos().GetDeviceSku()
	log.Debugf(ctx, "Device sku: %q.", deviceSku)
	if deviceSku == "" {
		return errors.Reason("dut device sku label is empty").Err()
	}
	return nil
}

const (
	// This token represents the string in config extra arguments that
	// conveys the expected string value(s) for a DUT attribute.
	stringValuesExtraArgToken = "string_values"
	// This token represents whether the success-status of an exec
	// should be inverted. For example, using this flag, we can
	// control whether the value of a DUT Model should, or should not
	// be present in the list of strings mentioned in the config.
	invertResultToken = "invert_result"
)

// dutCheckModelExec checks whether the model name for the current DUT
// matches any of the values specified in config. It returns an error
// based on the directive in config to invert the results.
func dutCheckModelExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	invertResultsFlag := argsMap.AsBool(ctx, invertResultToken, false)
	for _, m := range argsMap.AsStringSlice(ctx, stringValuesExtraArgToken, nil) {
		m = strings.TrimSpace(m)
		if strings.EqualFold(m, info.RunArgs.DUT.Model) {
			msg := fmt.Sprintf("DUT Model %s found in the list of models in config", info.RunArgs.DUT.Model)
			log.Debugf(ctx, "Dut Check Model Exec :%s.", msg)
			if invertResultsFlag {
				return errors.Reason("dut check model exec: %s", msg).Err()
			}
			return nil
		}
	}
	msg := "No matching model found"
	log.Debugf(ctx, "Dut Check Model Exec: %s", msg)
	if invertResultsFlag {
		return nil
	}
	return errors.Reason("dut check model exec: %s", msg).Err()
}

// dutCheckBoardExec checks whether the board name for the current DUT
// matches any of the values specified in config. It returns an error
// based on the directive in config to invert the results.
func dutCheckBoardExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	invertResultsFlag := argsMap.AsBool(ctx, invertResultToken, false)
	for _, m := range argsMap.AsStringSlice(ctx, stringValuesExtraArgToken, nil) {
		m = strings.TrimSpace(m)
		if strings.EqualFold(m, info.RunArgs.DUT.Board) {
			msg := fmt.Sprintf("DUT Board %s found in the list of boards in config", info.RunArgs.DUT.Model)
			log.Debugf(ctx, "Dut Check Board Exec :%s.", msg)
			if invertResultsFlag {
				return errors.Reason("dut check board exec: %s", msg).Err()
			}
			return nil
		}
	}
	msg := "No matching board found"
	log.Debugf(ctx, "Dut Check Board Exec: %s", msg)
	if invertResultsFlag {
		return nil
	}
	return errors.Reason("dut check board exec: %s", msg).Err()
}

// servoVerifySerialNumberExec verifies that the servo host attached
// to the DUT has a serial number configured.
func servoVerifySerialNumberExec(ctx context.Context, info *execs.ExecInfo) error {
	if d := info.RunArgs.DUT; d != nil && d.ServoHost != nil && d.ServoHost.SerialNumber != "" {
		log.Debugf(ctx, "Servo Verify Serial Number : %q", d.ServoHost.SerialNumber)
		return nil
	}
	return errors.Reason("servo verify serial number: serial number is not available").Err()
}

// servoHostPresentExec verifies that servo host present under DUT.
func servoHostPresentExec(ctx context.Context, info *execs.ExecInfo) error {
	if d := info.RunArgs.DUT; d == nil || d.ServoHost == nil {
		return errors.Reason("servo host present: data is not present").Err()
	}
	return nil
}

// dutInAudioBoxExec checks if DUT is in audio box.
func dutInAudioBoxExec(ctx context.Context, info *execs.ExecInfo) error {
	if d := info.RunArgs.DUT; d == nil || d.Audio == nil || !d.Audio.GetInBox() {
		return errors.Reason("is audio box: is not in audio-box").Err()
	}
	return nil
}

// hasBatteryExec checks if DUT is expected to have a battery.
func hasBatteryExec(ctx context.Context, info *execs.ExecInfo) error {
	if d := info.RunArgs.DUT; d == nil || d.Battery == nil {
		return errors.Reason("has battery: data is not present").Err()
	}
	return nil
}

// matchBatteryStateExec match statet of battery with expected.
//
// Please provide expected state by action args [state:your-state]
func matchBatteryStateExec(ctx context.Context, info *execs.ExecInfo) error {
	if d := info.RunArgs.DUT; d == nil || d.Battery == nil {
		return errors.Reason("match battery state: data is not present").Err()
	}
	actionMap := info.GetActionArgs(ctx)
	state := strings.ToUpper(actionMap.AsString(ctx, "state", ""))
	if state == "" {
		return errors.Reason("match battery state: state is not provided").Err()
	}
	s, ok := tlw.HardwareState_value[state]
	if !ok {
		return errors.Reason("match battery state: state %q is invalid", state).Err()
	}
	currentState := info.RunArgs.DUT.Battery.GetState()
	if s != int32(currentState) {
		return errors.Reason("match battery state: current state %s does not match expected state %q", currentState.String(), s).Err()
	}
	return nil
}

// hasDutHwidExec verifies that DUT has an HWID available.
func hasDutHwidExec(ctx context.Context, info *execs.ExecInfo) error {
	if d := info.RunArgs.DUT; d != nil && d.Hwid != "" {
		log.Debugf(ctx, "DUT Hwid: %q", d.Hwid)
		return nil
	}
	return errors.Reason("dut Hwid is empty").Err()
}

// hasDutSerialNumberExec verifies that DUT has a serial number
// available.
func hasDutSerialNumberExec(ctx context.Context, info *execs.ExecInfo) error {
	if d := info.RunArgs.DUT; d != nil && d.SerialNumber != "" {
		log.Debugf(ctx, "DUT Serial Number : %q", d.SerialNumber)
		return nil
	}
	return errors.Reason("dut serial number is empty").Err()
}

// regexNameMatchExec checks if name match to provided regex.
func regexNameMatchExec(ctx context.Context, info *execs.ExecInfo) error {
	actionMap := info.GetActionArgs(ctx)
	d := info.RunArgs.DUT
	if d == nil {
		return errors.Reason("regex name match: DUT not found").Err()
	}
	regex := actionMap.AsString(ctx, "regex", "")
	if regex == "" {
		return errors.Reason("regex name match: regex is empty").Err()
	}
	m, err := regexp.MatchString(regex, d.Name)
	if err != nil {
		return errors.Annotate(err, "regex name match").Err()
	}
	if !m {
		return errors.Reason("regex name match: not match").Err()
	}
	return nil
}

// setDutStateExec sets the state of the DUT to the value passed in
// the action arguments.
func setDutStateExec(ctx context.Context, info *execs.ExecInfo) error {
	args := info.GetActionArgs(ctx)
	newState := strings.ToLower(args.AsString(ctx, "state", ""))
	if newState == "" {
		return errors.Reason("set dut state: state is not provided").Err()
	}
	state := dutstate.State(newState)
	if dutstate.ConvertToUFSState(state) == ufsProto.State_STATE_UNSPECIFIED {
		return errors.Reason("set dut state: unsupported state %q", newState).Err()
	}
	log.Debugf(ctx, "Old DUT state: %s", info.RunArgs.DUT.State)
	info.RunArgs.DUT.State = state
	log.Infof(ctx, "New DUT state: %s", newState)
	return nil
}

// hasCr50PhaseExec verifies whether the Cr50 firmware is present on
// the DUT.
func hasCr50PhaseExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetChromeos().GetCr50Phase() == tlw.ChromeOS_CR50_PHASE_UNSPECIFIED {
		return errors.Reason("has Cr50 phase: Cr50 firmware phase could not be determined").Err()
	}
	return nil
}

// dutServoStateMandatesManualRepairExec checks whether the state of
// the servo on the DUT makes it imperative to set the state of the
// DUT to need manual repair.
func dutServoStateMandatesManualRepairExec(ctx context.Context, info *execs.ExecInfo) error {
	servoStates := info.GetActionArgs(ctx).AsStringSlice(ctx, "servo_states", nil)
	if servoStates == nil {
		log.Debugf(ctx, "DUT Servo State Mandates Manual Repair: no states mentioned in the action, don't have enough information to check whether servo state requires manual repair on the DUT.")
		return nil
	}
	sh := info.RunArgs.DUT.ServoHost
	// Assume this will always be non-nil because actions using this
	// exec will be conditioned in the config on availability of
	// ServoHost.
	for _, s := range servoStates {
		if state, ok := tlw.ServoHost_State_value[s]; ok && tlw.ServoHost_State(state) == sh.State {
			log.Debugf(ctx, "DUT Servo State Mandates Manual Repair")
			return nil
		}
	}
	return errors.Reason("dut servo state mandates manual repair: manual repair not required").Err()
}

func init() {
	execs.Register("dut_servo_host_present", servoHostPresentExec)
	execs.Register("dut_has_name", hasDutNameActionExec)
	execs.Register("dut_regex_name_match", regexNameMatchExec)
	execs.Register("dut_has_board_name", hasDutBoardActionExec)
	execs.Register("dut_has_model_name", hasDutModelActionExec)
	execs.Register("dut_has_device_sku", hasDutDeviceSkuActionExec)
	execs.Register("dut_check_model", dutCheckModelExec)
	execs.Register("dut_check_board", dutCheckBoardExec)
	execs.Register("dut_servoless", dutServolessExec)
	execs.Register("dut_is_in_audio_box", dutInAudioBoxExec)
	execs.Register("dut_servo_has_serial", servoVerifySerialNumberExec)
	execs.Register("dut_has_battery", hasBatteryExec)
	execs.Register("dut_match_battery_state", matchBatteryStateExec)
	execs.Register("dut_has_hwid", hasDutHwidExec)
	execs.Register("dut_has_serial_number", hasDutSerialNumberExec)
	execs.Register("dut_set_state", setDutStateExec)
	execs.Register("dut_has_cr50", hasCr50PhaseExec)
	execs.Register("dut_servo_state_required_manual_attention", dutServoStateMandatesManualRepairExec)
}
