// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servod

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/localtlw/ssh"
)

// proxy holds info to perform proxy confection to servod daemon.
type proxy struct {
	f *ssh.Forwarder
}

const (
	// Local address with dynamic port.
	localAddr = "127.0.0.1:0"
	// Local address template for remote host.
	remoteAddrFmt = "127.0.0.1:%d"
)

// newProxy creates a new proxy with forward from remote to local host.
// Function is using a goroutine to listen and handle each incoming connection.
// Initialization of proxy is going asynchronous after return proxy instance.
func newProxy(ctx context.Context, provider ssh.SSHProvider, host string, remotePort int32, errFuncs ...func(error)) (*proxy, error) {
	c, err := provider.GetContext(ctx, host)
	if err != nil {
		return nil, errors.Annotate(err, "new proxy for %q", host).Err()
	}
	defer func() { provider.Put(host, c) }()

	remoteAddr := fmt.Sprintf(remoteAddrFmt, remotePort)
	f, err := c.ForwardLocalToRemote(localAddr, remoteAddr, func(error) {
		for _, ef := range errFuncs {
			ef(err)
		}
	})
	if err != nil {
		return nil, errors.Annotate(err, "new proxy for %q", host).Err()
	}
	return &proxy{f: f}, nil
}

// Close closes proxy and used resources.
func (p *proxy) Close() error {
	err := p.f.Close()
	return errors.Annotate(err, "close proxy").Err()
}

// LocalAddr provides assigned local address.
// Example: 127.0.0.1:23456
func (p *proxy) LocalAddr() string {
	return p.f.LocalAddr().String()
}
