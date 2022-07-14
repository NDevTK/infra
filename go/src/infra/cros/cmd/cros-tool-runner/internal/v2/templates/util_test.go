package templates

import (
	"testing"
)

func TestParsePortBindingString(t *testing.T) {
	original := "80/tcp -> 10.88.0.1:42222"
	expect := &PortBinding{
		ContainerPort: 80,
		Protocol:      "tcp",
		HostIp:        "10.88.0.1",
		HostPort:      42222,
	}
	parsed, err := TemplateUtils.parsePortBindingString(original)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if *parsed != *expect {
		t.Fatalf("Result doesn't match\nexpect: %v\nactual: %v", expect, parsed)
	}
}

func TestParseMultilinePortBindings(t *testing.T) {
	original := "80/tcp -> 10.88.0.1:42222\n81/tcp -> 0.0.0.0:42223"
	expect := []*PortBinding{
		{ContainerPort: 80, Protocol: "tcp", HostIp: "10.88.0.1", HostPort: 42222},
		{ContainerPort: 81, Protocol: "tcp", HostIp: "0.0.0.0", HostPort: 42223},
	}
	parsed, err := TemplateUtils.parseMultilinePortBindings(original)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(parsed) != len(expect) || *parsed[0] != *expect[0] || *parsed[1] != *expect[1] {
		t.Fatalf("Result doesn't match\nexpect: %v\nactual: %v", expect, parsed)
	}
}
