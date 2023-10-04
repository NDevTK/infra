// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package utils

import (
	"log"
	"math"
	"os"

	"golang.org/x/crypto/ssh"

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
