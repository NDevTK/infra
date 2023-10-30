// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package utils

import (
	"context"
	"log"
	"math"
	"os"

	"golang.org/x/crypto/ssh"

	cmd_common "infra/cros/cmd/common_lib/common"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/satlabrpcserver/utils/constants"
)

// ReadSSHKey read a ssh private key file and then parse it to `ssh.Signer`
func ReadSSHKey(path string) (ssh.Signer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Can't read the ssh private key from %v", path)
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(b)
	if err != nil {
		log.Printf("Parse private key error, got %v", err)
		return nil, err
	}
	return signer, nil
}

// NearlyEqual check two float points are nearly equal.
func NearlyEqual(a, b float64) bool {
	return math.Abs(a-b) <= constants.F64Epsilon*(math.Abs(a)+math.Abs(b))
}

func AddLoggingContext(ctx context.Context) context.Context {
	// source log file from env-var
	logfilename := site.GetRPCServerLogFile()
	// append logs to the existing logfile
	logFile, err := os.OpenFile(logfilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Unable to open log file %v", err)
	}
	// format:
	// 1. time
	// 2. logging_level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
	// 3. process_id
	// 4. filename
	// Last: message to output
	format := `[%{time:2006-01-02T15:04:05.00Z07:00} | %{level:-8s} | pid:%{pid} | %{shortfile}] ` +
		`%{message}`
	logCfg := cmd_common.LoggerConfig{Out: logFile, Format: format}
	return logCfg.Use(ctx)
}
