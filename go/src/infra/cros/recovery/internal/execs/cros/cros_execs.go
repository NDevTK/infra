// Copyright 2021 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

// pingExec verifies the DUT is pingable.
func pingExec(ctx context.Context, info *execs.ExecInfo) error {
	return cros.WaitUntilPingable(ctx, info.GetExecTimeout(), cros.PingRetryInterval, 2, info.DefaultPinger(), info.NewLogger())
}

// sshExec verifies ssh access to the current plan's device (named by the default resource name).
func sshExec(ctx context.Context, info *execs.ExecInfo) error {
	return cros.WaitUntilSSHable(ctx, info.GetExecTimeout(), cros.SSHRetryInterval, info.DefaultRunner(), info.NewLogger())
}

// sshDUTExec verifies ssh access to the DUT.
func sshDUTExec(ctx context.Context, info *execs.ExecInfo) error {
	return cros.WaitUntilSSHable(ctx, info.GetExecTimeout(), cros.SSHRetryInterval, info.NewRunner(info.GetDut().Name), info.NewLogger())
}

// rebootExec reboots the cros DUT.
func rebootExec(ctx context.Context, info *execs.ExecInfo) error {
	if err := cros.Reboot(ctx, info.NewRunner(info.GetDut().Name), info.GetExecTimeout()); err != nil {
		return errors.Annotate(err, "cros reboot").Err()
	}
	return nil
}

// isOnStableVersionExec matches device OS version to stable CrOS version.
func isOnStableVersionExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	expected := argsMap.AsString(ctx, "os_name", "")
	if expected == "" {
		deviceType := argsMap.AsString(ctx, "device_type", components.VersionDeviceCros)
		sv, err := info.Versioner().GetVersion(ctx, deviceType, info.GetActiveResource())
		if err != nil {
			return errors.Annotate(err, "match os version").Err()
		}
		expected = sv.OSImage
	}
	if expected == "" {
		return errors.Reason("match os version: expected version is not specified").Err()
	}
	log.Debugf(ctx, "Expected version: %s", expected)
	fromDevice, err := cros.ReleaseBuildPath(ctx, info.DefaultRunner(), info.NewLogger())
	if err != nil {
		return errors.Annotate(err, "match os version").Err()
	}
	log.Debugf(ctx, "Version on device: %s", fromDevice)
	if fromDevice != expected {
		return errors.Reason("match os version: mismatch, expected %q, found %q", expected, fromDevice).Err()
	}
	return nil
}

// notOnStableVersionExec verifies devices OS is not matches stable CrOS version.
func notOnStableVersionExec(ctx context.Context, info *execs.ExecInfo) error {
	sv, err := info.Versioner().Cros(ctx, info.GetDut().Name)
	if err != nil {
		return errors.Annotate(err, "match os version").Err()
	}
	expected := sv.OSImage
	log.Debugf(ctx, "Expected version: %s", expected)
	fromDevice, err := cros.ReleaseBuildPath(ctx, info.DefaultRunner(), info.NewLogger())
	if err != nil {
		return errors.Annotate(err, "match os version").Err()
	}
	log.Debugf(ctx, "Version on device: %s", fromDevice)
	if fromDevice == expected {
		return errors.Reason("match os version: matched, expected %q, found %q", expected, fromDevice).Err()
	}
	return nil
}

// readOSVersionExec read devices OS version.
func readOSVersionExec(ctx context.Context, info *execs.ExecInfo) error {
	fromDevice, err := cros.ReleaseBuildPath(ctx, info.DefaultRunner(), info.NewLogger())
	if err != nil {
		return errors.Annotate(err, "read os version").Err()
	}
	log.Debugf(ctx, "OS version on device: %s", fromDevice)
	return nil
}

// isDefaultBootFromDiskExec confirms the resource is set to boot from disk by default.
func isDefaultBootFromDiskExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	err := cros.MatchCrossystemValueToExpectation(ctx, run, "dev_default_boot", "disk")
	return errors.Annotate(err, "default boot from disk").Err()
}

