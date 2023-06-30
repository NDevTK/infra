// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
)

// Permissions is the default file permissions for log files.
// Currently, we allow everyone to read and write and nobody to execute.
const defaultFilePermissions fs.FileMode = 0666

// DmesgExec grabs dmesg and persists the file into the log directory.
// DmesgExec fails if and only if the dmesg executable doesn't exist or returns nonzero.
//
// This exec function accepts the following parameters from the action:
// human_readable: whether the dmesg output is expected to be in human-readlable form.
// create_crashinfo_dir: whether the subdirectory for crashinfo needs to be created.
func dmesgExec(ctx context.Context, info *execs.ExecInfo) error {
	argMap := info.GetActionArgs(ctx)
	logRoot := info.GetLogRoot()
	if argMap.AsBool(ctx, "create_crashinfo_dir", false) {
		logRoot = filepath.Join(logRoot, fmt.Sprintf("crashinfo.%s", info.GetActiveResource()))
	}
	run := info.DefaultRunner()
	log := info.NewLogger()
	var output string
	var err error
	if argMap.AsBool(ctx, "human_readable", true) {
		output, err = run(ctx, time.Minute, "dmesg", "-H")
	} else {
		output, err = run(ctx, time.Minute, "dmesg")
	}
	if err != nil {
		return errors.Annotate(err, "dmesg exec").Err()
	}
	// Output is non-empty and dmesg ran successfully.

	// Attempting to create a directory that already exists will not
	// result in an error, hence we can just create this new directory
	// without checking whether it already exists.
	if err = os.MkdirAll(logRoot, os.ModePerm); err != nil {
		return errors.Annotate(err, "dmesg exec").Err()
	}
	f := filepath.Join(logRoot, "dmesg")
	log.Debugf("dmesg path to safe: %s", f)
	ioutil.WriteFile(f, []byte(output), defaultFilePermissions)
	return nil
}

// copyToLogsExec grabs a file or directory from the host and copy to
// the log directory.
//
// This exec function accepts the following parameters from the action:
// src_host_type: specifies the type of source, options are "dut" and "servo_host"
// src_path: specifies the source for copy operation.
// src_type: specifies whether the source is a file or a directory, options are "file" and "dir".
// use_host_dir: specifies whether the a subdirectory with the resource-name needs to be created.
// dest_suffix: specifies any subdirectory that needs to be created within the source path.
// filename: target name for the copied file. Default value is the complete file name of the source.
func copyToLogsExec(ctx context.Context, info *execs.ExecInfo) error {
	argMap := info.GetActionArgs(ctx)
	fullPath := argMap.AsString(ctx, "src_path", "")
	if fullPath == "" {
		return errors.Reason("copy to logs: src_path is empty or not provided").Err()
	}
	type srcType string
	srcTypeArg := srcType(argMap.AsString(ctx, "src_type", "file"))
	type hostType string
	srcHostType := hostType(argMap.AsString(ctx, "src_host_type", ""))
	const (
		dutHostType   = hostType("dut")
		servoHostType = hostType("servo_host")
	)
	var resource string
	switch srcHostType {
	case dutHostType:
		resource = info.GetDut().Name
	case servoHostType:
		resource = info.GetChromeos().GetServo().GetName()
	default:
		return errors.Reason("copy to logs: src_host_type %q is either empty or an un-recognized value", srcHostType).Err()
	}
	run := info.NewRunner(resource)
	log := info.NewLogger()
	logRoot := info.GetLogRoot()
	const (
		fileType = srcType("file")
		dirType  = srcType("dir")
	)
	var testCmdFlag string
	switch srcTypeArg {
	case fileType:
		testCmdFlag = "-f"
	case dirType:
		testCmdFlag = "-d"
	default:
		return errors.Reason("copy to logs: src_type %q is either empty, or an un-recognized value", srcTypeArg).Err()
	}
	if _, err := run(ctx, time.Minute, "test", testCmdFlag, fullPath); err != nil {
		return errors.Annotate(err, "copy to logs: the src_file %s does not exist or it not a %s", fullPath, srcTypeArg).Err()
	}
	newName := strings.TrimSpace(argMap.AsString(ctx, "filename", ""))
	if newName == "" {
		newName = filepath.Base(fullPath)
	}
	// Logs will be saved to the resource folder.
	if argMap.AsBool(ctx, "use_host_dir", false) {
		logRoot = filepath.Join(logRoot, resource)
	}
	// If a suffix for destination path has been specified, include it
	// in the path for storing files on destination side.
	if destSuffix := argMap.AsString(ctx, "dest_suffix", ""); destSuffix != "" {
		logRoot = filepath.Join(logRoot, destSuffix)
	}
	if err := exec.CommandContext(ctx, "mkdir", "-p", logRoot).Run(); err != nil {
		return errors.Annotate(err, "copy to logs").Err()
	}
	switch srcTypeArg {
	case fileType:
		if newName == "" {
			return errors.Reason("copy to logs: filename is empty and could not extracted from filepath").Err()
		}
		destDir := logRoot
		log.Debugf("Copy to Logs: Attempting to collect the logs from %q to %q!", fullPath, destDir)
		if err := info.CopyFrom(ctx, resource, fullPath, destDir); err != nil {
			return errors.Annotate(err, "copy to logs").Err()
		}
	case dirType:
		if newName == "" || newName == "." || newName == ".." || newName == "/" {
			return errors.Reason("copy to logs: filename is empty and could not extracted from filepath").Err()
		}
		log.Debugf("Copy to Logs: Attempting to collect the logs from %q to %q", fullPath, logRoot)
		if err := info.CopyDirectoryFrom(ctx, resource, fullPath, logRoot); err != nil {
			return errors.Annotate(err, "copy to logs").Err()
		}
		// Note: Any values others that the above cases will be an
		// error (which would normally be caught by 'default' for this
		// switch-case). Any such cases would have already been
		// handled above when srcTypeArg is first used. We don't need
		// to repeat that logic here since it will be never executed.
	}
	return nil
}

