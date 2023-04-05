// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	components_cros "infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/components/servo"
	components_topology "infra/cros/recovery/internal/components/servo/topology"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/cros"
	"infra/cros/recovery/internal/execs/cros/battery"
	"infra/cros/recovery/internal/execs/servo/topology"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/internal/retry"
	"infra/cros/recovery/logger/metrics"
	"infra/cros/recovery/tlw"
)

const (
	// Time between an usb disk plugged-in and detected in the system.
	usbDetectionDelay = 5

	// The prefix of the badblocks command for verifying USB
	// drives. The USB-drive path will be attached to it when
	// badblocks needs to be executed on a drive.
	badBlocksCommandPrefix = "badblocks -w -e 300 -b 4096 -t random %s"

	// This parameter represents the configuration for minimum number
	// of child servo devices to be verified-for.
	topologyMinChildArg = "min_child"

	// This parameter represents the configuration to control whether
	// the servo topology that read during servo topology verification
	// is persisted for use by other actions.
	persistTopologyArg = "persist_topology"

	// The default value that will be used to drive whether or not the
	// topology needs to be persisted. A value that is passed from the
	// configuration will over-ride this.
	persistTopologyDefaultValue = false

	// The default value that will be used for validating the number
	// of servo children in the servo topology. A value that is passed
	// from the configuration will over-ride this.
	topologyMinChildCountDefaultValue = 1

	// This command, when executed from servo host, checks whether the
	// servod process is responsive.
	servodHostCheckupCmd = "dut-control -p %d serialname"

	// This is the threshold voltage values between DUT and servo
	// Bus voltage on ppdut5. Value can be:
	//  - less than 500 - DUT is likely not connected
	//  - between 500 and 4000 - unexpected value
	//  - more than 4000 - DUT is likely connected
	maxPPDut5MVWhenNotConnected = 500
	minPPDut5MVWhenConnected    = 4000
	// File flag created in logs folder to request next servod start
	// use recovery mode by providing argument REC_MODE=1.
	servodUseRecoveryModeFlag = "servod_use_recovery_mode"
)

// servodInitActionExec init servod options and start servod on servo-host.
func servodInitActionExec(ctx context.Context, info *execs.ExecInfo) error {
	d := info.GetDut()
	if d == nil || d.Name == "" {
		return errors.Reason("init servod: DUT is not specified").Err()
	}
	chromeos := info.GetChromeos()
	if chromeos == nil {
		return errors.Reason("init servod: chromeos is not specified").Err()
	}
	sh := chromeos.GetServo()
	if sh == nil {
		return errors.Reason("init servod: servo-host or servo is not specified").Err()
	}
	actionArgs := info.GetActionArgs(ctx)
	useRecoveryMode := actionArgs.AsBool(ctx, "recovery_mode", false)
	if !useRecoveryMode {
		// The request to use recovery mode can be specified by presence of a specific file.
		logRoot := info.GetLogRoot()
		flagPath := filepath.Join(logRoot, servodUseRecoveryModeFlag)
		// If the call fail we think that file is not exist.
		// The call cannot fail as part of permission issue as file is created under the same user.
		if _, err := os.Stat(flagPath); err == nil {
			useRecoveryMode = true
		}
	}

	o := &tlw.ServodOptions{
		RecoveryMode:  useRecoveryMode,
		DutBoard:      chromeos.GetBoard(),
		DutModel:      chromeos.GetModel(),
		ServodPort:    int32(sh.GetServodPort()),
		ServoSerial:   sh.GetSerialNumber(),
		ServoDual:     false,
		UseCr50Config: false,
	}
	if vs, ok := d.ExtraAttributes[tlw.ExtraAttributeServoSetup]; ok {
		for _, v := range vs {
			if v == tlw.ExtraAttributeServoSetupDual {
				o.ServoDual = true
				break
			}
		}
	}
	if pools, ok := d.ExtraAttributes[tlw.ExtraAttributePools]; ok {
		for _, p := range pools {
			if strings.Contains(p, "faft-cr50") {
				o.UseCr50Config = true
				break
			}
		}
	}
	info.NewLogger().Debugf("Servod options: %s", o)
	am := info.GetActionArgs(ctx)
	req := &tlw.InitServodRequest{
		Resource: d.Name,
		Options:  o,
		NoServod: am.AsBool(ctx, "no_servod", false),
	}
	if err := info.GetAccess().InitServod(ctx, req); err != nil {
		return errors.Annotate(err, "init servod").Err()
	}
	return nil
}

func servodStopActionExec(ctx context.Context, info *execs.ExecInfo) error {
	if err := info.GetAccess().StopServod(ctx, info.GetDut().Name); err != nil {
		return errors.Annotate(err, "stop servod").Err()
	}
	return nil
}

func servodCreateFlagToUseRecoveryModeExec(ctx context.Context, info *execs.ExecInfo) error {
	logRoot := info.GetLogRoot()
	if logRoot == "" {
		return errors.Reason("servod create flag to use recovery-mode: log root is not specified").Err()
	}
	flagPath := filepath.Join(logRoot, servodUseRecoveryModeFlag)
	err := exec.CommandContext(ctx, "touch", flagPath).Run()
	return errors.Annotate(err, "servod create flag to use recovery-mode").Err()
}

