// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package image

import (
	"context"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// SelectBtpeerImageByUUID returns the first image in config.Images with a matching
// image UUID. Returns non-nil error if no matching images found.
func SelectBtpeerImageByUUID(ctx context.Context, config *labapi.RaspiosCrosBtpeerImageConfig, imageUUID string) (*labapi.RaspiosCrosBtpeerImageConfig_OSImage, error) {
	if len(config.GetImages()) == 0 {
		return nil, errors.Reason("select image by UUID: no images defined in config").Err()
	}
	var selectedImageConfig *labapi.RaspiosCrosBtpeerImageConfig_OSImage
	for _, imageConfig := range config.GetImages() {
		if imageConfig.GetUuid() == imageUUID {
			selectedImageConfig = imageConfig
			break
		}
	}
	if selectedImageConfig == nil {
		return nil, errors.Reason("select image by UUID: no image defined in config with UUID %q", imageUUID).Err()
	}
	imageConfigJSON, err := protojson.Marshal(selectedImageConfig)
	if err != nil {
		return nil, errors.Annotate(err, "select image by UUID: failed to marshal selected image to JSON").Err()
	}
	logging.Infof(ctx, "Selected btpeer image: %s", string(imageConfigJSON))
	return selectedImageConfig, nil
}

// SelectCurrentBtpeerImage returns the current image, as defined by
// config.CurrentImageUuid. Returns a non-nil error if no matching images were
// found or if config.CurrentImageUuid is empty.
func SelectCurrentBtpeerImage(ctx context.Context, config *labapi.RaspiosCrosBtpeerImageConfig) (*labapi.RaspiosCrosBtpeerImageConfig_OSImage, error) {
	if config.GetCurrentImageUuid() == "" {
		return nil, errors.Reason("select current btpeer image: no current image set").Err()
	}
	logging.Infof(ctx, "Selecting current btpeer image with UUID %q", config.GetCurrentImageUuid())
	imageConfig, err := SelectBtpeerImageByUUID(ctx, config, config.GetCurrentImageUuid())
	if err != nil {
		return nil, errors.Annotate(err, "select current btpeer image").Err()
	}
	return imageConfig, nil
}

// SelectNextBtpeerImage returns the next image, as defined by
// config.NextImageUuid. If config.NextImageUuid is empty, the current image is
// selected instead via SelectCurrentBtpeerImage. Returns a non-nil error if no
// matching images were found.
func SelectNextBtpeerImage(ctx context.Context, config *labapi.RaspiosCrosBtpeerImageConfig) (*labapi.RaspiosCrosBtpeerImageConfig_OSImage, error) {
	if config.GetNextImageUuid() == "" {
		// Fallback to current.
		logging.Infof(ctx, "Selecting current btpeer image instead of next, as no next image is defined")
		imageConfig, err := SelectCurrentBtpeerImage(ctx, config)
		if err != nil {
			return nil, errors.Annotate(err, "select current btpeer image: no next image set and failed to fallback to current image").Err()
		}
		return imageConfig, nil
	}
	logging.Infof(ctx, "Selecting next btpeer image with UUID %q", config.GetNextImageUuid())
	imageConfig, err := SelectBtpeerImageByUUID(ctx, config, config.GetNextImageUuid())
	if err != nil {
		return nil, errors.Annotate(err, "select next btpeer image").Err()
	}
	return imageConfig, nil
}

// SelectBtpeerImageForDut returns the next image if the dutHostname is in the
// config.NextImageVerificationDutPool (see SelectNextBtpeerImage), or the
// current image if not (see SelectCurrentBtpeerImage).
func SelectBtpeerImageForDut(ctx context.Context, config *labapi.RaspiosCrosBtpeerImageConfig, dutHostname string) (*labapi.RaspiosCrosBtpeerImageConfig_OSImage, error) {
	useNext := false
	for _, nextDutHostname := range config.GetNextImageVerificationDutPool() {
		if strings.EqualFold(nextDutHostname, dutHostname) {
			useNext = true
			break
		}
	}
	var imageConfig *labapi.RaspiosCrosBtpeerImageConfig_OSImage
	var err error
	if useNext {
		logging.Infof(ctx, "Selecting next btpeer image for dut %q, as dut is in the next image verification pool", dutHostname)
		imageConfig, err = SelectNextBtpeerImage(ctx, config)
	} else {
		logging.Infof(ctx, "Selecting current btpeer image for dut %q, as dut is not in the next image verification pool", dutHostname)
		imageConfig, err = SelectCurrentBtpeerImage(ctx, config)
	}
	if err != nil {
		return nil, errors.Annotate(err, "select btpeer image for dut %q", dutHostname).Err()
	}
	return imageConfig, nil
}
