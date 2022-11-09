// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
	labApi "go.chromium.org/chromiumos/config/go/test/lab/api"
)

func TestParsePortBindingString(t *testing.T) {
	original := "80/tcp -> 10.88.0.1:42222"
	expect := &api.Container_PortBinding{
		ContainerPort: 80,
		Protocol:      "tcp",
		HostIp:        "10.88.0.1",
		HostPort:      42222,
	}
	parsed, err := TemplateUtils.parsePortBindingString(original)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if parsed.String() != expect.String() {
		t.Fatalf("Result doesn't match\nexpect: %v\nactual: %v", expect, parsed)
	}
}

func TestParseMultilinePortBindings(t *testing.T) {
	original := "80/tcp -> 10.88.0.1:42222\n81/tcp -> 0.0.0.0:42223"
	expect := []*api.Container_PortBinding{
		{ContainerPort: 80, Protocol: "tcp", HostIp: "10.88.0.1", HostPort: 42222},
		{ContainerPort: 81, Protocol: "tcp", HostIp: "0.0.0.0", HostPort: 42223},
	}
	parsed, err := TemplateUtils.parseMultilinePortBindings(original)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(parsed) != len(expect) || parsed[0].String() != expect[0].String() || parsed[1].String() != expect[1].String() {
		t.Fatalf("Result doesn't match\nexpect: %v\nactual: %v", expect, parsed)
	}
}

func TestParseMultilinePortBindings_ipv6BindingIgnored(t *testing.T) {
	original := "80/tcp -> 0.0.0.0:42222\n80/tcp -> :::42222"
	expect := []*api.Container_PortBinding{
		{ContainerPort: 80, Protocol: "tcp", HostIp: "0.0.0.0", HostPort: 42222},
	}
	parsed, err := TemplateUtils.parseMultilinePortBindings(original)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(parsed) != len(expect) || parsed[0].String() != expect[0].String() {
		t.Fatalf("Result doesn't match\nexpect: %v\nactual: %v", expect, parsed)
	}
}

func TestParseMultilinePortBindings_empty(t *testing.T) {
	original := "\n"
	parsed, err := TemplateUtils.parseMultilinePortBindings(original)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(parsed) != 0 {
		t.Fatalf("Expect empty port bindings returned")
	}
}

func TestEndpointToAddress(t *testing.T) {
	endpoint := &labApi.IpEndpoint{
		Address: "xyz",
		Port:    123,
	}
	address := TemplateUtils.endpointToAddress(endpoint)
	if address != "xyz:123" {
		t.Fatalf("Incorrect address conversion %s", address)
	}
}