func runCheckOnHost(ctx context.Context, run execs.Runner, usbPath string, timeout time.Duration) (tlw.HardwareState, error) {
	command := fmt.Sprintf(badBlocksCommandPrefix, usbPath)
	log.Debugf(ctx, "Run Check On Host: Executing %q", command)
	// The execution timeout for this audit job is configured at the
	// level of the action. So the execution of this command will be
	// bound by that.
	out, err := run(ctx, timeout, command)
	switch {
	case err == nil:
		// TODO(vkjoshi@): recheck if this is required, or does stderr need to be examined.
		if len(out) > 0 {
			return tlw.HardwareState_HARDWARE_NEED_REPLACEMENT, nil
		}
		return tlw.HardwareState_HARDWARE_NORMAL, nil
	case execs.SSHErrorLinuxTimeout.In(err): // 124 timeout
		fallthrough
	case execs.SSHErrorCLINotFound.In(err): // 127 badblocks
		return tlw.HardwareState_HARDWARE_UNSPECIFIED, errors.Annotate(err, "run check on host: could not successfully complete check").Err()
	default:
		return tlw.HardwareState_HARDWARE_NEED_REPLACEMENT, nil
	}
}

func servoAuditUSBKeyExec(ctx context.Context, info *execs.ExecInfo) error {
	sh := info.GetChromeos().GetServo()
	if sh.GetName() == "" {
		return errors.Reason("servo audit usb key: servo is not present as part of dut info").Err()
	}
	dutUsb := ""
	dutRunner := info.NewRunner(info.GetDut().Name)
	if components_cros.IsSSHable(ctx, dutRunner, components_cros.DefaultSSHTimeout) == nil {
		log.Debugf(ctx, "Servo Audit USB-Key Exec: %q is reachable through SSH", info.GetDut().Name)
		var err error = nil
		dutUsb, err = GetUSBDrivePathOnDut(ctx, dutRunner, info.NewServod())
		if err != nil {
			log.Debugf(ctx, "Servo Audit USB-Key Exec: could not determine USB-drive path on DUT: %q, error: %q. This is not critical. We will continue the audit by setting the path to empty string.", info.GetDut().Name, err)
		}
	} else {
		log.Debugf(ctx, "Servo Audit USB-Key Exec: continue audit from servo-host because DUT %q is not reachable through SSH", info.GetDut().Name)
	}
	if dutUsb != "" {
		// DUT is reachable, and we found a USB drive on it.
		state, err := runCheckOnHost(ctx, dutRunner, dutUsb, 2*time.Hour)
		if err != nil {
			return errors.Reason("servo audit usb key exec: could not check DUT usb path %q", dutUsb).Err()
		}
		sh.UsbkeyState = state
	} else {
		// Either the DUT is not reachable, or it does not have a USB
		// drive attached to it.

		// This statement obtains the path of usb drive on
		// servo-host. It also switches the USB drive on servo
		// multiplexer to servo-host.
		servoUsbPath, err := servodGetString(ctx, info.NewServod(), "image_usbkey_dev")
		if err != nil {
			// A dependency has already checked that the Servo USB is
			// available. But here we again check that no errors
			// occurred while determining USB path, in case something
			// changed between execution of dependency, and this
			// operation.
			sh.UsbkeyState = tlw.HardwareState_HARDWARE_NOT_DETECTED
			return errors.Annotate(err, "servo audit usb key exec: could not obtain usb path on servo: %q", err).Err()
		}
		servoUsbPath = strings.TrimSpace(servoUsbPath)
		if servoUsbPath == "" {
			sh.UsbkeyState = tlw.HardwareState_HARDWARE_NOT_DETECTED
			log.Debugf(ctx, "Servo Audit USB-Key Exec: cannot continue audit because the path to USB-Drive is empty")
			return errors.Reason("servo audit usb key exec: the path to usb drive is empty").Err()
		}
		state, err := runCheckOnHost(ctx, info.NewRunner(sh.GetName()), servoUsbPath, 2*time.Hour)
		if err != nil {
			log.Debugf(ctx, "Servo Audit USB-Key Exec: error %q during audit of USB-Drive", err)
			return errors.Annotate(err, "servo audit usb key: could not check usb path %q on servo-host %q", servoUsbPath, info.GetChromeos().GetServo().GetName()).Err()
		}
		sh.UsbkeyState = state
	}
	return nil
}

// Verify that the root servo is enumerated/present on the host.
// To force re-read topology please specify `update_topology:true`.
func isRootServoPresentExec(ctx context.Context, info *execs.ExecInfo) error {
	sh := info.GetChromeos().GetServo()
	if sh.GetName() == "" {
		return errors.Reason("is root servo present: servo is not present as part of dut info").Err()
	}
	runner := info.NewRunner(sh.GetName())
	am := info.GetActionArgs(ctx)
	if am.AsBool(ctx, "update_topology", false) {
		servoTopology, err := topology.RetrieveServoTopology(ctx, runner, sh.GetSerialNumber())
		if err != nil {
			return errors.Annotate(err, "is root servo present exec").Err()
		}
		sh.ServoTopology = servoTopology
	}
	rootServo, err := topology.GetRootServo(ctx, runner, sh.GetSerialNumber())
	if err != nil {
		return errors.Annotate(err, "is root servo present exec").Err()
	}
	if !components_topology.IsItemGood(ctx, rootServo) {
		log.Infof(ctx, "Is Servo Root Present Exec: no good root servo found")
		return errors.Reason("is servo root present exec: no good root servo found").Err()
	}
	log.Infof(ctx, "is servo root present exec: success")
	return nil
}

