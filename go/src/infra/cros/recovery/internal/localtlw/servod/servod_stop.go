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

// StopServodRequest holds data to stop servod container.
type StopServodRequest struct {
	Host    string
	Options *tlw.ServodOptions
	SSHPool *sshpool.Pool
	// Containers info.
	ContainerName    string
	ContainerNetwork string
}

// StopServod stops servod daemon on servo-host.
// Method detect and working with all type of hosts.
func StopServod(ctx context.Context, req *StopServodRequest) error {
	switch {
	case req.Host == "":
		return errors.Reason("stop servod: host is ot specified").Err()
	case req.SSHPool == nil:
		return errors.Reason("stop servod: ssh pool is not specified").Err()
	case req.Options == nil:
		return errors.Reason("stop servod: options is not specified").Err()
	case req.Options.GetServodPort() <= 0:
		return errors.Reason("stop servod: servod port is not specified").Err()
	case req.ContainerName == "":
		// regular labstation
		return stopServodLabstation(ctx, req)
	case req.ContainerName == req.Host:
		return stopServodOnLocalContainer(ctx, req)
	case req.ContainerName != req.Host:
		return stopServodOnRemoteContainer(ctx, req)
	default:
		return errors.Reason("stop servod: unsupported case").Err()
	}
}

func stopServodOnLocalContainer(ctx context.Context, req *StopServodRequest) error {
	return errors.Reason("stop servod on local container: not implemented").Err()
}

func stopServodOnRemoteContainer(ctx context.Context, req *StopServodRequest) error {
	return errors.Reason("stop servod on remote container: not implemented").Err()
}

func stopServodLabstation(ctx context.Context, req *StopServodRequest) error {
	if stat, err := getServodStatus(ctx, req.Host, req.Options.GetServodPort(), req.SSHPool); err != nil {
		return errors.Annotate(err, "stop servod on labstation").Err()
	} else if stat == servodNotRunning {
		// Servod is not running.
		return nil
	}
	if err := stopServod(ctx, req.Host, req.Options.GetServodPort(), req.SSHPool); err != nil {
		return errors.Annotate(err, "stop servod on labstation").Err()
	}
	return nil
}
