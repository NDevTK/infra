// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"

	"infra/unifiedfleet/app/controller"
)

// getBotConfigs reads the bot configs.
//
// Bot configs are read from the files specified in the UFS config
// to get Ownership and security data needed by Puppet ENC
// which is stored in the datastore per bot(machine, vm etc).
func getBotConfigs(ctx context.Context) (retErr error) {
	// TODO - Imported bot configs should be saved to DataStore
	retErr = controller.ImportBotConfigs(ctx)
	return retErr
}