// collectCrashDumpsExec fetches the crash dumps from the DUT.
//
// This exec function accepts the following parameters from the action:
// clean: remove the source file on the DUT, even if the copy attempt did not succeed.
// cleanup_timeout: the amount of time within which file cleanup should happen.
func collectCrashDumpsExec(ctx context.Context, info *execs.ExecInfo) error {
	logRoot := info.GetLogRoot()
	resource := info.GetActiveResource()
	infoDir := filepath.Join(logRoot, fmt.Sprintf("crashinfo.%s", resource))
	if err := os.MkdirAll(infoDir, os.ModePerm); err != nil {
		return errors.Annotate(err, "collect crash dumps execs").Err()
	}
	// Note: at this time we are only interested in the collection of
	// logs. The legacy logic to additionally create the stacktrace
	// for the crashdumps does not work correctly. We will eventually
	// include a correct implementation of the same in Paris. Bug
	// http://b/262346604 has been created to keep track of this.

	// The location of crash dumps on the DUT.
	const crashDir = "/var/spool/crash/"
	const crashFiles = crashDir + "*"
	run := info.NewRunner(resource)
	output, err := run(ctx, time.Minute, "ls", "-1", crashFiles)
	if err != nil {
		return errors.Annotate(err, "collect crash dumps exec").Err()
	}
	log := info.NewLogger()
	if output != "" {
		orphans := strings.Split(output, "\n")
		argMap := info.GetActionArgs(ctx)
		cleanUp := argMap.AsBool(ctx, "clean", true)
		timeout := argMap.AsDuration(ctx, "cleanup_timeout", 10, time.Second)
		for _, f := range orphans {
			log.Debugf("Collect Crash Dumps Exec: Attempting to collect orphan file %q", f)
			if err := info.CopyFrom(ctx, info.GetDut().Name, f, infoDir); err != nil {
				log.Debugf("Collect Crash Dumps Exec: (non-critical) error %s while copying %q to %q", err.Error(), f, infoDir)
			}
			if cleanUp {
				fPath := crashDir + f
				if _, err := run(ctx, timeout, "rm", "-f", fPath); err != nil {
					log.Debugf("Collect Crash Dumps Exec: (non-critical) error %s while removing file %q on the DUT", err.Error(), fPath)
				}
			}
		}
		if len(orphans) == 0 {
			log.Debugf("Collect Crash Dumps Exec: There are no orphaned crashdump files on the source-side, hence deleting the destination folder %q", infoDir)
			if err := os.RemoveAll(infoDir); err != nil {
				log.Debugf("Collect Crash Dumps Exec: (non-critical) error %s while removing the source directory %q", err.Error(), infoDir)
			}
		}
	}
	return nil
}

