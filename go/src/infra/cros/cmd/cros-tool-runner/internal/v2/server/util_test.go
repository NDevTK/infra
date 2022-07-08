// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package server

import (
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestFirstLine(t *testing.T) {
	original := "first line\n - second line\n"
	expect := "first line"
	firstLine := utils.firstLine(original)
	if firstLine != expect {
		t.Fatalf("output %s doesn't match expected %s", firstLine, expect)
	}
}

func TestFirstLine_oneLine(t *testing.T) {
	original := "first line\n"
	expect := "first line"
	firstLine := utils.firstLine(original)
	if firstLine != expect {
		t.Fatalf("output %s doesn't match expected %s", firstLine, expect)
	}
}

func TestFirstLine_emptyLine(t *testing.T) {
	original := "\n"
	expect := ""
	firstLine := utils.firstLine(original)
	if firstLine != expect {
		t.Fatalf("output %s doesn't match expected %s", firstLine, expect)
	}
}

func TestFirstLine_empty(t *testing.T) {
	original := ""
	expect := ""
	firstLine := utils.firstLine(original)
	if firstLine != expect {
		t.Fatalf("output %s doesn't match expected %s", firstLine, expect)
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	element := "b"
	contains := utils.contains(slice, element)
	if !contains {
		t.Fatalf("contains is incorrect: %t", contains)
	}
}

func TestContains_false(t *testing.T) {
	slice := []string{"a", "b", "c"}
	element := "e"
	contains := utils.contains(slice, element)
	if contains {
		t.Fatalf("contains is incorrect: %t", contains)
	}
}

func TestContains_emptySlice(t *testing.T) {
	slice := make([]string, 0)
	element := "e"
	contains := utils.contains(slice, element)
	if contains {
		t.Fatalf("contains is incorrect: %t", contains)
	}
}

func TestMapToCode_NotFound(t *testing.T) {
	// docker network inspect a
	errMsg := "Error: No such network: a"
	code := utils.mapToCode(errMsg)
	if code != codes.NotFound {
		t.Fatalf("code is incorrect: %v", code)
	}
}

func TestMapToCode_PermissionDenied(t *testing.T) {
	// docker network create bridge
	errMsg := "Error response from daemon: operation is not permitted on predefined bridge network"
	code := utils.mapToCode(errMsg)
	if code != codes.PermissionDenied {
		t.Fatalf("code is incorrect: %v", code)
	}
}

func TestMapToCode_AlreadyExists(t *testing.T) {
	// docker network create mynet & docker network create mynet
	errMsg := "Error response from daemon: network with name mynet already exists"
	code := utils.mapToCode(errMsg)
	if code != codes.AlreadyExists {
		t.Fatalf("code is incorrect: %v", code)
	}
}

func TestMapToCode_Unknown(t *testing.T) {
	errMsg := "Some error not yet mapped"
	code := utils.mapToCode(errMsg)
	if code != codes.Unknown {
		t.Fatalf("code is incorrect: %v", code)
	}
}

func TestReverse(t *testing.T) {
	s := []string{"a", "b", "c"}
	expect := []string{"c", "b", "a"}
	reversed := utils.reverse(s)
	check(t, reversed, expect)
}

func TestReverse_empty(t *testing.T) {
	s := make([]string, 0)
	expect := make([]string, 0)
	reversed := utils.reverse(s)
	check(t, reversed, expect)
}

func TestReverse_oneElement(t *testing.T) {
	s := []string{"a"}
	expect := []string{"a"}
	reversed := utils.reverse(s)
	check(t, reversed, expect)
}

func check(t *testing.T, actual []string, expect []string) {
	actualStr := strings.Join(actual, ",")
	expectStr := strings.Join(expect, ",")
	if actualStr != expectStr {
		t.Fatalf("Slices do not match expect\nExpect: %v\nActual: %v", expect, actual)
	}
}
