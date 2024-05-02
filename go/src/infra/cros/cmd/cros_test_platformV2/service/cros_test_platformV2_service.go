// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"log"

	"google.golang.org/grpc"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

// Populate as needed.

type LocalState struct{}
type CrosTestRunnerServer struct {
	skylab_test_runner.UnimplementedCrosTestRunnerServiceServer

	metadata *ServerMetadata
	server   *grpc.Server

	sk *LocalState
}

func NewCTPv2Server(metadata *ServerMetadata) (*CrosTestRunnerServer, func(), error) {
	var conns []*grpc.ClientConn
	closer := func() {
		for _, conn := range conns {
			conn.Close()
		}
		conns = nil
	}

	return &CrosTestRunnerServer{metadata: metadata}, closer, nil
}

func (server *CrosTestRunnerServer) Start() error {
	// l, err := net.Listen("tcp", fmt.Sprintf(":%d", server.metadata.Port))
	// if err != nil {
	// 	return fmt.Errorf("failed to create listener at %d", server.metadata.Port)
	// }
	// // Construct state keeper to be used throughout the whole server session
	// server.sk = server.ConstructStateKeeper()

	// server.server = grpc.NewServer()

	// // TODO proto + blah
	// skylab_test_runner.RegisterCrosTestRunnerServiceServer(server.server, server)
	// reflection.Register(server.server)

	// log.Println("cros-test-runner-service listen to request at ", l.Addr().String())

	// return server.server.Serve(l)
	return nil
}

func (server *CrosTestRunnerServer) ConstructStateKeeper() *LocalState {
	sk := &LocalState{}

	return sk
}

// GRPC todo...
func (server *CrosTestRunnerServer) Execute(ctx context.Context, req *skylab_test_runner.ExecuteRequest) (*skylab_test_runner.ExecuteResponse, error) {
	log.Println("Received ExecuteRequest: ", req)
	out := &skylab_test_runner.ExecuteResponse{}

	log.Println("Execution finished successfully!")
	return out, nil
}
