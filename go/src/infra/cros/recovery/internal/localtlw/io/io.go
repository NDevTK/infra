// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package io

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
	"infra/libs/sshpool"
)

// These constants avoid magic numbers in the code. E.g. if we ever
// decide to change the default port etc, this will be the single
// place to change.
const (
	// the command use for creating as well as extracting the data
	// that is being copied
	tarCmd = "tar"

	// the port number to be used for creating SSH connections to
	// the remote device.
	defaultSSHPort = 22

	// permissions for use during potential destination directory
	// creation.
	dirPermission = os.FileMode(0755)
)

// CopyFileFrom copies a single file from remote device to local
// machine. req contains the complete path of the source file on the
// remote machine, and the complete path of the destination directory
// on the local machine where the source file will be copied. The
// destination path is just the directory name, and does not include
// the filename.
func CopyFileFrom(ctx context.Context, pool *sshpool.Pool, req *tlw.CopyRequest) error {
	if err := validateInputParams(ctx, pool, req); err != nil {
		return errors.Annotate(err, "copy file from").Err()
	}
	if err := ensurePathExists(ctx, req.PathDestination, true, true); err != nil {
		return errors.Annotate(err, "copy file from").Err()
	}

	addr := net.JoinHostPort(req.Resource, strconv.Itoa(defaultSSHPort))
	client, err := pool.Get(addr)
	if err != nil {
		return errors.Annotate(err, "copy file from: failed to get client for %q from pool", addr).Err()
	}
	defer pool.Put(addr, client)
	session, err := client.NewSession()
	if err != nil {
		return errors.Annotate(err, "copy file from: failed to create SSH session").Err()
	}
	defer session.Close()

	remoteSrc := req.PathSource
	remoteFileName := filepath.Base(remoteSrc)

	// On the remote device, read the input file and create a
	// compressed tar archive. Then write it to stdout. Here the
	// '-C' flag changes the current directory to the location of
	// the source file. This ensures that the tar archive includes
	// paths relative only to this directory.
	rCmd := fmt.Sprintf("%s -c --gzip -C %s %s", tarCmd, filepath.Dir(remoteSrc), remoteFileName)
	p, err := session.StdoutPipe()
	if err != nil {
		return errors.Annotate(err, "copy file from: error with obtaining the stdout pipe").Err()
	}
	if err := session.Start(rCmd); err != nil {
		return errors.Annotate(err, "copy file from: error with starting the remote command %q", rCmd).Err()
	}

	destFileName := filepath.Join(req.PathDestination, remoteFileName)
	log.Debug(ctx, "Copy file from: %q path to new file.", destFileName)
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return errors.Annotate(err, "copy file from: error with creating temporary dir %q", tmpDir).Err()
	}
	defer os.RemoveAll(tmpDir)

	// Read from stdin and extract the contents to tmpDir. Here,
	// the '-C' flag changes the working directory to tmpDir and
	// ensures that the output is placed there.
	lCmd := exec.CommandContext(ctx, tarCmd, "-x", "--gzip", "-C", tmpDir)
	lCmd.Stdin = p
	if err := lCmd.Run(); err != nil {
		return errors.Annotate(err, "copy file from: error with running the local command").Err()
	}
	var tmpLocalFile = filepath.Join(tmpDir, remoteFileName)
	if err := os.Rename(tmpLocalFile, destFileName); err != nil {
		return errors.Annotate(err, "copy file from: moving local file %q to %q failed", tmpLocalFile, destFileName).Err()
	}
	log.Debug(ctx, "Copy file from: successfully moved %q to %q.", tmpLocalFile, destFileName)
	return nil
}

