// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package buildbucket

import (
	"infra/cros/internal/cmd"
)

type Client struct {
	cmdRunner cmd.CommandRunner
}

// NewClient creates a new Buildbucket client.
func NewClient(cmdRunner cmd.CommandRunner) *Client {
	return &Client{
		cmdRunner: cmdRunner,
	}
}
