// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	test_api "go.chromium.org/chromiumos/config/go/test/api"
	lab_api "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectWithService connects with the service at the provided server address.
func ConnectWithService(ctx context.Context, serverAddress string) (*grpc.ClientConn, error) {
	if serverAddress == "" {
		return nil, fmt.Errorf("Cannot connect to empty service address.")
	}
	var err error
	step, ctx := build.StartStep(ctx, "Connect to server")
	defer func() { step.End(err) }()

	logging.Infof(ctx, "Trying to connect with address %q with %s timeout", serverAddress, ServiceConnectionTimeout.String())

	conn, err := grpc.Dial(serverAddress, getGrpcDialOpts(ctx, ServiceConnectionTimeout)...)
	if err != nil {
		return nil, errors.Annotate(err, "error during connecting to service address %s: ", serverAddress).Err()
	}

	return conn, nil
}

// getGrpcDialOpts provides the grpc dial options used to connect with a service.
func getGrpcDialOpts(ctx context.Context, timeout time.Duration) []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	return opts
}

// GetServerAddressFromGetContResponse gets the server address from get container response.
func GetServerAddressFromGetContResponse(resp *test_api.GetContainerResponse) (string, error) {
	if resp == nil || len(resp.Container.GetPortBindings()) == 0 {
		return "", fmt.Errorf("Cannot retrieve address from empty response.")
	}
	hostIp := resp.Container.GetPortBindings()[0].GetHostIp()
	hostPort := resp.Container.GetPortBindings()[0].GetHostPort()

	if hostIp == "" || hostPort == 0 {
		return "", fmt.Errorf("HostIp or HostPort is empty.")
	}

	return fmt.Sprintf("%s:%v", hostIp, hostPort), nil
}

// GetIpEndpoint creates IpEndpoint from provided server address.
// Server address example: (address:port) -> localhost:8080.
func GetIpEndpoint(serverAddress string) (*lab_api.IpEndpoint, error) {
	addressInfo := strings.Split(serverAddress, ":")
	if len(addressInfo) != 2 {
		return nil, fmt.Errorf("invalid dut server address!")
	}
	port, err := strconv.Atoi(addressInfo[1])
	if err != nil {
		return nil, fmt.Errorf("error during extracting port info: %s", err)
	}

	return &lab_api.IpEndpoint{Address: addressInfo[0], Port: int32(port)}, nil
}
