// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"reflect"
	"testing"

	"infra/cros/internal/assert"
)

// TestPrependString tests prependString.
func TestPrependString(t *testing.T) {
	t.Parallel()
	for i, tc := range []struct {
		newElem        string
		arr            []string
		expectedResult []string
	}{
		{"foo", []string{}, []string{"foo"}},
		{"foo", []string{"bar", "baz"}, []string{"foo", "bar", "baz"}},
	} {
		actualResult := prependString(tc.newElem, tc.arr)
		if !reflect.DeepEqual(actualResult, tc.expectedResult) {
			t.Errorf("#%d: prependString(%s, %s) returned %s; want %s", i, tc.newElem, tc.arr, actualResult, tc.expectedResult)
		}
	}
}

// TestSeparateBucketFromBuilder tests separateBucketFromBuilder.
func TestSeparateBucketFromBuilder(t *testing.T) {
	t.Parallel()
	for i, tc := range []struct {
		fullBuilderName string
		expectedBucket  string
		expectedBuilder string
		expectError     bool
	}{
		{"chromeos/release/release-main-orchestrator", "chromeos/release", "release-main-orchestrator", false},
		{"too/short", "", "", true},
		{"too/long/by/far", "", "", true},
	} {
		bucket, builder, err := separateBucketFromBuilder(tc.fullBuilderName)
		if err != nil && !tc.expectError {
			t.Errorf("#%d: separateBucketFromBuilder(%s) returned an error; want no error. Returned error: %+v", i, tc.fullBuilderName, err)
		}
		if err == nil && tc.expectError {
			t.Errorf("#%d: separateBucketFromBuilder(%s) returned no error; want error", i, tc.fullBuilderName)
		}
		if bucket != tc.expectedBucket {
			t.Errorf("#%d: separateBucketFromBuilder(%s) returned unexpected bucket: got %s; want %s", i, tc.fullBuilderName, bucket, tc.expectedBucket)
		}
		if builder != tc.expectedBuilder {
			t.Errorf("#%d: separateBucketFromBuilder(%s) returned unexpected builder: got %s; want %s", i, tc.fullBuilderName, builder, tc.expectedBuilder)
		}
	}
}

// TestParseEmailFromAuthInfo tests parseEmailFromAuthInfo.
func TestParseEmailFromAuthInfo(t *testing.T) {
	t.Parallel()

	email, err := parseEmailFromAuthInfo("Logged in as sundar@google.com.\n\nfoo...")
	assert.NilError(t, err)
	assert.StringsEqual(t, email, "sundar@google.com")

	email, err = parseEmailFromAuthInfo("Logged in as sundar@subdomain.google.com.\n\nfoo...")
	assert.NilError(t, err)
	assert.StringsEqual(t, email, "sundar@subdomain.google.com")

	email, err = parseEmailFromAuthInfo("Logged in as sundar.pichai@google.com.\n\nfoo...")
	assert.NilError(t, err)
	assert.StringsEqual(t, email, "sundar.pichai@google.com")

	email, err = parseEmailFromAuthInfo("Logged in as sundar+spam@google.com.\n\nfoo...")
	assert.NilError(t, err)
	assert.StringsEqual(t, email, "sundar+spam@google.com")

	_, err = parseEmailFromAuthInfo("\n\nfoo\nLogged in as sundar@google.com.\n\nfoo...")
	assert.NonNilError(t, err)

	_, err = parseEmailFromAuthInfo("Logged in as sundar!!\n\nfoo...")
	assert.NonNilError(t, err)

	_, err = parseEmailFromAuthInfo("Logged in as sundar@.\n\nfoo...")
	assert.NonNilError(t, err)
}

// TestPatchListToBBAddArgs tests patchListToBBAddArgs
func TestPatchListToBBAddArgs(t *testing.T) {
	t.Parallel()

	patchSets := []string{"crrev.com/c/1234567"}
	expectedBBAddArgs := []string{"-cl", "crrev.com/c/1234567"}
	bbAddArgs := patchListToBBAddArgs(patchSets)
	assert.StringArrsEqual(t, bbAddArgs, expectedBBAddArgs)

	patchSets = []string{"crrev.com/c/1234567", "crrev.com/c/8675309"}
	expectedBBAddArgs = []string{"-cl", "crrev.com/c/1234567", "-cl", "crrev.com/c/8675309"}
	bbAddArgs = patchListToBBAddArgs(patchSets)
	assert.StringArrsEqual(t, bbAddArgs, expectedBBAddArgs)

	patchSets = []string{}
	expectedBBAddArgs = []string{}
	bbAddArgs = patchListToBBAddArgs(patchSets)
	assert.StringArrsEqual(t, bbAddArgs, expectedBBAddArgs)
}
