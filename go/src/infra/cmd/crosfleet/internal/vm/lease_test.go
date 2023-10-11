// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vm

import (
	"errors"
	"flag"
	"fmt"
	"testing"

	"google.golang.org/protobuf/types/known/timestamppb"

	vmapi "infra/libs/vmlab/api"
)

type mockImageApi struct {
	listImagesFunc func(filter string) ([]*vmapi.GceImage, error)
}

func (m mockImageApi) ListImages(filter string) ([]*vmapi.GceImage, error) {
	return m.listImagesFunc(filter)
}

func (m mockImageApi) GetImage(builderPath string, wait bool) (*vmapi.GceImage, error) {
	return nil, errors.New("not implemented")
}

func (m mockImageApi) DeleteImage(imageName string, wait bool) error {
	return errors.New("not implemented")
}

func TestGetLatestImage(t *testing.T) {
	expectedFilter := "(labels.build-type:release AND labels.board:betty-arc-r)"
	expectedName := "image-latest"
	iapi := mockImageApi{
		listImagesFunc: func(filter string) ([]*vmapi.GceImage, error) {
			if filter != expectedFilter {
				return nil, fmt.Errorf("expected filter to be %s, but is %s", expectedFilter, filter)
			}
			return []*vmapi.GceImage{
				{
					Name:        "image-1",
					TimeCreated: &timestamppb.Timestamp{Seconds: 1},
				},
				{
					Name:        expectedName,
					TimeCreated: &timestamppb.Timestamp{Seconds: 3},
				},
				{
					Name:        "image-2",
					TimeCreated: &timestamppb.Timestamp{Seconds: 2},
				},
			}, nil
		},
	}
	actualName, err := getLatestImage(iapi, "betty-arc-r")

	if err != nil {
		t.Error(err)
	}

	if actualName != expectedName {
		t.Errorf("expected image name to be %s, but is %s", expectedName, actualName)
	}
}

var validLeaseFlags = []leaseFlags{
	{
		durationMins: 60,
		board:        "betty-arc-r",
		build:        "",
	},
	{
		durationMins: maxLeaseLengthMinutes,
		board:        "",
		build:        "betty-arc-r-release/R119-15626.0.0",
	},
}

func TestValidateValidFlags(t *testing.T) {
	for _, f := range validLeaseFlags {
		if err := f.validate(&flag.FlagSet{}); err != nil {
			t.Errorf("Expected flags %v to pass, but got error %v", f, err)
		}
	}
}

var invalidLeaseFlags = []leaseFlags{
	{
		durationMins: 60,
		board:        "",
		build:        "",
	},
	{
		durationMins: 60,
		board:        "betty-arc-r",
		build:        "betty-arc-r-release/R119-15626.0.0",
	},
	{
		durationMins: maxLeaseLengthMinutes + 1,
		board:        "betty-arc-r",
		build:        "",
	},
	{
		durationMins: 0,
		board:        "betty-arc-r",
		build:        "",
	},
}

func TestValidateInvalidFlags(t *testing.T) {
	for _, f := range invalidLeaseFlags {
		if err := f.validate(&flag.FlagSet{}); err == nil {
			t.Errorf("Expected flags %v to fail, but got nil error", f)
		}
	}
}