// CopyFileTo copies a single file from local machine to remote
// device. req contains the complete path of the source file on the
// local machine, and the complete path of the destination directory
// on the remote device where the source file will be copied.
func CopyFileTo(ctx context.Context, pool *sshpool.Pool, req *tlw.CopyRequest) error {
	if err := validateInputParams(ctx, pool, req); err != nil {
		return errors.Annotate(err, "copy file to").Err()
	}
	if err := ensurePathExists(ctx, req.PathSource, false, false); err != nil {
		return errors.Annotate(err, "copy file to: error while checking whether the source file exists").Err()
	}

	addr := net.JoinHostPort(req.Resource, strconv.Itoa(defaultSSHPort))
	client, err := pool.Get(addr)
	if err != nil {
		return errors.Annotate(err, "copy file to: failed to get client %q from pool", addr).Err()
	}
	defer pool.Put(addr, client)
	session, err := client.NewSession()
	if err != nil {
		return errors.Annotate(err, "copy file to: failed to create SSH session").Err()
	}
	defer session.Close()

	// Read the input path on the local machine and create a
	// compressed tar archive. Then write it to stdout. Here, the '-C'
	// flag changes the working directory to the location where the
	// input exists. This ensures that the archive includes paths only
	// relative to this directory.
	lCmd := exec.CommandContext(ctx, tarCmd, "-c", "--gzip", "-C", filepath.Dir(req.PathSource), filepath.Base(req.PathSource))
	p, err := lCmd.StdoutPipe()
	if err != nil {
		return errors.Annotate(err, "copy file to: could not obtain the stdout pipe").Err()
	}
	if err := lCmd.Start(); err != nil {
		return errors.Annotate(err, "copy file to: could not execute local command %q", lCmd).Err()
	}
	defer lCmd.Wait()
	p2, err2 := session.StdinPipe()
	if err2 != nil {
		return errors.Annotate(err, "copy file to: error with obtaining stdin pipe for the SSH Session").Err()
	}
	uploadErrors := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg1 *sync.WaitGroup) {
		defer wg1.Done()
		if _, err := io.Copy(p2, p); err != nil {
			uploadErrors <- errors.Annotate(err, "copy file to: error with copying contents from local stdout to remote stdin").Err()
		}
		defer p2.Close()
	}(&wg)

	// Read the stdin on the remote device and extract to the
	// destination path. The '-C' flag changes the current directory
	// to the destination path, and ensures that the output is placed
	// there.
	rCmd := fmt.Sprintf("%s -x --gzip -C %s", tarCmd, req.PathDestination)
	wg.Add(1)
	go func(wg2 *sync.WaitGroup) {
		defer wg2.Done()
		if err := session.Start(rCmd); err != nil {
			uploadErrors <- errors.Annotate(err, "copy file to: remote device could not read the uploaded contents").Err()
		} else if err := session.Wait(); err != nil {
			uploadErrors <- errors.Annotate(err, "copy file to: remote command did not exit cleanly").Err()
		}
	}(&wg)
	wg.Wait()

	select {
	case e, ok := <-uploadErrors:
		if ok {
			return errors.Annotate(e, "copy file to").Err()
		} else {
			// No one is closing the channel, but we want
			// to defensively handle this case.
			return nil
		}
	default:
		return nil
	}
}

func validateInputParams(ctx context.Context, pool *sshpool.Pool, req *tlw.CopyRequest) error {
	if pool == nil {
		return errors.New("validate input params: ssh pool is not initialized")
	} else if req.Resource == "" {
		return errors.New("validate input params: resource is empty")
	} else if req.PathSource == "" {
		return errors.New("validate input params: source path is empty")
	} else if req.PathDestination == "" {
		return errors.New("validate input params: destination path is empty")
	}
	log.Debug(ctx, "Source for transfer: %q.", req.PathSource)
	log.Debug(ctx, "Destination for transfer: %q.", req.PathDestination)
	log.Debug(ctx, "Resource: %q.", req.Resource)
	return nil
}

// ensurePathExists checks whether the path 'p' exists, and whether it
// is of the type indicated by 'd'. If the directory does not exist,
// the function will create it if 'c' is true. It returns any error
// encountered during checking existence and type of 'p' or during
// directory creation.
func ensurePathExists(ctx context.Context, p string, d bool, c bool) error {
	s, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			if d && c {
				log.Debug(ctx, "Ensure path exists: creating directory %q.", p)
				return os.MkdirAll(p, dirPermission)
			}
			return errors.Annotate(err, "ensure path exists: %q does not exist", p).Err()
		}
		// This means that 'err' is not known to report
		// whether or not the file or directory already
		// exists. Hence we cannot proceed with checking
		// whether the directory pre-exists, or creating
		// directory.
		return errors.Annotate(err, "ensure path exists: cannot determine if %q exists", p).Err()
	}

	if d {
		if s.IsDir() {
			log.Debug(ctx, "Ensure path exists: directory %q exists.", p)
			return nil
		} else {
			log.Debug(ctx, "Ensure path exists: path %q is a file, and not a directory.", p)
			return errors.Annotate(err, "ensure path exists: %q is a file, and not a directory", p).Err()
		}
	} else {
		if s.IsDir() {
			log.Debug(ctx, "Ensure path exists: path %q is a directory, and not a file.", p)
			return errors.Annotate(err, "ensure path exists: %q is a directory, and not a file", p).Err()
		} else {
			log.Debug(ctx, "Ensure path exists: file %q exists.", p)
			return nil
		}
	}
}