// isNotInDevModeExec confirms that the host is not in dev mode.
func isNotInDevModeExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	err := cros.MatchCrossystemValueToExpectation(ctx, run, "devsw_boot", "0")
	return errors.Annotate(err, "not in dev mode").Err()
}

// isBootedInSecureModeExec checks is device booted in secure mode.
func isBootedInSecureModeExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	if err := cros.MatchCrossystemValueToExpectation(ctx, run, "devsw_boot", "0"); err != nil {
		return errors.Annotate(err, "is booted in secure mode").Err()
	}
	out, err := run(ctx, time.Minute, "/usr/share/vboot/bin/get_gbb_flags.sh")
	if err != nil {
		return errors.Annotate(err, "is booted in secure mode").Err()
	}
	// Check if GBB flags is set as 0x0 as expected for device booted in secure mode
	if r, err := regexp.Compile(`Chrome OS GBB set flags:([0x ]*)$`); err != nil {
		return errors.Annotate(err, "is booted in secure mode").Err()
	} else if !r.MatchString(out) {
		return errors.Reason("is booted in secure mode: gbb flags are not set to 0(zero)").Err()
	}
	return nil
}

// runnerByHost return runner per specified host.
func runnerByHost(ctx context.Context, info *execs.ExecInfo) (components.Runner, error) {
	argsMap := info.GetActionArgs(ctx)
	host := argsMap.AsString(ctx, "host", "")
	switch host {
	case "dut":
		dut := info.GetDut()
		if dut == nil || dut.Name == "" {
			return nil, errors.Reason("runner by device_type: DUT does not exist or not specified").Err()
		}
		return info.NewRunner(dut.Name), nil
	default:
		return info.DefaultRunner(), nil
	}
}

// runCommandExec runs a given action exec arguments in shell.
func runCommandExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	command := argsMap.AsString(ctx, "command", "")
	if command == "" {
		return errors.Reason("run command: command not specified").Err()
	}
	run, err := runnerByHost(ctx, info)
	if err != nil {
		return errors.Annotate(err, "run shell command").Err()
	}
	log.Debugf(ctx, "Run command: %q.", command)
	if out, err := run(ctx, info.GetExecTimeout(), command); err != nil {
		return errors.Annotate(err, "run command").Err()
	} else {
		log.Debugf(ctx, "Run command: output: %s", out)
	}
	return nil
}

// runShellCommandExec runs a given action exec arguments in shell.
func runShellCommandExec(ctx context.Context, info *execs.ExecInfo) error {
	// TODO(gregorynisbet): Convert to single line command and always use linux shell.
	actionArgs := info.GetExecArgs()
	if len(actionArgs) > 0 {
		log.Debugf(ctx, "Run shell command: arguments %s.", actionArgs)
		run := info.DefaultRunner()
		if out, err := run(ctx, info.GetExecTimeout(), actionArgs[0], actionArgs[1:]...); err != nil {
			return errors.Annotate(err, "run shell command").Err()
		} else {
			log.Debugf(ctx, "Run shell command: output: %s", out)
		}
	} else {
		log.Debugf(ctx, "Run shell command: no arguments passed.")
	}
	return nil
}

// isFileSystemWritable checks whether the stateful file systems are writable.
func isFileSystemWritableExec(ctx context.Context, info *execs.ExecInfo) error {
	// N.B. Order matters here:  Encrypted stateful is loop-mounted from a file in unencrypted stateful,
	// so we don't test for errors in encrypted stateful if unencrypted fails.
	testDirs := []string{"/mnt/stateful_partition", "/var/tmp"}
	return cros.IsFileSystemWritable(ctx, info.DefaultRunner(), info.NewLogger(), testDirs)
}

