// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

const (
	// moSysSkuCmd will retrieve the SKU label of the DUT.
	moSysSkuCmd = "mosys platform sku"
)

// updateDeviceSKUExec updates device's SKU label if not present in inventory
// or keep it the same if the args.DUT already has the value for SKU label.
func updateDeviceSKUExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	r := args.NewRunner(args.ResourceName)
	skuLabelOutput, err := r(ctx, moSysSkuCmd)
	if err != nil {
		log.Debug(ctx, "Device sku label not found in the DUT.")
		return errors.Annotate(err, "update device sku label").Err()
	}
	args.DUT.DeviceSku = skuLabelOutput
	return nil
}

// isAudioLoopBackStateWorkingExec checks if the DUT's audio loop back state has already been in the working state.
func isAudioLoopBackStateWorkingExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	if args.DUT.AudioLoopbackState != tlw.AudioLoopbackStateWorking {
		return errors.Reason("audio loop back state working: not working").Err()
	}
	return nil
}

const (
	crasAudioHeadPhone = "HEADPHONE"
	crasAudioMic       = "MIC"
)

// updateAudioLoopBackLabelExec updates the DUT's audio loop back state to the correct state
/// based on the condition as follows:
// if both the Headphone and Mic exists on the DUT, then the state is working.
// For all other cases, state is unspecified.
func updateAudioLoopBackLabelExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	r := args.NewRunner(args.ResourceName)
	// check if the Headphone cras audio type exists on the DUT.
	isAudioHeadPhoneExist, err := crasAudioNodeTypeIsPlugged(ctx, r, crasAudioHeadPhone)
	if err != nil {
		args.DUT.AudioLoopbackState = tlw.AudioLoopbackStateUnspecified
		log.Info(ctx, "Setting DUT's Audio Loopback State to be %s", tlw.AudioLoopbackStateUnspecified)
		return errors.Annotate(err, "update audio loop back label").Err()
	}
	// check if the Mic cras audio type exists on the DUT.
	isAudioMicExist, err := crasAudioNodeTypeIsPlugged(ctx, r, crasAudioMic)
	if err != nil {
		args.DUT.AudioLoopbackState = tlw.AudioLoopbackStateUnspecified
		log.Info(ctx, "Setting DUT's Audio Loopback State to be %s", tlw.AudioLoopbackStateUnspecified)
		return errors.Annotate(err, "update audio loop back label").Err()
	}
	if isAudioHeadPhoneExist && isAudioMicExist {
		args.DUT.AudioLoopbackState = tlw.AudioLoopbackStateWorking
		log.Info(ctx, "Setting DUT's Audio Loopback State to be %s", tlw.AudioLoopbackStateWorking)
		return nil
	}
	args.DUT.AudioLoopbackState = tlw.AudioLoopbackStateUnspecified
	log.Info(ctx, "Setting DUT's Audio Loopback State to be %s", tlw.AudioLoopbackStateUnspecified)
	return nil
}

func init() {
	execs.Register("cros_update_device_sku", updateDeviceSKUExec)
	execs.Register("cros_is_audio_loop_back_state_working", isAudioLoopBackStateWorkingExec)
	execs.Register("cros_update_audio_loop_back_state_label", updateAudioLoopBackLabelExec)
}