// Verify that the root servo is enumerated/present on the host.
func servoTopologyUpdateExec(ctx context.Context, info *execs.ExecInfo) error {
	sh := info.GetChromeos().GetServo()
	if sh.GetName() == "" {
		return errors.Reason("servo topology update: servo is not present as part of dut info").Err()
	}
	runner := info.NewRunner(sh.GetName())
	servoTopology, err := topology.RetrieveServoTopology(ctx, runner, sh.GetSerialNumber())
	if err != nil {
		return errors.Annotate(err, "servo topology update exec").Err()
	}
	if servoTopology.Root == nil {
		return errors.Reason("servo topology update exec: root servo not found").Err()
	}
	argsMap := info.GetActionArgs(ctx)
	minChildCount := topologyMinChildCountDefaultValue
	persistTopology := persistTopologyDefaultValue
	for k, v := range argsMap {
		log.Debugf(ctx, "Servo Topology Update Exec: k:%q, v:%q", k, v)
		if v != "" {
			// a non-empty value string implies that the corresponding
			// action arg was parsed correctly.
			switch k {
			case topologyMinChildArg:
				// If the configuration contains any min_child parameters,
				// it will be used for validation here. If no such
				// argument is present, we will not conduct any validation
				// of number of child servo based min_child.
				minChildCount, err = strconv.Atoi(v)
				if err != nil {
					return errors.Reason("servo topology update exec: malformed min child config in action arg %q:%q", k, v).Err()
				}
			case persistTopologyArg:
				persistTopology, err = strconv.ParseBool(v)
				if err != nil {
					return errors.Reason("servo topology update exec: malformed update servo config in action arg %q:%q", k, v).Err()
				}
			}
		}
	}
	if len(servoTopology.Children) < minChildCount {
		return errors.Reason("servo topology update exec: expected a min of %d children, found %d", minChildCount, len(servoTopology.Children)).Err()
	}
	if persistTopology {
		// This verified topology will be used in all subsequent
		// action that need the servo topology. This will avoid time
		// with re-fetching the topology.
		sh.ServoTopology = servoTopology
	}
	return nil
}

// Verify that servod is responsive
func servoServodEchoHostExec(ctx context.Context, info *execs.ExecInfo) error {
	sh := info.GetChromeos().GetServo()
	if sh.GetName() == "" {
		return errors.Reason("servo servod echo host: servo is not present as part of dut info").Err()
	}
	runner := info.NewRunner(sh.GetName())
	v, err := runner(ctx, time.Minute, fmt.Sprintf(servodHostCheckupCmd, info.NewServod().Port()))
	if err != nil {
		return errors.Annotate(err, "servo servod echo host exec: servod is not responsive for dut-control commands").Err()
	}
	log.Debugf(ctx, "Servo Servod Echo Host Exec: Servod is responsive: %q", v)
	return nil
}

// Verify that the servo firmware is up-to-date.
func servoFirmwareNeedsUpdateExec(ctx context.Context, info *execs.ExecInfo) error {
	sh := info.GetChromeos().GetServo()
	if sh.GetName() == "" {
		return errors.Reason("servo firmware needs update: servo is not present as part of dut info").Err()
	}
	runner := info.NewRunner(sh.GetName())
	// The servo topology check should have already been done in an
	// action. The topology determined at that time would have been
	// saved in this data structure if the 'updateServo' argument was
	// passed for that action. We will make use of any such persisting
	// topology instead of re-computing it. This is avoid unnecessary
	// expenditure of time in obtaining the topology here.
	devices := topology.Devices(sh.GetServoTopology(), "")
	err := components_topology.VerifyServoTopologyItems(ctx, devices)
	if err != nil {
		// This situation can arise if the servo topology has been
		// verified in an earlier action, but the topology was not
		// persisted because the updateServo parameter was not set in
		// that action, or for some reason the stored topology is
		// corrupted. In this case we do not have any choice but to
		// re-compute the topology.
		devices, err = topology.ListOfDevices(ctx, runner, sh.GetSerialNumber())
		if err != nil {
			errors.Annotate(err, "servo firmware needs update exec").Err()
		}
		log.Debugf(ctx, "Servo Firmware Needs Update Exec: topology re-computed because pre-existing servo topology not found, or had errors.")
	}
	for _, d := range devices {
		if components_topology.IsItemGood(ctx, d) {
			log.Debugf(ctx, "Servo Firmware Needs Update Exec: device type (d.Type) :%q.", d.Type)
			if needsUpdate(ctx, runner, d, sh.GetFirmwareChannel()) {
				log.Debugf(ctx, "Servo Firmware Needs Update Exec: needs update is true")
				return errors.Reason("servo firmware needs update exec: servo needs update").Err()
			}
		}
	}
	return nil
}

// servoSetExec sets the command of the servo a specific value using servod.
// It reads the command and its value from the actionArgs argument.
//
// the actionArgs should be in the format of ["command:....", "string_value:...."]
func servoSetExec(ctx context.Context, info *execs.ExecInfo) error {
	m := info.GetActionArgs(ctx)
	command, existed := m["command"]
	if !existed {
		return errors.Reason("servo set state: command not found in the argument").Err()
	}
	stringValue, existed := m["string_value"]
	if !existed {
		return errors.Reason("servo set state: string value not found in the argument").Err()
	}
	command = strings.TrimSpace(command)
	stringValue = strings.TrimSpace(stringValue)
	if err := info.NewServod().Set(ctx, command, stringValue); err != nil {
		return errors.Annotate(err, "servo set state").Err()
	}
	return nil
}

