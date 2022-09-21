// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/logging"
	"google.golang.org/protobuf/encoding/prototext"

	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/external"

	configpb "go.chromium.org/luci/swarming/proto/config"
)

func ImportENCBotConfig(ctx context.Context) error {
	es, err := external.GetServerInterface(ctx)
	if err != nil {
		return err
	}
	ownershipConfig := config.Get(ctx).GetOwnershipConfig()
	gitClient, err := es.NewGitInterface(ctx, ownershipConfig.GetGitilesHost(), ownershipConfig.GetProject(), ownershipConfig.GetBranch())
	if err != nil {
		logging.Errorf(ctx, "Got Error for git client : %s", err.Error())
		return fmt.Errorf("failed to initialize connection to Gitiles while importing enc bot configs")
	}

	for _, cfg := range ownershipConfig.GetEncConfig() {
		logging.Debugf(ctx, "########### Parse %s ###########", cfg.GetName())
		conf, err := gitClient.GetFile(ctx, cfg.GetRemotePath())
		if err != nil {
			return err
		}
		content := &configpb.BotsCfg{}
		err = prototext.Unmarshal([]byte(conf), content)
		if err != nil {
			return err
		}
	}
	return nil
}
