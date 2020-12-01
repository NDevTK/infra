// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"infra/libs/lro"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/chromiumos/config/go/api/test/tls"
	"go.chromium.org/chromiumos/config/go/api/test/tls/dependencies/longrunning"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"infra/libs/sshpool"
)

type server struct {
	tls.UnimplementedCommonServer
	grpcServ *grpc.Server
	// wiringConn is a connection to the wiring service.
	wiringConn *grpc.ClientConn
	clientPool *sshpool.Pool
	sshConfig  *ssh.ClientConfig
	lroMgr     *lro.Manager
}

// Option to use to create a new TLS server.
type Option func(*server) error

func newServer(c *grpc.ClientConn, options ...Option) (*server, error) {
	s := server{
		grpcServ:   grpc.NewServer(),
		wiringConn: c,
		sshConfig: &ssh.ClientConfig{
			User: "root",
			// We don't care about the host key for DUTs.
			// Attackers intercepting our connections to DUTs is not part
			// of our attack profile.
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         5 * time.Second,
			// Use the well known testing RSA key as the default SSH auth
			// method.
			Auth: []ssh.AuthMethod{ssh.PublicKeys(sshSigner)},
		},
	}
	for _, option := range options {
		if err := option(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func (s *server) Serve(l net.Listener) error {
	s.clientPool = sshpool.New(s.sshConfig)
	defer s.clientPool.Close()
	s.lroMgr = lro.New()
	defer s.lroMgr.Close()

	tls.RegisterCommonServer(s.grpcServ, s)
	longrunning.RegisterOperationsServer(s.grpcServ, s.lroMgr)
	return s.grpcServ.Serve(l)
}

func (s *server) GracefulStop() {
	s.grpcServ.GracefulStop()
}

func (s *server) ProvisionDut(ctx context.Context, req *tls.ProvisionDutRequest) (*longrunning.Operation, error) {
	op := s.lroMgr.NewOperation()
	go s.provision(req, op.Name)

	return op, status.Error(codes.OK, "ProvisionDut started")
}

func (s *server) ExecDutCommand(req *tls.ExecDutCommandRequest, stream tls.Common_ExecDutCommandServer) error {
	// Batch size of stdout, stderr.
	const messageSize = 5000

	ctx := stream.Context()

	resp := &tls.ExecDutCommandResponse{
		ExitInfo: &tls.ExecDutCommandResponse_ExitInfo{
			Started: false,
			Status:  255,
		},
	}

	addr, err := s.getSSHAddr(ctx, req.GetName())

	if err != nil {
		resp.ExitInfo.ErrorMessage = err.Error()
		stream.Send(resp)
		return status.Errorf(codes.FailedPrecondition, err.Error())
	}

	c, err := s.clientPool.Get(addr)
	if err != nil {
		resp.ExitInfo.ErrorMessage = err.Error()
		_ = stream.Send(resp)
		return status.Errorf(codes.FailedPrecondition, fmt.Sprintf("ExecDutCommand %s %#v: %s", req.GetName(), req.GetCommand(), err))
	}
	defer s.clientPool.Put(addr, c)
	session, err := c.NewSession()
	if err != nil {
		resp.ExitInfo.ErrorMessage = err.Error()
		_ = stream.Send(resp)
		return status.Errorf(codes.FailedPrecondition, fmt.Sprintf("ExecDutCommand %s %#v: %s", req.GetName(), req.GetCommand(), err))
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	// Reading stdout of session and stream to client.
	stdoutReader, stdoutReaderErr := session.StdoutPipe()
	if stdoutReaderErr != nil {
		return status.Errorf(codes.FailedPrecondition, fmt.Sprintf("ExecDutCommand %s %#v: %s", req.GetName(), req.GetCommand(), stdoutReaderErr))
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		stdout := make([]byte, messageSize)
		stdoutResp := &tls.ExecDutCommandResponse{}
		for {
			stdoutN, stdoutReaderErr := stdoutReader.Read(stdout)
			if stdoutN > 0 {
				stdoutResp.Stdout = stdout[:stdoutN]
				_ = stream.Send(stdoutResp)
			}
			if stdoutReaderErr != nil {
				break
			}
		}
	}()

	// Reading stderr of session and stream to client.
	stderrReader, stderrReaderErr := session.StderrPipe()
	if stderrReaderErr != nil {
		return status.Errorf(codes.FailedPrecondition, fmt.Sprintf("ExecDutCommand %s %#v: %s", req.GetName(), req.GetCommand(), stderrReaderErr))
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		stderr := make([]byte, messageSize)
		stderrResp := &tls.ExecDutCommandResponse{}
		for {
			stderrN, stderrReaderErr := stderrReader.Read(stderr)
			if stderrN > 0 {
				stderrResp.Stderr = stderr[:stderrN]
				_ = stream.Send(stderrResp)
			}
			if stderrReaderErr != nil {
				break
			}
		}
	}()

	defer session.Close()

	args := req.GetArgs()
	if len(args) == 0 {
		err = session.Run(req.GetCommand())
	} else {
		err = session.Run(req.GetCommand() + " " + strings.Join(args, " "))
	}

	resp.ExitInfo.Started = true

	switch err := err.(type) {
	case nil:
		resp.ExitInfo.Status = 0
	case *ssh.ExitError:
		resp.ExitInfo.Status = int32(err.Waitmsg.ExitStatus())
		if err.Waitmsg.Signal() != "" {
			resp.ExitInfo.Signaled = true
		}
		resp.ExitInfo.ErrorMessage = err.Error()
	case *ssh.ExitMissingError:
		resp.ExitInfo.ErrorMessage = err.Error()
	default:
		resp.ExitInfo.ErrorMessage = err.Error()
	}

	_ = stream.Send(resp)

	return nil
}

// getSSHAddr returns the SSH address to use for the DUT, through the wiring service.
func (s *server) getSSHAddr(ctx context.Context, name string) (string, error) {
	c := tls.NewWiringClient(s.wiringConn)
	resp, err := c.OpenDutPort(ctx, &tls.OpenDutPortRequest{
		Name: name,
		Port: 22,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", resp.GetAddress(), resp.GetPort()), nil
}

func (s *server) FetchCrashes(req *tls.FetchCrashesRequest, stream tls.Common_FetchCrashesServer) error {
	const (
		// Largest size of blob or coredumps to include in an individual response.
		// Note that, due to serialization overhead or small metadata fields, protos returned
		// might be slightly larger than this.
		protoChunkSize = 1024 * 1024
		// Location of the serializer binary on disk.
		serializerPath = "/usr/local/sbin/crash_serializer"
	)

	ctx := stream.Context()

	addr, err := s.getSSHAddr(ctx, req.Dut)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "Failed to get address of %s: %s", req.Dut, err)
	}

	c, err := s.clientPool.Get(addr)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "Failed to get client pool for %s: %s", req.Dut, err)
	}
	defer s.clientPool.Put(addr, c)
	session, err := c.NewSession()
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "Failed to start ssh session for %s: %s", req.Dut, err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "Failed to get stdout: %s", err)
	}

	stderrReader, err := session.StderrPipe()
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "Failed to get stderr: %s", err)
	}
	stderr := bufio.NewScanner(stderrReader)

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	// Grab stderr concurrently to reading the protos.
	go func() {
		defer wg.Done()

		for stderr.Scan() {
			log.Printf("crash_serializer: %s\n", stderr.Text())
		}
		if err := stderr.Err(); err != nil {
			log.Printf("Failed to get stderr: %s\n", err)
		}
	}()

	args := []string{serializerPath, fmt.Sprintf("--chunk_size=%d", protoChunkSize)}
	if req.FetchCore {
		args = append(args, "--fetch_coredumps")
	}

	err = session.Start(strings.Join(args, " "))
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "Failed to run serializer: %s", err.Error())
	}

	var sizeBytes [8]byte
	crashResp := &tls.FetchCrashesResponse{}

	var protoBytes bytes.Buffer

	for {
		// First, read the size of the proto.
		n, err := io.ReadFull(stdout, sizeBytes[:])
		if err != nil {
			if n == 0 && err == io.EOF {
				// We've come to the end of the stream -- expected condition.
				break
			}
			// Read only a partial int. Abort.
			return status.Errorf(codes.Unavailable, "Failed to read a size: %s", err.Error())
		}
		size := binary.BigEndian.Uint64(sizeBytes[:])

		// Next, read the actual proto and parse it.
		if n, err := io.CopyN(&protoBytes, stdout, int64(size)); err != nil {
			return status.Errorf(codes.Unavailable, "Failed to read complete proto. Read %d bytes but wanted %d. err: %s", n, size, err)
		}
		// CopyN guarantees that n == protoByes.Len() == size now.

		if err := proto.Unmarshal(protoBytes.Bytes(), crashResp); err != nil {
			return status.Errorf(codes.Internal, "Failed to unmarshal proto: %s; %v", err.Error(), protoBytes.Bytes())
		}
		protoBytes.Reset()
		_ = stream.Send(crashResp)
	}

	if err := session.Wait(); err != nil {
		return status.Errorf(codes.Internal, "Failed to execute crash_serializer: %s", err.Error())
	}

	return nil
}