// Verify that the DUT is connected to Servo using the 'ppdut5_mv'
// servod control.
func servoLowPPDut5Exec(ctx context.Context, info *execs.ExecInfo) error {
	s := info.NewServod()
	if err := s.Has(ctx, servodPPDut5Cmd); err != nil {
		return errors.Annotate(err, "servo low ppdut5 exec").Err()
	}
	voltageValue, err := servodGetDouble(ctx, s, servodPPDut5Cmd)
	if err != nil {
		return errors.Annotate(err, "servo low ppdut5 exec").Err()
	}
	if voltageValue < maxPPDut5MVWhenNotConnected {
		return errors.Reason("servo low ppdut5 exec: the ppdut5_mv value %v is lower than the threshold %d", voltageValue, maxPPDut5MVWhenNotConnected).Err()
	}
	// TODO(vkjoshi@): add metrics to collect the value of the
	// servod control ppdut5_mv when it is below a certain threshold.
	// (ref:http://cs/chromeos_public/src/third_party/labpack/files/server/hosts/servo_repair.py?l=640).
	return nil
}

// Verify that the control has min double value by servod control.
func servoControlMinDoubleValueExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	control := argsMap.AsString(ctx, "control", "")
	minValue := argsMap.AsFloat64(ctx, "min_value", 0)
	if control == "" {
		return errors.Reason("servo control low double value: control is not provided").Err()
	}
	s := info.NewServod()
	if err := s.Has(ctx, control); err != nil {
		return errors.Annotate(err, "servo control low double value").Err()
	}
	receivedValue, err := servodGetDouble(ctx, s, control)
	if err != nil {
		return errors.Annotate(err, "servo control low double value").Err()
	}
	info.AddObservation(metrics.NewStringObservation("control", control))
	info.AddObservation(metrics.NewFloat64Observation("receivedValue", receivedValue))
	if receivedValue < minValue {
		return errors.Reason("servo control low double value: the value %v is lower than the threshold %v", receivedValue, minValue).Err()
	}
	log.Debugf(ctx, "Servo %q: the value %v is >= than the threshold %d", control, receivedValue, minValue)
	return nil
}

// servoCheckServodControlExec verifies that servod supports the
// control mentioned in action args. Additionally, if actionArgs
// includes the expected value, this function will verify that the
// value returned by servod for this control matches the expected
// value.
func servoCheckServodControlExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	command := argsMap.AsString(ctx, "command", "")
	if len(command) == 0 {
		return errors.Reason("servo check servod control exec: command not provided").Err()
	}
	info.AddObservation(metrics.NewStringObservation("command", command))
	var compare func(ctx context.Context) error
	// TODO(vkjoshi@): revisit the logic of implementations of the
	// function 'compare', e.g., will it make sense to use a helper
	// function for this?
	const expectedStringKey = "expected_string_value"
	const expectedIntKey = "expected_int_value"
	const expectedFloatKey = "expected_float_value"
	const expectedBoolKey = "expected_bool_value"
	if argsMap.Has(expectedStringKey) {
		expectedValue := argsMap.AsString(ctx, expectedStringKey, "")
		controlValue, err := servodGetString(ctx, info.NewServod(), command)
		if err != nil {
			return errors.Annotate(err, "servo check servod control exec").Err()
		}
		info.AddObservation(metrics.NewStringObservation("received", controlValue))
		compare = func(ctx context.Context) error {
			log.Infof(ctx, "Compare (String), expected value %q, actual value %q", expectedValue, controlValue)
			if controlValue != expectedValue {
				return errors.Reason("compare (string): expected value %q, actual value %q do not match.", expectedValue, controlValue).Err()
			}
			return nil
		}
	} else if argsMap.Has(expectedIntKey) {
		expectedValue := argsMap.AsInt(ctx, expectedIntKey, 0)
		controlValue, err := servodGetInt(ctx, info.NewServod(), command)
		if err != nil {
			return errors.Annotate(err, "servo check servod control exec").Err()
		}
		info.AddObservation(metrics.NewInt64Observation("received", int64(controlValue)))
		compare = func(ctx context.Context) error {
			log.Debugf(ctx, "Compare (Int), expected value %s, actual value %d", expectedValue, controlValue)
			if controlValue != int32(expectedValue) {
				return errors.Reason("compare: expected value %d, actual value %d do not match", int32(expectedValue), controlValue).Err()
			}
			return nil
		}
	} else if argsMap.Has(expectedFloatKey) {
		expectedValue := argsMap.AsFloat64(ctx, expectedFloatKey, 0)
		controlValue, err := servodGetDouble(ctx, info.NewServod(), command)
		if err != nil {
			return errors.Annotate(err, "servo check servod control exec").Err()
		}
		info.AddObservation(metrics.NewFloat64Observation("received", controlValue))
		compare = func(ctx context.Context) error {
			log.Debugf(ctx, "Compare (Double), expected value %s, actual value %f", expectedValue, controlValue)
			if controlValue != expectedValue {
				return errors.Reason("compare: expected value %f, actual value %f do not match", expectedValue, controlValue).Err()
			}
			return nil
		}
	} else if argsMap.Has(expectedBoolKey) {
		expectedValue := argsMap.AsBool(ctx, expectedBoolKey, false)
		controlValue, err := servodGetBool(ctx, info.NewServod(), command)
		if err != nil {
			return errors.Annotate(err, "servo check servod control exec").Err()
		}
		info.AddObservation(metrics.NewStringObservation("received", fmt.Sprintf("%v", controlValue)))
		compare = func(ctx context.Context) error {
			log.Debugf(ctx, "Compare (Bool), expected value %s, actual value %t", expectedValue, controlValue)
			if controlValue != expectedValue {
				return errors.Reason("compare: expected value %t, actual value %t do not match", expectedValue, controlValue).Err()
			}
			return nil
		}
	}
	if compare == nil {
		log.Infof(ctx, "Servo Check Servod Control Exec: expected value type not specified in config, or did not match any known types.")
		res, err := info.NewServod().Get(ctx, command)
		if err != nil {
			return errors.Annotate(err, "servo check servod control exec").Err()
		}
		// The value can contain different value types.
		// Ex.: "double:xxxx.xx"
		resRawString := strings.TrimSpace(res.String())
		info.AddObservation(metrics.NewStringObservation("received", resRawString))
		log.Infof(ctx, "Servo Check Servod Control Exec: for command %q, received %q.", command, resRawString)
	} else if err := compare(ctx); err != nil {
		return errors.Annotate(err, "servo check servod control exec").Err()
	}
	return nil
}

