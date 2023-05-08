// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package constants

import "time"

// BucketName the bucket we want to get the blobs from.
// TODO we need to parse the service account to get the partner bucket
const BucketName = "chromeos-moblab-cienet-dev"

// SSHKeyPath the path where ssh key
const SSHKeyPath = "/usr/src/satlab_rpcserver/server/testing_rsa"

// SSHUser the username of DUT ssh
const SSHUser = "root"

// SSHPort the port of DUT
const SSHPort = "22"

// SSHConnectionTimeout timeout of ssh connection
const SSHConnectionTimeout = time.Second * 20

// SSHMaxRetry the retry time
const SSHMaxRetry = 2

// GetPeripheralInfoCommand the command of get peripheral information
const GetPeripheralInfoCommand = "fwupdmgr get-devices --json"

const ListFirmwareCommand = "fwid=`timeout 5 crossystem fwid`;" +
	"model=`timeout 5 cros_config / name`;" +
	"fw_update=`timeout 5 chromeos-firmwareupdate --manifest`;" +
	"printf \"{\\\"fwid\\\": \\\"%s\\\",\\\"model\\\": \\\"%s\\\", \\\"fw_update\\\":%s}\" $fwid $model \"$fw_update\""
