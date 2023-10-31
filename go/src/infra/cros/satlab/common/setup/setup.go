// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package setup

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"go.chromium.org/luci/auth"
	"google.golang.org/api/option"

	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/misc"
)

type Setup struct {
	Bucket            string
	GSAccessKeyId     string
	GSSecretAccessKey string
}

// droneSA is the pathname of the skylab drone service account key
var droneSA = fmt.Sprintf("%s/%s", site.KeyFolder, site.SkylabDroneKeyFilename)

// sa is the path name of the Satlab service account key
var sa = fmt.Sprintf("%s/%s", site.KeyFolder, site.SatlabSAFilename)

// cf is the path of the Satlab config file
var cf = fmt.Sprintf("%s/%s", site.KeyFolder, site.SatlabConfigFilename)

// StartSetup trigger the setup process for Satlab box
func (s *Setup) StartSetup(ctx context.Context) error {
	if err := s.createKeyFolder(ctx); err != nil {
		return err
	}
	// Download service account key
	if s.GSAccessKeyId != "" && s.GSSecretAccessKey != "" {
		if err := s.setupWithBoto(ctx); err != nil {
			return fmt.Errorf("failed to download key with boto key: %w", err)
		}
	} else {
		if err := s.setupWithUser(ctx); err != nil {
			return fmt.Errorf("failed to download key with user credential: %w", err)
		}
	}
	// Create symlink to skylab_drone.json.
	if err := runCmd(fmt.Sprintf("sudo ln -f %s %s", sa, droneSA)); err != nil {
		return fmt.Errorf("create skylab drone symlink: %w", err)
	}

	return nil
}

func (s *Setup) createKeyFolder(ctx context.Context) error {
	// Create key/config folder if did not exist
	if err := runCmd(fmt.Sprintf("sudo mkdir -p %s", site.RecoveryVersionDirectory)); err != nil {
		return fmt.Errorf("failed to create recovery version folder: %w", err)
	}
	if err := runCmd(fmt.Sprintf("sudo chmod -R 666 %s", site.KeyFolder)); err != nil {
		return fmt.Errorf("failed to set access for key folder: %w", err)
	}
	return nil
}

// setupWithBoto setups Satlab with provided boto key/id
func (s *Setup) setupWithBoto(ctx context.Context) error {
	if err := s.createBotoConfigFile(); err != nil {
		return fmt.Errorf("fail to create .boto config: %w", err)
	}
	if err := s.downloadKeyGsutil(); err != nil {
		return err
	}
	if err := s.downloadConfigGsutil(); err != nil {
		return err
	}
	return nil
}

// setupWithUser setups Satlab with interactive user login via a terminal
func (s *Setup) setupWithUser(ctx context.Context) error {
	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, site.DefaultAuthOptions)
	tokenSource, err := authenticator.TokenSource()
	if errors.Is(err, auth.ErrLoginRequired) {
		return fmt.Errorf("login required: run `satlab login`")
	}

	fmt.Print(site.SatlabSetupInstruction)

	// If bucket is not provided; ask user for bucket name
	if s.Bucket == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please enter your GS Bucket name (for details please read the instructions above): ")
		bucket, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read user bucket")
		}
		s.Bucket = strings.ToLower(strings.TrimSpace(bucket))
	}

	// Download and prepare the service account file from user bucket.
	client, err := storage.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()
	if err = downloadGSBucket(ctx, client, s.Bucket, site.SatlabSAFilename, sa); err != nil {
		return fmt.Errorf("download service account file: %w", err)
	}

	reboot, _ := misc.AskConfirmation("Do you want to reboot now?")
	if reboot {
		cmd := exec.Command("reboot")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("fail to reboot Satlab: %w", err)
		}
	} else {
		fmt.Println("You MUST restart to apply start Satlab.")
	}
	//  Download Satlab configuration file.
	if err = downloadGSBucket(ctx, client, s.Bucket, site.SatlabConfigFilename, cf); err != nil {
		return fmt.Errorf("failed to download Satlab config file: %w", err)
	}
	return nil
}

// downloadGSBucket downloads file from gs bucket to dest file.
func downloadGSBucket(ctx context.Context, client *storage.Client, bucket, object, dest string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("%q Error: %w", object, err)
	}
	defer rc.Close()
	if err = runCmd(fmt.Sprintf("sudo touch %s", dest)); err != nil {
		return err
	}
	if err = runCmd(fmt.Sprintf("sudo chmod 666 %s", dest)); err != nil {
		return err
	}
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, rc); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	return nil
}

// createBotoConfigFile writes the key id and secret to .boto config file
func (s *Setup) createBotoConfigFile() error {
	buf := bytes.Buffer{}

	line := func(s string) {
		buf.WriteString(s)
		buf.WriteRune('\n')
	}

	opts := func(name, value string) {
		if value != "" {
			buf.WriteString(name)
			buf.WriteString(" = ")
			buf.WriteString(value)
			buf.WriteRune('\n')
		}
	}
	line("# Autogenerated by Satlab. Do not edit.")
	line("")
	line("[Credentials]")
	opts("gs_access_key_id", s.GSAccessKeyId)
	opts("gs_secret_access_key", s.GSSecretAccessKey)
	homeDir, _ := os.UserHomeDir()
	botoCfg := filepath.Join(homeDir, ".boto")
	return os.WriteFile(botoCfg, buf.Bytes(), 0600)
}

// downloadKeyGsutil download the Satlab service account using gsutil
func (s *Setup) downloadKeyGsutil() error {
	cmd := fmt.Sprintf("sudo gsutil cp gs://%s/%s %s", s.Bucket, site.SatlabSAFilename, sa)
	return runCmd(cmd)
}

// downloadConfigGsutil download the Satlab config file using gsutil
func (s *Setup) downloadConfigGsutil() error {
	cmd := fmt.Sprintf("sudo gsutil cp gs://%s/%s %s", s.Bucket, site.SatlabConfigFilename, cf)
	return runCmd(cmd)
}

// runCmd is a wrapper to run a cmd with/without sudo.
func runCmd(c string) error {
	cmd := exec.Command("/bin/sh", "-c", c)
	return cmd.Run()
}
