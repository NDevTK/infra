// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cloudsdk

import (
	"context"
	"reflect"
	"testing"
	"time"

	"infra/libs/vmlab/api"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	"go.chromium.org/luci/common/errors"
)

type mockImageClient struct {
	getFunc    func() (*computepb.Image, error)
	importFunc func() (*compute.Operation, error)
}

func (m *mockImageClient) Get(ctx context.Context, req *computepb.GetImageRequest, opts ...gax.CallOption) (*computepb.Image, error) {
	return m.getFunc()
}

func (m *mockImageClient) Insert(context.Context, *computepb.InsertImageRequest, ...gax.CallOption) (*compute.Operation, error) {
	return m.importFunc()
}

func TestDescribeImageError(t *testing.T) {
	imageApi := &cloudsdkImageApi{}
	client := &mockImageClient{getFunc: func() (*computepb.Image, error) {
		return nil, errors.New("404")
	}}
	gceImage := &api.GceImage{
		Name:    "my-image",
		Project: "my-project",
	}

	_, err := imageApi.describeImage(client, gceImage)

	if err == nil {
		t.Errorf("describeImage() expected error, got nil")
	}
	if gceImage.Status != api.GceImage_NOT_FOUND {
		t.Errorf("describeImage() expected status api.GceImage_NOT_FOUND, got %v", gceImage.Status)
	}
}

func TestDescribeImageReady(t *testing.T) {
	imageApi := &cloudsdkImageApi{}
	client := &mockImageClient{getFunc: func() (*computepb.Image, error) {
		status := "READY"
		i := &computepb.Image{
			Status: &status,
		}
		return i, nil
	}}
	gceImage := &api.GceImage{
		Name:    "my-image",
		Project: "my-project",
	}

	gceImage, err := imageApi.describeImage(client, gceImage)

	if err != nil {
		t.Errorf("describeImage() expected nil error, got %v", err)
	}
	if gceImage.Status != api.GceImage_READY {
		t.Errorf("describeImage() expected status api.GceImage_READY, got %v", gceImage.Status)
	}
}

func TestDescribeImagePending(t *testing.T) {
	imageApi := &cloudsdkImageApi{}
	client := &mockImageClient{getFunc: func() (*computepb.Image, error) {
		status := "PENDING"
		i := &computepb.Image{
			Status: &status,
		}
		return i, nil
	}}
	gceImage := &api.GceImage{
		Name:    "my-image",
		Project: "my-project",
	}

	gceImage, err := imageApi.describeImage(client, gceImage)

	if err != nil {
		t.Errorf("describeImage() expected nil error, got %v", err)
	}
	if gceImage.Status != api.GceImage_PENDING {
		t.Errorf("describeImage() expected status api.GceImage_PENDING, got %v", gceImage.Status)
	}
}

func TestHandleReady(t *testing.T) {
	imageApi := &cloudsdkImageApi{}
	client := &mockImageClient{getFunc: func() (*computepb.Image, error) {
		status := "READY"
		i := &computepb.Image{
			Status: &status,
		}
		return i, nil
	}}
	gceImage := &api.GceImage{
		Name:    "my-image",
		Project: "my-project",
	}

	op, gceImage, err := imageApi.handle(client, nil, gceImage)

	if op != nil {
		t.Errorf("handle() expected nil operation, got %v", op)
	}
	if err != nil {
		t.Errorf("handle() expected nil error, got %v", err)
	}
	if gceImage.Status != api.GceImage_READY {
		t.Errorf("handle() expected status api.GceImage_READY, got %v", gceImage.Status)
	}
}

func TestHandlePending(t *testing.T) {
	imageApi := &cloudsdkImageApi{}
	client := &mockImageClient{
		getFunc: func() (*computepb.Image, error) {
			status := "PENDING"
			i := &computepb.Image{
				Status: &status,
			}
			return i, nil
		},
	}
	gceImage := &api.GceImage{
		Name:    "my-image",
		Project: "my-project",
	}

	_, gceImage, err := imageApi.handle(client, nil, gceImage)

	if err != nil {
		t.Errorf("handle() expected nil error, got %v", err)
	}
	if gceImage.Status != api.GceImage_PENDING {
		t.Errorf("handle() expected status api.GceImage_PENDING, got %v", gceImage.Status)
	}
}

