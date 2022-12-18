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
		logRoot = filepath.Join(logRoot, fmt.Sprintf("crashinfo.%s", info.GetDut().Name))
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
		destFile := filepath.Join(logRoot, newName)
		log.Debugf("Copy to Logs: Attempting to collect the logs from %q to %q!", fullPath, destFile)
		if err := info.CopyDirectoryFrom(ctx, resource, fullPath, destFile); err != nil {
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

func init() {
	execs.Register("cros_dmesg", dmesgExec)
	execs.Register("cros_copy_to_logs", copyToLogsExec)
}
