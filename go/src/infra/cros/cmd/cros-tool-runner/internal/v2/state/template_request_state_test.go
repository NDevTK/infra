// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package state

import (
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestAdd_exist(t *testing.T) {
	state := newTemplateRequestState()
	state.Add("1", &api.StartTemplatedContainerRequest{Name: "name1"})
	exist := state.Exist("1")
	if !exist {
		t.Fatalf("exist should be true")
	}
}

func TestAdd_notExist(t *testing.T) {
	state := newTemplateRequestState()
	state.Add("1", &api.StartTemplatedContainerRequest{Name: "name1"})
	exist := state.Exist("2")
	if exist {
		t.Fatalf("exist should be false")
	}
}

func TestGet_exist(t *testing.T) {
	state := newTemplateRequestState()
	state.Add("1", &api.StartTemplatedContainerRequest{Name: "name1"})
	request := state.Get("1")
	if request == nil {
		t.Fatalf("request should be fetched")
	}
	if request.Name != "name1" {
		t.Fatalf("request name should be name1 instead of %s", request.Name)
	}
}

func TestGet_notExist(t *testing.T) {
	state := newTemplateRequestState()
	state.Add("1", &api.StartTemplatedContainerRequest{Name: "name1"})
	request := state.Get("2")
	if request != nil {
		t.Fatalf("request should be nil")
	}
}
