// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package taskspb

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	sync "sync"
	"time"

	"github.com/google/uuid"
	client "go.chromium.org/luci/cipd/client/cipd"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/tq"

	"infra/appengine/poros/api/entities"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	protobuf "google.golang.org/protobuf/proto"
)

func CreateAssetHandler(ctx context.Context, payload protobuf.Message) error {
	task := payload.(*AssetAdditionOrDeletionTask)
	logging.Infof(ctx, "Got %d", task.AssetInstanceId)

	// Validate the asset instance id
	assetInstance := &entities.AssetInstanceEntity{AssetInstanceId: task.AssetInstanceId}
	if err := datastore.Get(ctx, assetInstance); err != nil {
		logging.Infof(ctx, "Failed to find Asset Instance from given assetInstanceId: %v", err)
		return err
	}

	// Generate the api.textpb temp file
	assetFile, err := createAssetFile(ctx, task.AssetInstanceId)
	if err != nil {
		updateStatusLogs(ctx, task.AssetInstanceId, "STATUS_FAILED", "", err)
		return err
	}
	defer os.Remove(assetFile) // clean up

	// Fetch cel_ctl binary from CIPD
	tr, err := auth.GetRPCTransport(ctx, auth.AsSelf)
	if err != nil {
		return err
	}

	clientOps := client.ClientOptions{
		AuthenticatedClient: &http.Client{Transport: tr},
		ServiceURL:          "https://chrome-infra-packages.appspot.com",
	}
	cipdClient, err := client.NewClient(clientOps)
	if err != nil {
		logging.Infof(ctx, "Failed to initialize CIPD client: %v", err)
		return err
	}
	pin, err := cipdClient.ResolveVersion(ctx, "infra/celab/celab/linux-amd64", "dev")
	if err != nil {
		logging.Infof(ctx, "Failed to collect latest ref: %v", err)
		return err
	}
	tmpfile, err := ioutil.TempFile("", "*.asset.host.zip")
	if err != nil {
		logging.Infof(ctx, "Failed to create the temp asset file: %v", err)
		return err
	}
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	var writerSeeker io.WriteSeeker = tmpfile

	if err = cipdClient.FetchInstanceTo(ctx, pin, writerSeeker); err != nil {
		logging.Infof(ctx, "Failed to get the instance of package: %v", err)
		return err
	}

	tmpDir, err := ioutil.TempDir("", "celab-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if _, err := Unzip(tmpfile.Name(), tmpDir); err != nil {
		return err
	}

	// Run the binary
	executeCommand(ctx, tmpDir, task.Operation, assetFile, task.AssetInstanceId)
	return nil
}

func createAssetFile(ctx context.Context, assetInstanceId string) (string, error) {
	// create a host file having details about the gcp project and storage buckets
	assetFileTemplate, err := os.ReadFile("./taskspb/template/deploy.api.textpb")
	if err != nil {
		return "", err
	}
	assetConfiguration := fmt.Sprintf(string(assetFileTemplate), assetInstanceId)
	content := []byte(assetConfiguration)
	tmpfile, err := os.CreateTemp("", "*.deploy.api.textpb")
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.Write(content); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}
	logging.Infof(ctx, "Asset File name: %s", tmpfile.Name())
	return tmpfile.Name(), nil
}

func executeCommand(ctx context.Context, binaryDir string, operation string, assetFile string, assetInstanceId string) {
	celBinary := filepath.Join(binaryDir, "linux_amd64", "bin", "cel_ctl")
	cmd := exec.Command(celBinary, operation, "--builtins", "--timeout", "300", assetFile)
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()

	if err != nil {
		logging.Infof(ctx, "Command failed to start: %v", err)
		updateStatusLogs(ctx, assetInstanceId, "STATUS_FAILED", "cel_ctl command failed to start", err)
	}

	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)

	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		// In this case the command timed out as expected because we set
		// the timeout flag. Call the waitfor command which will be shorter
		// in duration and won't hold the cloud task for long.
		if strings.Contains(string(stdout), "OnHost configuration timed out") {
			enqueueWaitForTask(ctx, assetInstanceId, "waitfor")
			return
		}

		updateStatusLogs(ctx, assetInstanceId, "STATUS_FAILED", "cel_ctl command failed to run", err)
		logging.Infof(ctx, "cmd.Run() failed with %s\n", err.Error())
		return
	}

	if errStdout != nil || errStderr != nil {
		updateStatusLogs(ctx, assetInstanceId, "STATUS_FAILED", "failed to capture stdout or stderr", err)
		return
	}
	outStr, errStr := string(stdout), string(stderr)
	updateStatusLogs(ctx, assetInstanceId, "STATUS_COMPLETED", fmt.Sprintf("\nout:\n%s\nerr:\n%s\n", outStr, errStr), nil)
}

// Enqueues a task to execute waitfor command. This command is run with a timeout to
// avoid holding the queue for a long time.
func enqueueWaitForTask(ctx context.Context, assetInstanceId string, operation string) error {
	uniqId := uuid.New().String()
	return tq.AddTask(ctx, &tq.Task{
		// The body of the task. Also identifies what TaskClass to use.
		Payload: &AssetAdditionOrDeletionTask{AssetInstanceId: assetInstanceId, Operation: operation},
		// Title appears in logs and URLs, useful for debugging.
		Title: fmt.Sprintf("AssetInstanceId-%v--Operation-%v-%v", assetInstanceId, operation, uniqId),
		// How long to wait before executing this task. Not super precise.
		ETA: time.Now().Add(2 * time.Minute),
	})
}

// Update the status and Logs in datstore
func updateStatusLogs(ctx context.Context, assetInstanceId string, status string, log string, errors error) error {
	assetInstance := &entities.AssetInstanceEntity{AssetInstanceId: assetInstanceId}
	if err := datastore.Get(ctx, assetInstance); err != nil {
		return err
	}
	assetInstance.Status = status
	if errors != nil {
		assetInstance.Errors = errors.Error()
	}
	assetInstance.Logs = log
	if err := datastore.Put(ctx, assetInstance); err != nil {
		return err
	}
	return nil
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			return out, err
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

func Unzip(src string, dst string) ([]string, error) {
	var filenames []string
	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	for f := range r.File {
		dstpath := filepath.Join(dst, r.File[f].Name)
		if !strings.HasPrefix(dstpath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return nil, fmt.Errorf("%s: illegal file path", src)
		}
		if r.File[f].FileInfo().IsDir() {
			if err := os.MkdirAll(dstpath, os.ModePerm); err != nil {
				return nil, err
			}
		} else {
			if rc, err := r.File[f].Open(); err != nil {
				return nil, err
			} else {
				defer rc.Close()
				if err := os.MkdirAll(filepath.Dir(dstpath), os.ModePerm); err != nil {
					return nil, err
				}
				if of, err := os.OpenFile(dstpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0775); err != nil {
					return nil, err
				} else {
					defer of.Close()
					if _, err = io.Copy(of, rc); err != nil {
						return nil, err
					} else {
						of.Close()
						rc.Close()
						filenames = append(filenames, dstpath)
					}
				}
			}
		}
	}
	if len(filenames) == 0 {
		return nil, fmt.Errorf("zip file is empty")
	}
	return filenames, nil
}