func TestHandleImportImage(t *testing.T) {
	imageApi := &cloudsdkImageApi{}
	expected := &compute.Operation{}
	client := &mockImageClient{
		getFunc: func() (*computepb.Image, error) {
			return nil, errors.New("404")
		},
		importFunc: func() (*compute.Operation, error) {
			return expected, nil
		},
	}
	gceImage := &api.GceImage{
		Name:    "my-image",
		Project: "my-project",
	}

	op, gceImage, err := imageApi.handle(client, nil, gceImage)

	if op != expected {
		t.Errorf("handle() expected operation %v, got %v", expected, op)
	}
	if err != nil {
		t.Errorf("handle() expected nil error, got %v", err)
	}
	if gceImage.Status != api.GceImage_PENDING {
		t.Errorf("handle() expected status api.GceImage_PENDING, got %v", gceImage.Status)
	}
}

func TestHandleImportError(t *testing.T) {
	imageApi := &cloudsdkImageApi{}
	client := &mockImageClient{
		getFunc: func() (*computepb.Image, error) {
			return nil, errors.New("404")
		},
		importFunc: func() (*compute.Operation, error) {
			return nil, errors.New("Outage")
		},
	}
	gceImage := &api.GceImage{
		Name:    "my-image",
		Project: "my-project",
	}

	_, gceImage, err := imageApi.handle(client, nil, gceImage)

	if err == nil {
		t.Errorf("handle() expected error, got nil")
	}
	if gceImage.Status != api.GceImage_NOT_FOUND {
		t.Errorf("handle() expected status api.GceImage_NOT_FOUND, got %v", gceImage.Status)
	}
}

func TestParseBuildPathInvalid(t *testing.T) {
	buildPath := "R108-15164.0.0-71927-8801111609984657185"
	_, err := parseBuildPath(buildPath)
	if err == nil {
		t.Errorf("parseBuildPath() expected error, got nil")
	}
}

func TestParseBuildPathAndConvertNameCq(t *testing.T) {
	buildPath := "betty-arc-r-cq/R108-15164.0.1-71927-8801111609984657185"
	info, err := parseBuildPath(buildPath)
	if err != nil {
		t.Fatalf("parseBuildPath() error: %v", err)
	}

	expectedInfo := buildInfo{
		buildType:    "cq",
		board:        "betty-arc-r",
		milestone:    "108",
		majorVersion: "15164",
		minorVersion: "0",
		patchNumber:  "1",
		snapshot:     "71927",
		buildNumber:  "8801111609984657185",
	}
	if !reflect.DeepEqual(*info, expectedInfo) {
		t.Errorf("Expected build info: %s, but is actual: %s", expectedInfo, info)
	}

	actualName := getImageName(*info)
	expectedName := "betty-arc-r-108-15164-0-1-71927-8801111609984657185-cq"
	if expectedName != actualName {
		t.Errorf("Expected image name: %s, but is actual: %s", expectedName, actualName)
	}
}

func TestParseBuildPathAndConvertNamePostsubmit(t *testing.T) {
	buildPath := "betty-pi-arc-postsubmit/R113-15376.0.0-79071-8787141177342104481"
	info, err := parseBuildPath(buildPath)
	if err != nil {
		t.Fatalf("parseBuildPath() error: %v", err)
	}

	expectedInfo := buildInfo{
		buildType:    "postsubmit",
		board:        "betty-pi-arc",
		milestone:    "113",
		majorVersion: "15376",
		minorVersion: "0",
		patchNumber:  "0",
		snapshot:     "79071",
		buildNumber:  "8787141177342104481",
	}
	if !reflect.DeepEqual(*info, expectedInfo) {
		t.Errorf("Expected build info: %s, but is actual: %s", expectedInfo, info)
	}

	actualName := getImageName(*info)
	expectedName := "betty-pi-arc-113-15376-0-0-79071-8787141177342104481-postsubmit"
	if expectedName != actualName {
		t.Errorf("Expected image name: %s, but is actual: %s", expectedName, actualName)
	}
}