const (
	// removeFileCmd is the linux file removal command that used to remove files in the filesToRemoveSlice.
	removeFileCmd = `rm %s`
)

var filesToRemoveSlice = []string{
	"/var/lib/metrics/uma-events",
	"/var/spool/crash/*",
	"/var/log/chrome/*",
	"/var/log/ui/*",
	"/home/chronos/BrowserMetrics/*",
}

// servoLabstationDiskCleanUpExec remove files that are in the filesToRemoveSlice.
func servoLabstationDiskCleanUpExec(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	// Remove all files in the filesToRemoveSlice during the labstation disk clean up process.
	for _, filePath := range filesToRemoveSlice {
		if _, err := r(ctx, time.Minute, fmt.Sprintf(removeFileCmd, filePath)); err != nil {
			log.Debugf(ctx, "servo labstation disk clean up: %s", err.Error())
		}
		log.Infof(ctx, "labstation file removed: %s", filePath)
	}
	return nil
}

const (
	// removeOldServodLogsCmd is the command to remove any servod files that is older than the maximum days specified by d.
	removeOldServodLogsCmd = `/usr/bin/find /var/log/servod_* -mtime +%d -print -delete`
)

// servoServodOldLogsCleanupExec removes the old servod log files that existed more than keepLogsMaxDays days.
//
// @params: actionArgs should be in the format of: ["max_days:5"]
func servoServodOldLogsCleanupExec(ctx context.Context, info *execs.ExecInfo) error {
	daysMap := info.GetActionArgs(ctx)
	keepLogsMaxDaysString, existed := daysMap["max_days"]
	if !existed {
		return errors.Reason("servod old logs: missing max days information in the argument").Err()
	}
	keepLogsMaxDaysString = strings.TrimSpace(keepLogsMaxDaysString)
	if keepLogsMaxDaysString == "" {
		return errors.Reason("servod old logs: max days information is empty").Err()
	}
	keepLogsMaxDays, err := strconv.ParseInt(keepLogsMaxDaysString, 10, 64)
	if err != nil {
		return errors.Annotate(err, "servod old logs").Err()
	}
	log.Infof(ctx, "The max number of days for keeping old servod logs is: %v", keepLogsMaxDays)
	r := info.DefaultRunner()
	// remove old servod logs.
	if _, err := r(ctx, time.Minute, fmt.Sprintf(removeOldServodLogsCmd, keepLogsMaxDays)); err != nil {
		log.Debugf(ctx, "servo servod old logs clean up: %s", err.Error())
	}
	return nil
}

// servoValidateBatteryChargingExec uses servod controls to check
// whether or not the battery on a DUT is capable of getting
// charged. It marks the DUT for replacement if its battery cannot be
// charged.
func servoValidateBatteryChargingExec(ctx context.Context, info *execs.ExecInfo) error {
	// This is the number of times we will try to read the value of battery controls from servod.
	const servodBatteryReadRetryLimit = 3
	// This is the servod control to determine battery's last full charge.
	const batteryFullChargeServodControl = "battery_full_charge_mah"
	// This is the servod control to determine the bettery's full capacity by design.
	const batteryDesignFullCapacityServodControl = "battery_full_design_mah"
	var lastFullCharge, batteryCapacity int32
	var getLastFullCharge = func() error {
		var err error
		lastFullCharge, err = servodGetInt(ctx, info.NewServod(), batteryFullChargeServodControl)
		return err
	}
	if info.GetChromeos().GetBattery() == nil {
		return errors.Reason("servo validate battery charging: data is not present in dut info").Err()
	}
	if err := retry.LimitCount(ctx, servodBatteryReadRetryLimit, -1, getLastFullCharge, "get last full charge"); err != nil {
		log.Debugf(ctx, "Servo Validate Battery Charging: could not read last full charge despite trying %d times", servodBatteryReadRetryLimit)
		return errors.Annotate(err, "servo validate battery charging").Err()
	}
	log.Debugf(ctx, "Servo Validate Battery Charging: last full charge is %d", lastFullCharge)
	var getBatteryCapacity = func() error {
		var err error
		batteryCapacity, err = servodGetInt(ctx, info.NewServod(), batteryDesignFullCapacityServodControl)
		return err
	}
	if err := retry.LimitCount(ctx, servodBatteryReadRetryLimit, -1, getBatteryCapacity, "get battery capacity"); err != nil {
		log.Debugf(ctx, "Servo Validate Battery Charging: could not read battery capacity despite trying %d times", servodBatteryReadRetryLimit)
		return errors.Annotate(err, "servo validate battery charging").Err()
	}
	log.Debugf(ctx, "Servo Validate Battery Charging: battery capacity is %d", batteryCapacity)
	hardwareState := battery.DetermineHardwareStatus(ctx, float64(lastFullCharge), float64(batteryCapacity))
	log.Infof(ctx, "Battery hardware state: %s", hardwareState)
	if hardwareState == tlw.HardwareState_HARDWARE_UNSPECIFIED {
		return errors.Reason("servo validate battery charging: dut battery did not detected or state cannot extracted").Err()
	}
	if hardwareState == tlw.HardwareState_HARDWARE_NEED_REPLACEMENT {
		log.Infof(ctx, "Detected issue with storage on the DUT.")
		info.GetChromeos().GetBattery().State = tlw.HardwareState_HARDWARE_NEED_REPLACEMENT
	}
	return nil
}

