// Copyright 2022 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package servod provides functions to manage connection and communication with servod daemon on servo-host.
package servod

import (
	"context"
	"net"
	"strconv"
	"time"

	xmlrpc_value "go.chromium.org/chromiumos/config/go/api/test/xmlrpc"
	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/docker"
	"infra/cros/recovery/internal/localtlw/localproxy"
	"infra/cros/recovery/internal/localtlw/ssh"
	"infra/cros/recovery/internal/localtlw/xmlrpc"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// ServodCallRequest holds data to call servod daemon.
type ServodCallRequest struct {
	Host        string
	Options     *tlw.ServodOptions
	SSHProvider ssh.SSHProvider
	// Containers info.
	ContainerName string
	// Call info.
	// Example: power_state:rec is methods `set` with arguments ["power_state"|"rec"]
	CallMethod    string
	CallArguments []*xmlrpc_value.Value
	CallTimeout   time.Duration
}

// CallServod executes a command on the servod daemon running on servo-host and returns the output.
// Method detect and working with all type of hosts.
func CallServod(ctx context.Context, req *ServodCallRequest) (*xmlrpc_value.Value, error) {
	switch {
	case req.Host == "":
		return nil, errors.Reason("call servod: host is ot specified").Err()
	case req.SSHProvider == nil:
		return nil, errors.Reason("call servod: SSH provider is not specified").Err()
	case req.Options == nil:
		return nil, errors.Reason("call servod: options is not specified").Err()
	case req.Options.ServodPort <= 0:
		return nil, errors.Reason("call servod: servod port is not specified").Err()
	case req.ContainerName == "":
		// regular labstation
		return callServodLabstation(ctx, req)
	case req.ContainerName == req.Host:
		return callServodOnLocalContainer(ctx, req)
	case req.ContainerName != req.Host:
		return callServodOnRemoteContainer(ctx, req)
	default:
		return nil, errors.Reason("call servod: unsupported case").Err()
	}
}
func callServodOnLocalContainer(ctx context.Context, req *ServodCallRequest) (*xmlrpc_value.Value, error) {
	log.Debugf(ctx, "Start call with %#v", req.Options)
	d, err := newDockerClient(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "call servod on local container").Err()
	}
	addr, err := d.IPAddress(ctx, req.ContainerName)
	if err != nil {
		return nil, errors.Annotate(err, "call servod on local container").Err()
	}
	log.Debugf(ctx, "Call container by IP address: %v", addr)
	c := xmlrpc.New(addr, int(req.Options.ServodPort))
	return Call(ctx, c, req.CallTimeout, req.CallMethod, req.CallArguments)
}

func callServodOnRemoteContainer(ctx context.Context, req *ServodCallRequest) (*xmlrpc_value.Value, error) {
	return nil, errors.Reason("call servod on remote container: not implemented").Err()
}

func callServodLabstation(ctx context.Context, req *ServodCallRequest) (*xmlrpc_value.Value, error) {
	// Convert hostname to the proxy name used for local when called.
	host := localproxy.BuildAddr(req.Host)
	p, err := newProxy(ctx, req.SSHProvider, host, req.Options.GetServodPort(), func(err error) {
		log.Debugf(ctx, "Fail on proxy: %s", err)
	})
	if err != nil {
		return nil, errors.Annotate(err, "call servod labstation").Err()
	}
	defer p.Close()
	newAddr := p.LocalAddr()
	host, portString, err := net.SplitHostPort(newAddr)
	if err != nil {
		return nil, errors.Annotate(err, "call servod labstation on %q", newAddr).Err()
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, errors.Annotate(err, "call servod labstation on %q", newAddr).Err()
	}
	c := xmlrpc.New(host, port)
	return Call(ctx, c, req.CallTimeout, req.CallMethod, req.CallArguments)
}

func newDockerClient(ctx context.Context) (docker.Client, error) {
	d, err := docker.NewClient(ctx)
	return d, errors.Annotate(err, "new docker client").Err()
}
