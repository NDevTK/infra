// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"

	"infra/unifiedfleet/app/controller"
)

// getEncBotConfigs reads the bot configs.
//
// Bot configs are read from the files specified in the UFS config
// to get Ownership data needed by Puppet ENC
// which is stored in the datastore per bot(machine, vm etc).
func getEncBotConfigs(ctx context.Context) (retErr error) {
	// TODO - Imported bot configs should be saved to DataStore
	retErr = controller.ImportENCBotConfig(ctx)
	return retErr
}