// initDutForServoExec initializes the DUT and sets all servo signals
// to default values.
func initDutForServoExec(ctx context.Context, info *execs.ExecInfo) error {
	verbose := true
	s := info.NewServod()
	if _, err := s.Call(ctx, "hwinit", components.ServodDefaultTimeout, verbose); err != nil {
		return errors.Annotate(err, "init dut for servo exec").Err()
	}
	usbMuxControl := "usb_mux_oe1"
	if err := s.Has(ctx, usbMuxControl); err == nil {
		if err2 := s.Set(ctx, usbMuxControl, "on"); err2 != nil {
			return errors.Annotate(err, "init dut for servo exec").Err()
		}
		if err := s.Set(ctx, "image_usbkey_pwr", "off"); err != nil {
			return errors.Annotate(err, "init dut for servo exec").Err()
		}
	} else {
		log.Debugf(ctx, "Init Dut For Servo Exec: servod control %q is not available.", usbMuxControl)
	}
	return nil
}

// servoUpdateServoFirmwareExec updates all the servo devices' firmware based on the condition specified by the actionArgs
//
// @params try_attempt_count:  Count of attempts to update servo. For force option the count attempts is always 1 (one).
// @params try_force_update_after_fail:   Try force force option if fail to update in normal mode.
// @params force_update:       Run updater with force option. Override try_force_update_after_fail option.
// @params ignore_version:     Skip check the version on the device.
//
// @params: actionArgs should be in the format of:
// Ex: ["try_attempt_count:x", "try_force_update_after_fail:true/false",
//
//	"force_update:true/false", "ignore_version:true/false",
//	"servo_board:servo_micro"]
func servoUpdateServoFirmwareExec(ctx context.Context, info *execs.ExecInfo) (err error) {
	sh := info.GetChromeos().GetServo()
	if sh.GetName() == "" {
		return errors.Reason("servo update servo firmware: servo is not present as part of dut info").Err()
	}
	fwUpdateMap := info.GetActionArgs(ctx)
	// If the passed in "try_attempt_count" is either 0 or cannot be parsed successfully,
	// then, we default the count to be 1 to at least try to update it once.
	tryAttemptCount := fwUpdateMap.AsInt(ctx, "try_attempt_count", 1)
	tryForceUpdateAfterFail := fwUpdateMap.AsBool(ctx, "try_force_update_after_fail", false)
	forceUpdate := fwUpdateMap.AsBool(ctx, "force_update", false)
	ignoreVersion := fwUpdateMap.AsBool(ctx, "ignore_version", false)
	filteredServoBoard := fwUpdateMap.AsString(ctx, "servo_board", "")
	if filteredServoBoard != "" {
		log.Debugf(ctx, "Servo update servo firmware: Only updating board: %q's firmware", filteredServoBoard)
	}
	startTime := time.Now()
	if info.GetMetrics() != nil {
		// Record fw flash time to karte.
		action := &metrics.Action{
			// TODO(b/248635230): When karte' Search API is capable of taking in asset tag,
			// change the query to use asset tag instead of using hostname.
			Hostname:   info.GetDut().Name,
			ActionKind: metrics.ServoFwUpdateKind,
			StartTime:  startTime,
			Status:     metrics.ActionStatusSuccess,
		}
		if mErr := info.GetMetrics().Create(ctx, action); mErr != nil {
			log.Debugf(ctx, "Servo update servo firmware: cannot create karte metrics: %s", mErr)
		}
		defer func() {
			// Recoding servo fw update to Karte.
			log.Debugf(ctx, "Updating servo firmware information in Karte.")
			action.StopTime = time.Now()
			if err != nil {
				action.Status = metrics.ActionStatusFail
				action.FailReason = err.Error()
			}
			if mErr := info.GetMetrics().Create(ctx, action); mErr != nil {
				log.Debugf(ctx, "Servo update servo firmware: Metrics error: %s", mErr)
			}
		}()
	}
	run := info.NewRunner(sh.GetName())
	if forceUpdate {
		// If requested to update with force then first attempt will be with force
		// and there no second attempt.
		tryAttemptCount = 1
		tryForceUpdateAfterFail = false
	}
	req := FwUpdaterRequest{
		UseContainer:            IsContainerizedServoHost(ctx, sh),
		FirmwareChannel:         sh.GetFirmwareChannel(),
		TryAttemptCount:         tryAttemptCount,
		TryForceUpdateAfterFail: tryForceUpdateAfterFail,
		ForceUpdate:             forceUpdate,
		IgnoreVersion:           ignoreVersion,
	}
	devicesToUpdate := topology.Devices(sh.GetServoTopology(), filteredServoBoard)
	if len(devicesToUpdate) == 0 {
		return errors.Reason("servo update servo firmware: the number of servo devices to update fw is 0").Err()
	}
	failDevices := UpdateDevicesServoFw(ctx, run, req, devicesToUpdate)
	if info.GetMetrics() != nil {
		// Record every single servo device fw flash time as well as status to karte.
		for _, device := range devicesToUpdate {
			eachBoardAction := &metrics.Action{
				// TODO(b/248635230): When karte' Search API is capable of taking in asset tag,
				// change the query to use asset tag instead of using hostname.
				Hostname:   info.GetDut().Name,
				ActionKind: fmt.Sprintf(metrics.ServoEachDeviceFwUpdateKind, device.Type),
				StartTime:  startTime,
				StopTime:   time.Now(),
				Status:     metrics.ActionStatusSuccess,
			}
			var isDeviceUpdateFailed bool
			for _, failDevice := range failDevices {
				if failDevice.Type == device.Type {
					isDeviceUpdateFailed = true
					break
				}
			}
			if isDeviceUpdateFailed {
				eachBoardAction.Status = metrics.ActionStatusFail
			}
			if mErr := info.GetMetrics().Create(ctx, eachBoardAction); mErr != nil {
				log.Debugf(ctx, "Servo update servo firmware: cannot create karte metrics: %s", mErr)
			}
		}
	}
	if len(failDevices) != 0 {
		sh.State = tlw.ServoHost_NEED_REPLACEMENT
		return errors.Reason("servo update servo firmware: %d servo devices fails the update process", len(failDevices)).Err()
	}
	return nil
}

