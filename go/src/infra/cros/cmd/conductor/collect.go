// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	bbpb "go.chromium.org/luci/buildbucket/proto"
)

func (c *collectRun) Collect(config *pb.CollectConfig) error {
	watchSet := c.bbids

	for len(watchSet) > 0 {
		c.LogOut("Sleeping for %d seconds", c.pollingIntervalSeconds)
		time.Sleep(time.Duration(c.pollingIntervalSeconds) * time.Second)

		sort.Strings(watchSet)
		c.LogOut("Waiting for %s", strings.Join(watchSet, ","))
		builds, err := c.bbClient.GetBuilds(context.Background(), watchSet)
		if err != nil {
			return err
		}

		watchSetMap := map[string]bool{}
		for _, bbid := range watchSet {
			watchSetMap[bbid] = true
		}

		newWatchSet := []string{}
		for _, build := range builds {
			if _, ok := watchSetMap[fmt.Sprintf("%d", build.GetId())]; !ok {
				continue
			}
			if (int(build.GetStatus()) & int(bbpb.Status_ENDED_MASK)) == 0 {
				newWatchSet = append(newWatchSet, fmt.Sprintf("%d", build.GetId()))
			} else {
				c.LogOut("Build %d finished with status %s", build.GetId(), build.GetStatus())
			}
		}
		watchSet = newWatchSet
	}

	return nil
}
