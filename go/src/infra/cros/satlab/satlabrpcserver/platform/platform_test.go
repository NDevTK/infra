// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package platform

import (
	"context"
	"errors"
	"testing"

	"infra/cros/satlab/satlabrpcserver/fake"
)

func TestChromeosGetHostIdentifierShouldWork(t *testing.T) {
	expectedRes := "expected result"
	c := NewChromeosPlatform().(*Chromeos)
	c.execCommander = &fake.FakeCommander{
		CmdOutput: expectedRes,
	}

	id, err := c.GetHostIdentifier(context.Background())
	if err != nil {
		t.Errorf("get host identifier from chromeos should work, but got an {%v}", err)
	}

	if id != expectedRes {
		t.Errorf("Expected {%v}, got {%v}", expectedRes, id)
	}
}

func TestChromeosGetHostIdentifierCachedShouldWork(t *testing.T) {
	expectedRes := "expected result"
	c := NewChromeosPlatform().(*Chromeos)
	c.hostIdentifier = &HostIdentifier{ID: expectedRes}
	c.execCommander = &fake.FakeCommander{
		Err: errors.New("got an error"),
	}

	id, err := c.GetHostIdentifier(context.Background())
	if err != nil {
		t.Errorf("get host identifier from chromeos should work, but got an {%v}", err)
	}

	if id != expectedRes {
		t.Errorf("Expected {%v}, got {%v}", expectedRes, id)
	}
}

func TestChromeosGetHostIdentifierShouldFail(t *testing.T) {
	c := NewChromeosPlatform().(*Chromeos)
	c.execCommander = &fake.FakeCommander{
		Err: errors.New("got an error"),
	}

	_, err := c.GetHostIdentifier(context.Background())
	if err == nil {
		t.Errorf("get host identifier from chromeos should fail")
	}
}

func TestDebianGetHostIdentifierShouldWork(t *testing.T) {
	expectedRes := "expected result"
	c := NewDebianPlatform().(*Debian)
	c.execCommander = &fake.FakeCommander{
		CmdOutput: expectedRes,
	}

	id, err := c.GetHostIdentifier(context.Background())
	if err != nil {
		t.Errorf("get host identifier from debian should work, but got an {%v}", err)
	}

	if id != expectedRes {
		t.Errorf("Expected {%v}, got {%v}", expectedRes, id)
	}
}

func TestDebianGetHostIdentifierCachedShouldWork(t *testing.T) {
	expectedRes := "expected result"
	c := NewDebianPlatform().(*Debian)
	c.hostIdentifier = &HostIdentifier{ID: expectedRes}
	c.execCommander = &fake.FakeCommander{
		Err: errors.New("got an error"),
	}

	id, err := c.GetHostIdentifier(context.Background())
	if err != nil {
		t.Errorf("get host identifier from debian should work, but got an {%v}", err)
	}

	if id != expectedRes {
		t.Errorf("Expected {%v}, got {%v}", expectedRes, id)
	}
}

func TestDebianGetHostIdentifierShouldFail(t *testing.T) {
	c := NewDebianPlatform().(*Debian)
	c.execCommander = &fake.FakeCommander{
		Err: errors.New("got an error"),
	}

	_, err := c.GetHostIdentifier(context.Background())
	if err == nil {
		t.Errorf("get host identifier from debian should fail")
	}
}
