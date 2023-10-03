// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package constants

import "time"

// F64Epsilon Machine epsilon value for f64
const F64Epsilon = 2.2204460492503131e-16

// SSHKeyPath the path where ssh key
const SSHKeyPath = "/home/moblab/.ssh/testing_rsa"

// SSHUser the username of DUT ssh
const SSHUser = "root"

// SSHPort the port of DUT
const SSHPort = "22"

// SSHConnectionTimeout timeout of ssh connection
const SSHConnectionTimeout = time.Second * 20

// SSHRetryDelay retry delay of ssh
const SSHRetryDelay = time.Millisecond * 300

// SSHMaxRetry the retry time
const SSHMaxRetry = 2

// VPDKeySerialNumber VPD key for serial number
const VPDKeySerialNumber = "serial_number"

// VPDKeyEthernetMAC VPD key for ethernet mac
const VPDKeyEthernetMAC = "ethernet_mac"

// GetPeripheralInfoCommand the command of get peripheral information
const GetPeripheralInfoCommand = "fwupdmgr get-devices --json"

const ListFirmwareCommand = "fwid=`timeout 5 crossystem fwid`;" +
	"model=`timeout 5 cros_config / name`;" +
	"fw_update=`timeout 5 chromeos-firmwareupdate --manifest`;" +
	"printf \"{\\\"fwid\\\": \\\"%s\\\",\\\"model\\\": \\\"%s\\\", \\\"fw_update\\\":%s}\" $fwid $model \"$fw_update\""

const UpdateFirmwareCommand = "/usr/sbin/chromeos-firmwareupdate --mode autoupdate --force"

const CheckDUTIsConnectedCommand = "timeout 2 cat /etc/lsb-release"

const ChromeosTestImageReleaseTrack = "chromeos_release_track=testimage-channel"
