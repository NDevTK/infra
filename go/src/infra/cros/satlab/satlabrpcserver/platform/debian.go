// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package platform

import (
	"context"
	"os/exec"

	"infra/cros/satlab/satlabrpcserver/utils"
)

type Debian struct {
	hostIdentifier *HostIdentifier
	execCommander  utils.IExecCommander
}

func NewDebianPlatform() IPlatform {
	return &Debian{
		hostIdentifier: nil,
		execCommander:  &utils.ExecCommander{},
	}
}

// GetHostIdentifier get a stable identifier for the machine
//
// re-install OS changes `/etc/machine-id`
func (d *Debian) GetHostIdentifier(ctx context.Context) (string, error) {
	if d.hostIdentifier != nil {
		return d.hostIdentifier.ID, nil
	}

	cmd := exec.CommandContext(ctx, "cat", "/etc/machine-id")
	out, err := d.execCommander.Exec(cmd)
	if err != nil {
		return "", err
	}

	d.hostIdentifier = &HostIdentifier{ID: string(out)}
	return d.hostIdentifier.ID, nil
}
