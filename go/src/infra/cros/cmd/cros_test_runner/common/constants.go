// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import "time"

// All common constants used throughout the service.
const (
	HwSwarmingBotIdPrefix                  = "crossk-"
	ServiceConnectionTimeout               = 5 * time.Minute
	CtrCipdPackage                         = "chromiumos/infra/cros-tool-runner/${platform}"
	ContainerDefaultNetwork                = "host"
	LabDockerKeyFileLocation               = "/creds/service_accounts/skylab-drone.json"
	LroSleepTime                           = 5 * time.Second
	GcsPublishTestArtifactsDir             = "/tmp/gcs-publish-test-artifacts/"
	TKOPublishTestArtifactsDir             = "/tmp/tko-publish-test-artifacts/"
	StainlessUrlPrefix                     = "https://stainless.corp.google.com/browse/"
	TesthausUrlPrefix                      = "https://cros-test-analytics.appspot.com/p/chromeos/logs/browse/"
	GcsUrlPrefix                           = "https://pantheon.corp.google.com/storage/browser/"
	HwTestCtrInputPropertyName             = "$chromeos/cros_tool_runner"
	CftServiceMetadataFileName             = ".cftmeta"
	CftServiceMetadataLineContentSeparator = "="
	CftServiceMetadataServicePortKey       = "SERVICE_PORT"
	TestDidNotRunErr                       = "Test did not run"
	CtrCancelingCmdErrString               = "canceling Cmd"
)