// servoFakeDisconnectDUTExec tries to unplug DUT from servo and restore the connection.
//
// @params: actionArgs should be in the format of:
// Ex: ["delay_in_ms:x", "timeout_in_ms:x"]
func servoFakeDisconnectDUTExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	// Delay to disconnect in milliseconds. Default to be 100ms.
	delayMS := argsMap.AsInt(ctx, "delay_in_ms", 100)
	// Timeout to wait to restore the connection. Default to be 2000ms.
	timeoutMS := argsMap.AsInt(ctx, "timeout_in_ms", 2000)
	disconnectCmd := fmt.Sprintf(`fakedisconnect %d %d`, delayMS, timeoutMS)
	if err := info.NewServod().Set(ctx, "root.servo_uart_cmd", disconnectCmd); err != nil {
		return errors.Annotate(err, "servod fake disconnect servo").Err()
	}
	// Formula to cover how long we wait to see the effect
	// when we convert params to seconds and then +2 seconds to apply effect.
	waitFinishExecutionTimeout := time.Duration(delayMS+timeoutMS)*time.Millisecond + 2*time.Second
	time.Sleep(waitFinishExecutionTimeout)
	return nil
}

// servoServodCCToggleExec is the servo repair action that toggles cc line off and then on.
//
// @params: actionArgs should be in the format of:
// Ex: ["off_timeout:x", "on_timeout:x"]
func servoServodCCToggleExec(ctx context.Context, info *execs.ExecInfo) error {
	ccToggleMap := info.GetActionArgs(ctx)
	// Timeout for shut down configuration channel. Default to be 10s.
	ccOffTimeout := ccToggleMap.AsInt(ctx, "off_timeout", 10)
	// Timeout for initialize configuration channel. Default to be 30s.
	ccOnTimeout := ccToggleMap.AsInt(ctx, "on_timeout", 30)
	// Turning off configuration channel.
	log.Infof(ctx, "Turn off configuration channel and wait %d seconds.", ccOffTimeout)
	if err := info.NewServod().Set(ctx, "root.servo_uart_cmd", "cc off"); err != nil {
		return errors.Annotate(err, "servod cc toggle").Err()
	}
	time.Sleep(time.Duration(ccOffTimeout) * time.Second)
	// Turning on configuration channel.
	log.Infof(ctx, "Turn on configuration channel and wait %d seconds.", ccOnTimeout)
	if err := info.NewServod().Set(ctx, servodPdRoleCmd, servodPdRoleValueSrc); err != nil {
		return errors.Annotate(err, "servod cc toggle").Err()
	}
	// "servo_dts_mode" is the servod command to enable/disable DTS mode on servo.
	// It has two value for this cmd: on and off.
	if err := info.NewServod().Set(ctx, "servo_dts_mode", "on"); err != nil {
		return errors.Annotate(err, "servod cc toggle").Err()
	}
	time.Sleep(time.Duration(ccOnTimeout) * time.Second)
	return nil
}

// servoSetEcUartCmdExec will set "ec_uart_cmd" to the specific value based on the passed in parameter.
// Before and after the set of the "ec_uart_cmd", it will toggle the value of "ec_uart_flush".
//
// @params: actionArgs should be in the format of:
// Ex: ["wait_timeout:x", "value:xxxx"]
func servoSetEcUartCmdExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	// Timeout to wait for setting the ec_uart_cmd. Default to be 1s.
	waitTimeout := argsMap.AsDuration(ctx, "wait_timeout", 1, time.Second)
	// The value of 'pd dualrole' postfix for the servod command 'ec_uart_cmd'
	value := argsMap.AsString(ctx, "value", "")
	if value == "" {
		return errors.Reason("servo set ec uart cmd: the passed in value cannot be empty").Err()
	}
	servod := info.NewServod()
	if err := servo.SetEcUartCmd(ctx, servod, value, waitTimeout); err != nil {
		return errors.Annotate(err, "servo set ec uart cmd").Err()
	}
	return nil
}

