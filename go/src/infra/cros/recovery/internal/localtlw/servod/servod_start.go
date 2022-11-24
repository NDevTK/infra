// Copyright 2022 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package servod provides functions to manage connection and communication with servod daemon on servo-host.
package servod

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/tlw"
	"infra/libs/sshpool"
)

// StartServodRequest holds data to start servod container.
type StartServodRequest struct {
	Host    string
	Options *tlw.ServodOptions
	SSHPool *sshpool.Pool
	// Containers info.
	ContainerName    string
	ContainerNetwork string
}

// StartServod starts servod daemon on servo-host.
// Method detect and working with all type of hosts.
func StartServod(ctx context.Context, req *StartServodRequest) error {
	switch {
	case req.Host == "":
		return errors.Reason("start servod: host is ot specified").Err()
	case req.SSHPool == nil:
		return errors.Reason("start servod: ssh pool is not specified").Err()
	case req.Options == nil:
		return errors.Reason("start servod: options is not specified").Err()
	case req.Options.GetServodPort() <= 0:
		return errors.Reason("start servod: servod port is not specified").Err()
	case req.ContainerName == "":
		// regular labstation
		return startServodLabstation(ctx, req)
	case req.ContainerName == req.Host:
		return startServodOnLocalContainer(ctx, req)
	case req.ContainerName != req.Host:
		return startServodOnRemoteContainer(ctx, req)
	default:
		return errors.Reason("start servod: unsupported case").Err()
	}
}

func startServodOnLocalContainer(ctx context.Context, req *StartServodRequest) error {
	return errors.Reason("start servod on local container: not implemented").Err()
}

func startServodOnRemoteContainer(ctx context.Context, req *StartServodRequest) error {
	return errors.Reason("start servod on remote container: not implemented").Err()
}

func startServodLabstation(ctx context.Context, req *StartServodRequest) error {
	if stat, err := getServodStatus(ctx, req.Host, req.Options.GetServodPort(), req.SSHPool); err != nil {
		return errors.Annotate(err, "start servod on labstation").Err()
	} else if stat == servodRunning {
		// Servod is running already.
		return nil
	}
	if err := startServod(ctx, req.Host, req.Options.GetServodPort(), GenerateParams(req.Options), req.SSHPool); err != nil {
		return errors.Annotate(err, "start servod on labstation").Err()
	}
	return nil
}
