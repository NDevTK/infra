// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package fake

import (
	"log"
	"net"

	"github.com/gliderlabs/ssh"
)

const (
	Localhost = "127.0.0.1:0"
	Password  = "fake_password"
)

// SSHServer the object fake ssh server
type SSHServer struct {
	Listener *net.Listener
	server   *ssh.Server
}

// NewFakeServer create a fake ssh server
func NewFakeServer(handler ssh.Handler) (*SSHServer, error) {
	listener, err := net.Listen("tcp", Localhost)
	if err != nil {
		log.Printf("Can't listen to addr: %v", Localhost)
		return nil, err
	}
	server := ssh.Server{
		Handler: handler,
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			return password == Password
		},
	}

	return &SSHServer{
		Listener: &listener,
		server:   &server,
	}, nil
}

// Serve listens on the TCP network address srv.Addr
func (s *SSHServer) Serve() error {
	return s.server.Serve(*s.Listener)
}

// Close the ssh server, if there is error, returns it.
func (s *SSHServer) Close() error {
	return s.server.Close()
}

// GetAddr get which address is listening
func (s *SSHServer) GetAddr() string {
	return (*s.Listener).Addr().String()
}
