// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package sshpool

import (
	"reflect"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/ssh"
)

func TestFromClientConfig(t *testing.T) {
	t.Parallel()
	Convey("FromClientConfig", t, func() {
		originalClientConfig := &ssh.ClientConfig{
			Config: ssh.Config{
				Ciphers: []string{"aes128-ctr"},
			},
			Timeout: 5 * time.Second,
			User:    "user",
		}
		c, err := FromClientConfig(originalClientConfig)
		Convey("Returns SSH client config with populated values", func() {
			So(err, ShouldBeNil)
			clientConfig := c.GetSSHConfig("")
			So(clientConfig, ShouldNotBeNil)
			So(clientConfig.Auth, ShouldEqual, c.(*config).auth)
			So(reflect.TypeOf(clientConfig.HostKeyCallback), ShouldEqual, reflect.TypeOf(ssh.InsecureIgnoreHostKey()))
			So(clientConfig.Ciphers, ShouldEqual, []string{"aes128-ctr"})
			So(clientConfig.Timeout, ShouldEqual, 5*time.Second)
			So(clientConfig.User, ShouldEqual, "user")
		})
		Convey("Returns nil ProxyConfig", func() {
			So(c.GetProxy(""), ShouldBeNil)
		})
	})
}