// hasPythonInterpreterExec confirms the presence of a working Python interpreter.
func hasPythonInterpreterExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	_, err := run(ctx, time.Minute, `python -c "import json"`)
	switch {
	case err == nil:
		// Python detected and import is working. do nothing
		return nil
	case execs.SSHErrorCLINotFound.In(err):
		if pOut, pErr := run(ctx, time.Minute, "which python"); pErr != nil {
			return errors.Annotate(pErr, "has python interpreter: python is missing").Err()
		} else if pOut == "" {
			return errors.Reason("has python interpreter: python is missing; may be caused by powerwash").Err()
		}
		fallthrough
	default:
		return errors.Annotate(err, "has python interpreter: interpreter is broken").Err()
	}
}

// hasCriticalKernelErrorExec confirms we have seen critical file system kernel errors
func hasCriticalKernelErrorExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	// grep for stateful FS errors of the type "EXT4-fs error (device sda1):"
	command := `dmesg | grep -E "EXT4-fs error \(device $(cut -d ' ' -f 5,9 /proc/$$/mountinfo | grep -e '^/mnt/stateful_partition ' | cut -d ' ' -f 2 | cut -d '/' -f 3)\):"`
	out, _ := run(ctx, time.Minute, command)
	if out != "" {
		sample := strings.Split(out, `\n`)[0]
		// Log the first file system error.
		log.Errorf(ctx, "first file system error: %q", sample)
		return errors.Reason("has critical kernel error: saw file system error: %s", sample).Err()
	}
	// Check for other critical FS errors.
	command = `dmesg | grep "This should not happen!!  Data will be lost"`
	out, _ = run(ctx, time.Minute, command)
	if out != "" {
		return errors.Reason("has critical kernel error: saw file system error: Data will be lost").Err()
	}
	log.Debugf(ctx, "Could not determine stateful mount.")
	return nil
}

// isNotVirtualMachineExec confirms that the given DUT is not a virtual device.
func isNotVirtualMachineExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	out, err := run(ctx, time.Minute, `cat /proc/cpuinfo | grep "model name"`)
	if strings.Contains(strings.ToLower(out), "qemu") {
		return errors.Reason("is not virtual machine: qemu is a virtual machine").Err()
	}
	if err != nil {
		log.Debugf(ctx, "Is Not Virtual Machine: error while determining whether cpuinfo contains model name (non-critical):%s.", err)
	}
	return nil
}

// waitForSystemExec waits for system-service to be running.
//
// Sometimes, update_engine will take a while to update firmware, so we
// should give this some time to finish. See crbug.com/765686#c38 for details.
func waitForSystemExec(ctx context.Context, info *execs.ExecInfo) error {
	serviceName := "system-services"
	// Check the status of an upstart init script
	cmd := fmt.Sprintf("status %s", serviceName)
	r := info.DefaultRunner()
	output, err := r(ctx, time.Minute, cmd)
	if err != nil {
		return errors.Annotate(err, "wait for system").Err()
	}
	if !strings.Contains(output, "start/running") {
		return errors.Reason("wait for system: service %s not running", serviceName).Err()
	}
	return nil
}

// isToolPresentExec checks the presence of the tool on the DUT.
//
// For example, the tool "dfu-programmer" is checked by running the command:
// "hash dfu-programmer" on the DUT
// The actionArgs should be in the format of ["tools:dfu-programmer,python,..."]
func isToolPresentExec(ctx context.Context, info *execs.ExecInfo) error {
	toolMap := info.GetActionArgs(ctx)
	toolNames := toolMap.AsStringSlice(ctx, "tools", nil)
	if len(toolNames) == 0 {
		return errors.Reason("tool present: tools argument is empty or not provided").Err()
	}
	r := info.DefaultRunner()
	for _, toolName := range toolNames {
		toolName = strings.TrimSpace(toolName)
		if toolName == "" {
			return errors.Reason("tool present: tool name given in the tools argument is empty").Err()
		}
		if _, err := r(ctx, time.Minute, "hash", toolName); err != nil {
			return errors.Annotate(err, "tool present").Err()
		}
	}
	return nil
}

