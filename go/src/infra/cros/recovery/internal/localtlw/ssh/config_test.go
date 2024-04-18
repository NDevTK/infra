// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"crypto/tls"
	"reflect"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/ssh"
)

func TestFromSSHConfig(t *testing.T) {
	t.Parallel()

	Convey("fromSSHConfig", t, func() {
		Convey("Default host directive", func() {
			sshConfig := `Host *`
			c, err := fromSSHConfig(sshConfig, nil)

			So(err, ShouldBeNil)
			Convey("Returns SSH client config for empty hostname", func() {
				clientConfig := c.GetSSHConfig("")
				So(clientConfig, ShouldNotBeNil)
				So(clientConfig.Auth, ShouldResemble, c.(*config).auth)
				So(reflect.TypeOf(clientConfig.HostKeyCallback), ShouldEqual, reflect.TypeOf(ssh.InsecureIgnoreHostKey()))
				So(clientConfig.Timeout, ShouldBeZeroValue)
				So(clientConfig.User, ShouldBeEmpty)
			})
			Convey("Returns SSH client config for non-empty hostname", func() {
				clientConfig := c.GetSSHConfig("anyHostname")
				So(clientConfig, ShouldNotBeNil)
				So(clientConfig.Auth, ShouldResemble, c.(*config).auth)
				So(reflect.TypeOf(clientConfig.HostKeyCallback), ShouldEqual, reflect.TypeOf(ssh.InsecureIgnoreHostKey()))
				So(clientConfig.Timeout, ShouldBeZeroValue)
				So(clientConfig.User, ShouldBeEmpty)
			})
		})
		Convey("Restricted host directive", func() {
			sshConfig := `Host cr-*`
			c, err := fromSSHConfig(sshConfig, nil)

			So(err, ShouldBeNil)
			Convey("Returns nil for empty hostname", func() {
				So(c.GetSSHConfig(""), ShouldBeNil)
			})
			Convey("Returns nil for non-matching hostname", func() {
				So(c.GetSSHConfig("anyHostname"), ShouldBeNil)
			})
			Convey("Returns SSH client config for matching hostname", func() {
				clientConfig := c.GetSSHConfig("cr-anyHostname")
				So(clientConfig, ShouldNotBeNil)
				So(clientConfig.Auth, ShouldResemble, c.(*config).auth)
				So(reflect.TypeOf(clientConfig.HostKeyCallback), ShouldEqual, reflect.TypeOf(ssh.InsecureIgnoreHostKey()))
				So(clientConfig.Timeout, ShouldBeZeroValue)
				So(clientConfig.User, ShouldBeEmpty)
			})
		})
		Convey("SSH config with no Proxy command", func() {
			sshConfig := `Host *
	User root
	StrictHostKeyChecking no
  Ciphers 3des-cbc blowfish-cbc cast128-cbc
	ConnectTimeout 2`
			c, err := fromSSHConfig(sshConfig, nil)

			So(err, ShouldBeNil)
			Convey("Returns SSH client config with populated values", func() {
				clientConfig := c.GetSSHConfig("")
				So(clientConfig, ShouldNotBeNil)
				So(clientConfig.Auth, ShouldResemble, c.(*config).auth)
				So(reflect.TypeOf(clientConfig.HostKeyCallback), ShouldEqual, reflect.TypeOf(ssh.InsecureIgnoreHostKey()))
				So(clientConfig.Ciphers, ShouldResemble, []string{"3des-cbc", "blowfish-cbc", "cast128-cbc"})
				So(clientConfig.Timeout, ShouldEqual, 2*time.Second)
				So(clientConfig.User, ShouldEqual, "root")
			})
			Convey("Returns nil ProxyConfig", func() {
				So(c.GetProxy(""), ShouldBeNil)
			})
		})
		Convey("SSH config with Proxy command (host only)", func() {
			sshConfig := `Host *
  Hostname %h.google.com
  ProxyCommand openssl s_client -connect 1.2.3.4:443 -servername %h`
			c, err := fromSSHConfig(sshConfig, nil)

			So(err, ShouldBeNil)
			Convey("Returns default SSH client config", func() {
				clientConfig := c.GetSSHConfig("")
				So(clientConfig, ShouldNotBeNil)
				So(clientConfig.Auth, ShouldResemble, c.(*config).auth)
				So(reflect.TypeOf(clientConfig.HostKeyCallback), ShouldEqual, reflect.TypeOf(ssh.InsecureIgnoreHostKey()))
			})
			Convey("Given hostname - Returns SSH client config", func() {
				clientConfig := c.GetSSHConfig("test")
				So(clientConfig, ShouldNotBeNil)
				So(clientConfig.Auth, ShouldResemble, c.(*config).auth)
				So(reflect.TypeOf(clientConfig.HostKeyCallback), ShouldEqual, reflect.TypeOf(ssh.InsecureIgnoreHostKey()))
			})
			Convey("Given hostname:port - Returns SSH client config", func() {
				clientConfig := c.GetSSHConfig("test:22")
				So(clientConfig, ShouldNotBeNil)
				So(clientConfig.Auth, ShouldResemble, c.(*config).auth)
				So(reflect.TypeOf(clientConfig.HostKeyCallback), ShouldEqual, reflect.TypeOf(ssh.InsecureIgnoreHostKey()))
			})
			Convey("Given hostname - Returns ProxyConfig", func() {
				pc := c.GetProxy("test")
				So(pc, ShouldNotBeNil)
				So(pc.GetAddr(), ShouldEqual, "1.2.3.4:443")
				So(pc.GetConfig(), ShouldResemble, &tls.Config{
					ServerName: "test.google.com",
				})
			})
			Convey("Given hostname:default port - Returns ProxyConfig", func() {
				pc := c.GetProxy("test:22")
				So(pc, ShouldNotBeNil)
				So(pc.GetAddr(), ShouldEqual, "1.2.3.4:443")
				So(pc.GetConfig(), ShouldResemble, &tls.Config{
					ServerName: "test.google.com",
				})
			})
			Convey("Given hostname:port - Returns ProxyConfig", func() {
				pc := c.GetProxy("test:2222")
				So(pc, ShouldNotBeNil)
				So(pc.GetAddr(), ShouldEqual, "1.2.3.4:443")
				So(pc.GetConfig(), ShouldResemble, &tls.Config{
					ServerName: "test.google.com",
				})
			})
		})
		Convey("SSH config with Proxy command (host and port)", func() {
			sshConfig := `Host *
  Hostname %h.google.com
  ProxyCommand openssl s_client -connect 1.2.3.4:443 -servername %h:%p`
			c, err := fromSSHConfig(sshConfig, nil)

			So(err, ShouldBeNil)
			Convey("Given hostname - Returns ProxyConfig", func() {
				pc := c.GetProxy("test")
				So(pc, ShouldNotBeNil)
				So(pc.GetAddr(), ShouldEqual, "1.2.3.4:443")
				So(pc.GetConfig(), ShouldResemble, &tls.Config{
					ServerName: "test.google.com",
				})
			})
			Convey("Given hostname:default port - Returns ProxyConfig", func() {
				pc := c.GetProxy("test:22")
				So(pc, ShouldNotBeNil)
				So(pc.GetAddr(), ShouldEqual, "1.2.3.4:443")
				So(pc.GetConfig(), ShouldResemble, &tls.Config{
					ServerName: "test.google.com",
				})
			})
			Convey("Given hostname:port - Returns ProxyConfig", func() {
				pc := c.GetProxy("test:2222")
				So(pc, ShouldNotBeNil)
				So(pc.GetAddr(), ShouldEqual, "1.2.3.4:443")
				So(pc.GetConfig(), ShouldResemble, &tls.Config{
					ServerName: "test.google.com",
				})
			})
		})
	})
}

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
		c, err := fromClientConfig(originalClientConfig, nil)
		Convey("Returns SSH client config with populated values", func() {
			So(err, ShouldBeNil)
			clientConfig := c.GetSSHConfig("")
			So(clientConfig, ShouldNotBeNil)
			So(clientConfig.Auth, ShouldResemble, c.(*config).auth)
			So(reflect.TypeOf(clientConfig.HostKeyCallback), ShouldEqual, reflect.TypeOf(ssh.InsecureIgnoreHostKey()))
			So(clientConfig.Ciphers, ShouldResemble, []string{"aes128-ctr"})
			So(clientConfig.Timeout, ShouldEqual, 5*time.Second)
			So(clientConfig.User, ShouldEqual, "user")
		})
		Convey("Returns nil ProxyConfig", func() {
			So(c.GetProxy(""), ShouldBeNil)
		})
	})
}

func TestNewDefaultConfig(t *testing.T) {
	t.Parallel()
	Convey("NewDefaultConfig", t, func() {
		c, err := NewDefaultConfig(nil)
		Convey("Returns default SSH client config", func() {
			So(err, ShouldBeNil)
			clientConfig := c.GetSSHConfig("")
			So(reflect.TypeOf(clientConfig.HostKeyCallback), ShouldEqual, reflect.TypeOf(ssh.InsecureIgnoreHostKey()))
			So(clientConfig.Timeout, ShouldEqual, 2*time.Second)
			So(clientConfig.User, ShouldEqual, "root")
		})
		Convey("Returns nil ProxyConfig", func() {
			So(c.GetProxy(""), ShouldBeNil)
		})
	})
}
