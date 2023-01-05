// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/system/signals"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"infra/cros/cmd/cros-tool-runner/internal/v2/templates"
)

var serverCleanup = &serverStateManager{}

// NewContainerServer returns a new gRPC server for container services.
func NewContainerServer() (*grpc.Server, func()) {
	containerServer := &ContainerServerImpl{
		executor:          &DefaultCommandExecutor{},
		templateProcessor: &templates.RequestRouter{},
		containerLookuper: &templates.TemplateUtils,
	}
	// Only unary interceptor is needed as CTRv2 has no streaming endpoint.
	s := grpc.NewServer(grpc.UnaryInterceptor(panicInterceptor))
	destructor := serverCleanup.cleanup
	api.RegisterCrosToolRunnerContainerServiceServer(s, containerServer)
	reflection.Register(s)
	return s, destructor
}

// panicInterceptor implements grpc.UnaryServerInterceptor to handle panic
// (caused by bugs) with proper cleanup for CTRv2 container service.
func panicInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	defer serverCleanup.handlePanic()
	return handler(ctx, req)
}

// StartServer starts server on the requested port.
func StartServer(port int) int {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return 1
	}

	grpcServer, destructor := NewContainerServer()

	errChan := make(chan error, 1)
	stopChan := make(chan os.Signal, 1)

	// Bind signal to stopChan
	signal.Notify(stopChan, signals.Interrupts()...)

	// Start server in a goroutine, send errors to errChan
	go func() {
		log.Printf("server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- err
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Graceful stop server
	defer func() {
		grpcServer.GracefulStop()
		destructor()
	}()

	// Wait for channel operations
	select {
	case err := <-errChan:
		log.Println("fatal error:", err)
		return 1
	case <-stopChan:
		log.Println("interrupt signal received")
	}
	return 0
}
