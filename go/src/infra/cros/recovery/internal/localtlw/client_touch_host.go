// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package localtlw provides local implementation of TLW Access.
package localtlw

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	xmlrpclib "go.chromium.org/chromiumos/config/go/api/test/xmlrpc"
	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/localtlw/localproxy"
	"infra/cros/recovery/internal/localtlw/ssh"
	"infra/cros/recovery/internal/localtlw/xmlrpc"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

const (
	defaultTouchhostdPort = 9992
	// Local address with dynamic port.
	localAddr = "127.0.0.1:0"
	// Local address template for remote host.
	remoteAddrFmt = "127.0.0.1:%d"
)

// CallTouchHostd executes a command on touchhostd.
func (c *tlwClient) CallTouchHostd(ctx context.Context, req *tlw.CallTouchHostdRequest) *tlw.CallTouchHostdResponse {
	// Translator to convert error to response structure.
	fail := func(err error) *tlw.CallTouchHostdResponse {
		return &tlw.CallTouchHostdResponse{
			Value: &xmlrpclib.Value{
				ScalarOneof: &xmlrpclib.Value_String_{
					String_: fmt.Sprintf("failed to call touchhostd with hostname %s: %s", req.GetResource(), err),
				},
			},
			Fault: true,
		}
	}
	// Check if the name was detected by loaded device.
	_, err := c.getDevice(ctx, req.GetResource())
	if err != nil {
		return fail(err)
	}

	val, err := callTouchHostd(ctx, req, c.sshProvider)
	if err != nil {
		return fail(err)
	}
	return &tlw.CallTouchHostdResponse{
		Value: val,
		Fault: false,
	}

}

// callTouchHostd implements the generic XMLRPC call to any API of touchhostd.
func callTouchHostd(ctx context.Context, req *tlw.CallTouchHostdRequest, sp ssh.SSHProvider) (*xmlrpclib.Value, error) {
	log.Debugf(ctx, "calling hostname %v...", req.GetResource())
	if req.GetMethod() == "" {
		return nil, errors.Reason("missing API method").Err()
	}
	if req.GetResource() == "" {
		return nil, errors.Reason("missing API resource (hostname)").Err()
	}

	// port forwarding
	newAddressStr, err := getForwardedAddress(ctx, req.GetResource(), sp)
	if err != nil {
		return nil, errors.Annotate(err, "unable to establish port forwarding").Err()
	}
	newAddress, newPort, err := addressParser(*newAddressStr)
	if err != nil {
		return nil, errors.Annotate(err, "unable to parse address").Err()
	}

	// prepare the XMLRPC call
	callTimeout := 30 * time.Second
	if req.GetTimeout().GetSeconds() > 0 {
		callTimeout = req.GetTimeout().AsDuration()
	}
	client := xmlrpc.New(newAddress, *newPort)
	val, err := callXMLRpc(ctx, client, callTimeout, req.Method, req.GetArgs())
	if err != nil {
		return nil, errors.Annotate(err, "unable to call touchhost with hostname: %s, port %q", newAddress, *newPort).Err()
	}
	return val, nil
}

// getForwardedAddress make port forwarding and return the address of forwarder.
func getForwardedAddress(ctx context.Context, hostname string, sp ssh.SSHProvider) (*string, error) {
	host := localproxy.BuildAddr(hostname)

	sc, err := sp.Get(host)
	if err != nil {
		return nil, errors.Annotate(err, "unable to establish SSH client").Err()
	}
	defer func() {
		if err := sc.Close(); err != nil {
			// TODO(b:270462604): Delete the log after finish migration.
			log.Debugf(ctx, "SSH client closed with error: %s", err)
		} else {
			// TODO(b:270462604): Delete the log after finish migration.
			log.Debugf(ctx, "SSH client closed!")
		}
	}()

	remoteAddr := fmt.Sprintf(remoteAddrFmt, defaultTouchhostdPort)
	f, err := sc.ForwardLocalToRemote(localAddr, remoteAddr, func(fErr error) {
		log.Debugf(ctx, "failed while forwarding: %s", fErr)
	})
	if err != nil {
		return nil, errors.Annotate(err, "call touchhost Pi").Err()
	}
	defer f.Close()

	newAddr := f.LocalAddr().String()
	return &newAddr, nil
}

// addressParser parses address into host and port
func addressParser(address string) (string, *int, error) {
	host, portString, err := net.SplitHostPort(address)
	if err != nil {
		return host, nil, errors.Annotate(err, "unable to split address %s", address).Err()
	}
	newPort, err := strconv.Atoi(portString)
	if err != nil {
		return host, &newPort, errors.Annotate(err, "unable to parse port %s", portString).Err()
	}
	return host, &newPort, nil
}

// callXMLRpc calls xmlrpc service with provided method and arguments.
func callXMLRpc(ctx context.Context, client *xmlrpc.XMLRpc, timeout time.Duration, method string, args []*xmlrpclib.Value) (*xmlrpclib.Value, error) {
	var iArgs []interface{}
	for _, ra := range args {
		iArgs = append(iArgs, ra)
	}
	log.Debugf(ctx, "calling touchhostd XMLRPC api with timeout %s", timeout)
	call := xmlrpc.NewCallTimeout(timeout, method, iArgs...)
	val := &xmlrpclib.Value{}
	if err := client.Run(ctx, call, val); err != nil {
		return nil, errors.Annotate(err, "unable to call touchhostd %q: %s", client, method).Err()
	}
	return val, nil
}