// createLogCollectionInfoExec creates a marker file that indicates
// that the log files have been collected.
//
// This exec accepts the following parameters from the action:
// info_file: the name of the info file that will be created.
func createLogCollectionInfoExec(ctx context.Context, info *execs.ExecInfo) error {
	argMap := info.GetActionArgs(ctx)
	logRoot := info.GetLogRoot()
	infoFilePath := filepath.Join(logRoot, argMap.AsString(ctx, "info_file", "log_collection_info"))
	infoFile, err := os.Create(infoFilePath)
	if err != nil {
		return errors.Annotate(err, "create log collection info").Err()
	}
	_, err = infoFile.WriteString(fmt.Sprintf("Retrieved the prior logs at %q\n", time.Now()))
	return errors.Annotate(err, "create log collection info").Err()
}

// confirmFileNotExistsExec confirms that the file mentioned in the
// parameters does not exist on the actual file system.
//
// This exec accepts the following parameters from the action:
// target_file: complete path of the file that needs to be checked.
func confirmFileNotExistsExec(ctx context.Context, info *execs.ExecInfo) error {
	argMap := info.GetActionArgs(ctx)
	logRoot := info.GetLogRoot()
	infoFilePath := filepath.Join(logRoot, argMap.AsString(ctx, "target_file", "log_collection_info"))
	_, err := os.Stat(infoFilePath)
	if err == nil {
		return errors.Reason("confirm file not exists: the file %q already exists", infoFilePath).Err()
	}
	log := info.NewLogger()
	if errors.Is(err, os.ErrNotExist) {
		log.Debugf("Confirm File Not Exists: the file %q does not exist beforehand", infoFilePath)
		return nil
	}
	log.Debugf("Confirm File Not Exists: cannot determine whether the file %q exists or not.", infoFilePath)
	return errors.Annotate(err, "confirm file not exists").Err()
}

// verboseShillLogsExec enables/disables verbose shill logs.
func verboseShillLogsExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	enabled := argsMap.AsBool(ctx, "is_enabled", true)
	runner := info.DefaultRunner()

	var cmd string
	if enabled {
		cmd = "ff_debug cellular+modem+device+dbus+manager --level -3"
	} else {
		cmd = "ff_debug reset --level 0"

	}

	if _, err := runner(ctx, info.GetExecTimeout(), cmd); err != nil {
		return errors.Annotate(err, "verbose network logs: set shill logging level").Err()
	}
	return nil
}

// verboseModemManagerLogsExec enables/disables verbose ModemManager logs.
func verboseModemManagerLogsExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	enabled := argsMap.AsBool(ctx, "is_enabled", true)
	runner := info.DefaultRunner()

	var cmd string
	if enabled {
		cmd = "modem set-logging debug"
	} else {
		cmd = "modem set-logging info"
	}

	if _, err := runner(ctx, info.GetExecTimeout(), cmd); err != nil {
		return errors.Annotate(err, "verbose network logs: set shill logging level").Err()
	}
	return nil
}

func init() {
	execs.Register("cros_dmesg", dmesgExec)
	execs.Register("cros_copy_to_logs", copyToLogsExec)
	execs.Register("cros_collect_crash_dumps", collectCrashDumpsExec)
	execs.Register("cros_create_log_collection_info", createLogCollectionInfoExec)
	execs.Register("cros_confirm_file_not_exists", confirmFileNotExistsExec)
	execs.Register("cros_set_verbose_shill_logs", verboseShillLogsExec)
	execs.Register("cros_set_verbose_mm_logs", verboseModemManagerLogsExec)
}
