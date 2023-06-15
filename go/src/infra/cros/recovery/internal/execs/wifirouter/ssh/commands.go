// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"context"
	"time"

	"go.chromium.org/luci/common/errors"
	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/log"
)

// TestPath runs `test <testFlag> <remotePath>` on the host and returns the
// result of the test.
//
// Returns:
//   - (true, nil) if the run exit code is 0, meaning the test passed
//   - (false, nil) if the run exit code is 1, meaning the test ran successfully
//     but did not pass
//   - (false, non-nil error) if the exit code was any other value, meaning that
//     the test did not execute successfully on the remote host
func TestPath(ctx context.Context, sshRunner Runner, testFlag, remotePath string) (bool, error) {
	runResult := sshRunner.RunForResult(ctx, 0, false, "test", testFlag, remotePath)
	exitCode := runResult.GetExitCode()
	if exitCode == 0 {
		return true, nil
	}
	if exitCode == 1 {
		return false, nil
	}
	return false, errors.Reason("failed to run 'test %s %q' on host: got unexpected exit code: %s", testFlag, remotePath, runResult.GetCommand()).Err()
}

// TestFileExists checks the remote host for a file's existence at remoteFilePath
// by running `test -f <remoteFilePath>` on the remote host.
//
// Returns:
// - (true, nil) if a regular file is found at the given path on the host
// - (false, nil) if no file was found
// - (false, non-nil error) if it is unable to determine if the file exists or not
func TestFileExists(ctx context.Context, sshRunner Runner, remoteFilePath string) (bool, error) {
	return TestPath(ctx, sshRunner, "-f", remoteFilePath)
}

// CatFile runs `cat <remoteFilePath>` on the host and returns the
// stdout (the file contents) as a string.
func CatFile(ctx context.Context, sshRunner Runner, remoteFilePath string) (string, error) {
	return sshRunner.Run(ctx, 0, "cat", remoteFilePath)
}

// Reboot runs `reboot` on the host and then waits for the reboot to
// complete. The reboot is seen as finished when the device is ssh-able again.
func Reboot(ctx context.Context, sshRunner Runner) error {
	// Start reboot over ssh.
	log.Debugf(ctx, "Rebooting host over ssh")
	if err := sshRunner.RunInBackground(ctx, 0, "reboot"); err != nil {
		return err
	}

	// Wait to give it time to shut down ssh until verifying it is back up.
	log.Debugf(ctx, "Waiting 10s for host to start rebooting before reconnecting")
	time.Sleep(10 * time.Second)
	log.Debugf(ctx, "Attempting to reconnect to host after reboot")
	return cros.WaitUntilSSHable(ctx, 3*time.Minute, 10*time.Second, sshRunner.Run, log.Get(ctx))
}

// TryAccess will attempt to run a simple bash command, `true`, on the host over
// ssh to validate that it can connect to the device and run a command.
func TryAccess(ctx context.Context, sshRunner Runner) error {
	_, err := sshRunner.Run(ctx, 0, "true")
	return err
}