// servoPowerStateResetExec using servod command to reset power state
// to achieve the behaviour of DUT reboot to recover some servo controls depending on EC console.
//
// Some servo controls, like lid_open, requires communicating with DUT through EC UART console.
// Failure of this kinds of controls can be recovered by rebooting the DUT.
//
// @params: actionArgs should be in the format of:
// Ex: ["wait_timeout:x"]
func servoPowerStateResetExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	// Timeout to wait for resetting the power state. Default to be 1s.
	waitTimeout := argsMap.AsDuration(ctx, "wait_timeout", 1, time.Second)
	servod := info.NewServod()
	if err := servod.Set(ctx, "power_state", "reset"); err != nil {
		return errors.Annotate(err, "servo power state reset").Err()
	}
	time.Sleep(waitTimeout)
	// Get the lid_open value which requires EC console.
	lidOpen, err := servodGetString(ctx, info.NewServod(), "lid_open")
	if err != nil {
		return errors.Annotate(err, "servo power state reset").Err()
	}
	if lidOpen != "yes" && lidOpen != "not_applicable" {
		return errors.Reason("servo power state reset: still fail to contact EC console after rebooting DUT").Err()
	}
	return nil
}

const (
	// rebootCmd is the reboot command that is used to restart the servo v3.
	rebootCmd = "sync & sleep 5; reboot & sleep 60; reboot -f & sleep 10; reboot -nf & sleep 10; telinit 6"
	// runCmdInBackgroundCmd will run the cmd in the background and return immediately returns immediately without waiting for the command's completion.
	runCmdInBackgroundCmd = `( %s ) </dev/null >/dev/null 2>&1 & echo -n $!`
)

// servoHostV3RebootExec will reboot the servo host v3 and check
// whether the reboot succeed by comparing the old reboot id and the
// new reboot id after restart.
//
// @params: actionArgs should be in the format of:
// Ex: ["reboot_timeout:x"]
func servoHostV3RebootExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	// Timeout to for executing the reboot command on the labstation.
	rebootTimeout := argsMap.AsDuration(ctx, "reboot_timeout", 10, time.Second)
	run := info.DefaultRunner()
	oldBootId, err := cros.BootID(ctx, run)
	if err != nil {
		log.Debugf(ctx, "Servo Host V3 Reboot: (non-critical) could not determine the old boot id, err :%q. Continuing with reboot action.", err)
	}
	log.Debugf(ctx, "Servo Host V3 Reboot: Old boot id: %q", oldBootId)
	// Restart the device using the reboot command.
	if _, err := run(ctx, rebootTimeout, fmt.Sprintf(runCmdInBackgroundCmd, rebootCmd)); err != nil {
		return errors.Annotate(err, "servo host v3 reboot").Err()
	}
	// Wait for the complete restart, i.e. wait for the device to go
	// down and come up again.
	if restartErr := cros.WaitForRestart(ctx, info); restartErr != nil {
		return errors.Annotate(restartErr, "servo host v3 reboot").Err()
	}
	// We will compare the old and new boot IDs only when the old boot
	// ID is known.
	if oldBootId != "" {
		newBootId, err := cros.BootID(ctx, run)
		if err != nil {
			return errors.Annotate(err, "servo host v3 reboot").Err()
		}
		if newBootId == oldBootId {
			return errors.Reason("servo host v3 reboot: reboot fail as new boot id: %s equal to old boot id: %s", newBootId, oldBootId).Err()
		}
	}
	log.Debugf(ctx, "Servo Host V3 Reboot: reboot is successful")
	return nil
}

func init() {
	execs.Register("servo_host_servod_init", servodInitActionExec)
	execs.Register("servo_host_servod_stop", servodStopActionExec)
	execs.Register("servo_create_flag_to_use_recovery_mode", servodCreateFlagToUseRecoveryModeExec)
	execs.Register("servo_audit_usbkey", servoAuditUSBKeyExec)
	execs.Register("servo_v4_root_present", isRootServoPresentExec)
	execs.Register("servo_topology_update", servoTopologyUpdateExec)
	execs.Register("servo_servod_echo_host", servoServodEchoHostExec)
	execs.Register("servo_fw_need_update", servoFirmwareNeedsUpdateExec)
	execs.Register("servo_set", servoSetExec)
	execs.Register("servo_low_ppdut5", servoLowPPDut5Exec)
	execs.Register("servo_control_min_double_value", servoControlMinDoubleValueExec)
	execs.Register("servo_check_servod_control", servoCheckServodControlExec)
	execs.Register("servo_labstation_disk_cleanup", servoLabstationDiskCleanUpExec)
	execs.Register("servo_servod_old_logs_cleanup", servoServodOldLogsCleanupExec)
	execs.Register("servo_battery_charging", servoValidateBatteryChargingExec)
	execs.Register("init_dut_for_servo", initDutForServoExec)
	execs.Register("servo_update_servo_firmware", servoUpdateServoFirmwareExec)
	execs.Register("servo_fake_disconnect_dut", servoFakeDisconnectDUTExec)
	execs.Register("servo_servod_cc_toggle", servoServodCCToggleExec)
	execs.Register("servo_set_ec_uart_cmd", servoSetEcUartCmdExec)
	execs.Register("servo_power_state_reset", servoPowerStateResetExec)
	execs.Register("servo_host_v3_reboot", servoHostV3RebootExec)
}
