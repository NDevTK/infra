// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package platform

import (
	"context"
	"log"
	"os/exec"

	"infra/cros/satlab/satlabrpcserver/utils"
	"infra/cros/satlab/satlabrpcserver/utils/constants"
)

type Chromeos struct {
	hostIdentifier *HostIdentifier
	execCommander  utils.IExecCommander
}

func NewChromeosPlatform() IPlatform {
	return &Chromeos{
		hostIdentifier: nil,
		execCommander:  &utils.ExecCommander{},
	}
}

// GetHostIdentifier get the serial number of the device from VPD.
//
// Use the serial number from the VPD as the host identifier, if no
// serial number is set use the ethernet_mac value.
//
// Once a value has been read from the VPD cache it in memory for all
// future calls, the identifier should not change hitting the VPD
// can be slow/error prone.
func (c *Chromeos) GetHostIdentifier(ctx context.Context) (string, error) {
	if c.hostIdentifier != nil {
		return c.hostIdentifier.ID, nil
	}

	for _, key := range []string{constants.VPDKeySerialNumber, constants.VPDKeyEthernetMAC} {
		cmd := exec.CommandContext(ctx, "vpd", "-g", key)
		out, err := c.execCommander.Exec(cmd)
		if err != nil {
			log.Printf("got host identifier by `%v`", key)
			continue
		}
		c.hostIdentifier = &HostIdentifier{ID: string(out)}
		return c.hostIdentifier.ID, nil
	}

	return "", utils.HostIdentifierNotFound
}