func TestParseBuildPathAndConvertNameRelease(t *testing.T) {
	buildPath := "betty-arc-r-release/R108-15178.0.0"
	info, err := parseBuildPath(buildPath)
	if err != nil {
		t.Fatalf("parseBuildPath() error: %v", err)
	}

	expectedInfo := buildInfo{
		buildType:    "release",
		board:        "betty-arc-r",
		milestone:    "108",
		majorVersion: "15178",
		minorVersion: "0",
		patchNumber:  "0",
		snapshot:     "",
		buildNumber:  "",
	}
	if !reflect.DeepEqual(*info, expectedInfo) {
		t.Errorf("Expected build info: %s, but is actual: %s", expectedInfo, info)
	}

	actualName := getImageName(*info)
	expectedName := "betty-arc-r-108-15178-0-0---release"
	if expectedName != actualName {
		t.Errorf("Expected image name: %s, but is actual: %s", expectedName, actualName)
	}
}

func TestParseBuildPathAndConvertNameLong(t *testing.T) {
	buildPath := "betty-arc-r-postsubmit/R113-15376.99.99-79071-8787141177342104481"
	info, err := parseBuildPath(buildPath)
	if err != nil {
		t.Fatalf("parseBuildPath() error: %v", err)
	}

	expectedInfo := buildInfo{
		buildType:    "postsubmit",
		board:        "betty-arc-r",
		milestone:    "113",
		majorVersion: "15376",
		minorVersion: "99",
		patchNumber:  "99",
		snapshot:     "79071",
		buildNumber:  "8787141177342104481",
	}
	if !reflect.DeepEqual(*info, expectedInfo) {
		t.Errorf("Expected build info: %s, but is actual: %s", expectedInfo, info)
	}

	actualName := getImageName(*info)
	expectedName := "betty-arc-r-113-15376-99-99-79071-8787141177342104481-postsubmi"
	if expectedName != actualName {
		t.Errorf("Expected image name: %s, but is actual: %s", expectedName, actualName)
	}
}

func TestGetImageLabels(t *testing.T) {
	actual := getImageLabels(&buildInfo{
		buildType: "cq",
		board:     "betty-arc-r",
		milestone: "100",
	})
	expected := map[string]string{
		"build-type": "cq",
		"board":      "betty-arc-r",
		"milestone":  "100",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected labels: %v, but is actual: %v", expected, actual)
	}
}

func TestPollQuitOnError(t *testing.T) {
	expected := 2
	count := 1
	ctx, cancel := context.WithTimeout(context.Background(), 310*time.Millisecond)
	defer cancel()
	f := func(ctx context.Context) (bool, error) {
		count++
		if count == 2 {
			return false, errors.New("error on 2")
		}
		return false, nil
	}
	interval := 100 * time.Millisecond

	err := poll(ctx, f, interval)
	actual := count

	if expected != actual {
		t.Errorf("count expected: %v, actual: %v", expected, actual)
	}
	if err == nil {
		t.Errorf("poll() expected error, got nil")
	}
}

func TestPollQuitOnSuccess(t *testing.T) {
	expected := 3
	count := 1
	ctx, cancel := context.WithTimeout(context.Background(), 310*time.Millisecond)
	defer cancel()
	f := func(ctx context.Context) (bool, error) {
		count++
		if count == 3 {
			return true, nil
		}
		return false, nil
	}
	interval := 100 * time.Millisecond

	_ = poll(ctx, f, interval)
	actual := count

	if expected != actual {
		t.Errorf("count expected: %v, actual: %v", expected, actual)
	}
}

func TestPollNoDeadlineError(t *testing.T) {
	ctx := context.Background()
	f := func(ctx context.Context) (bool, error) {
		return false, nil
	}
	interval := time.Duration(1)

	err := poll(ctx, f, interval)

	if err == nil {
		t.Errorf("poll() expected error, got nil")
	}
}