// crosSetGbbFlagsExec sets the GBB flags on the DUT.
func crosSetGbbFlagsExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.NewRunner(info.GetDut().Name)
	if _, err := run(ctx, info.GetExecTimeout(), "/usr/share/vboot/bin/set_gbb_flags.sh 0"); err != nil {
		log.Debugf(ctx, "Cros Set Gbb Flags: %s", err)
		return errors.Annotate(err, "cros set gbb flags").Err()
	}
	return nil
}

// crosSwitchToSecureModeExec disables booting into dev-mode on the DUT.
func crosSwitchToSecureModeExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.NewRunner(info.GetDut().Name)
	if _, err := run(ctx, info.GetExecTimeout(), "crossystem", "disable_dev_request=1"); err != nil {
		log.Debugf(ctx, "Cros Switch to Secure Mode %s", err)
		return errors.Annotate(err, "cros switch to secure mode").Err()
	}
	return nil
}

// updateCrossystemExec update the value of the command to the value passed in from the config file.
//
// the actionArgs should be in the format of ["command:....", "value:....", "check_after_update:true/false"]
func updateCrossystemExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	command := argsMap.AsString(ctx, "command", "")
	if command == "" {
		return errors.Reason("update crossystem: command cannot be empty").Err()
	}
	val := argsMap.AsString(ctx, "value", "")
	if val == "" {
		return errors.Reason("update crossystem: value cannot be empty").Err()
	}
	checkAfterUpdate := argsMap.AsBool(ctx, "check_after_update", false)
	run := info.NewRunner(info.GetDut().Name)
	return errors.Annotate(cros.UpdateCrossystem(ctx, run, command, val, checkAfterUpdate), "update crossystem").Err()
}

// logTypeCStatus logs the type-C status from the DUT's perspective.
func logTypeCStatus(ctx context.Context, info *execs.ExecInfo) error {
	const status0 = "ectool typecstatus 0"
	const status1 = "ectool typecstatus 1"
	run := info.NewRunner(info.GetDut().Name)
	out, err := run(ctx, time.Minute, status0)
	if err != nil {
		return errors.Annotate(err, "log type C status").Err()
	}
	log.Infof(ctx, "(%s) %s", status0, out)
	run(ctx, time.Minute, status0)
	out, err = run(ctx, time.Minute, status1)
	if err != nil {
		return errors.Annotate(err, "log type C status").Err()
	}
	log.Infof(ctx, "(%s) %s", status1, out)
	return nil
}

func init() {
	execs.Register("cros_ping", pingExec)
	execs.Register("cros_ssh", sshExec)
	execs.Register("cros_ssh_dut", sshDUTExec)
	execs.Register("cros_reboot", rebootExec)
	execs.Register("cros_is_on_stable_version", isOnStableVersionExec)
	execs.Register("cros_not_on_stable_version", notOnStableVersionExec)
	execs.Register("cros_read_os_version", readOSVersionExec)
	execs.Register("cros_is_default_boot_from_disk", isDefaultBootFromDiskExec)
	execs.Register("cros_is_not_in_dev_mode", isNotInDevModeExec)
	execs.Register("cros_is_booted_in_secure_mode", isBootedInSecureModeExec)
	execs.Register("cros_run_shell_command", runShellCommandExec)
	execs.Register("cros_run_command", runCommandExec)
	execs.Register("cros_is_file_system_writable", isFileSystemWritableExec)
	execs.Register("cros_has_python_interpreter_working", hasPythonInterpreterExec)
	execs.Register("cros_has_critical_kernel_error", hasCriticalKernelErrorExec)
	execs.Register("cros_is_not_virtual_machine", isNotVirtualMachineExec)
	execs.Register("cros_wait_for_system", waitForSystemExec)
	execs.Register("cros_is_tool_present", isToolPresentExec)
	execs.Register("cros_set_gbb_flags", crosSetGbbFlagsExec)
	execs.Register("cros_switch_to_secure_mode", crosSwitchToSecureModeExec)
	execs.Register("cros_update_crossystem", updateCrossystemExec)
	execs.Register("cros_log_typec_status", logTypeCStatus)
}
