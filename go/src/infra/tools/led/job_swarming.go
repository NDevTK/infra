// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/json"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/isolated"
	"go.chromium.org/luci/common/isolatedclient"
	api "go.chromium.org/luci/swarming/proto/api"
)

func extractCmdCwdFromIsolated(ctx context.Context, isoClient *isolatedclient.Client, rootIso isolated.HexDigest) (cmd []string, cwd string, err error) {
	seenIsolateds := map[isolated.HexDigest]struct{}{}
	queue := isolated.HexDigests{rootIso}

	// borrowed from go.chromium.org/luci/client/downloader.
	//
	// It's rather silly that there's no library functionality to do this.
	for len(queue) > 0 {
		iso := queue[0]
		if _, ok := seenIsolateds[iso]; ok {
			err = errors.Reason("loop detected when resolving isolate %q", rootIso).Err()
			return
		}
		seenIsolateds[iso] = struct{}{}

		buf := bytes.Buffer{}
		if err = isoClient.Fetch(ctx, iso, &buf); err != nil {
			err = errors.Annotate(err, "fetching isolated %q", iso).Err()
			return
		}
		isoFile := isolated.Isolated{}
		if err = json.Unmarshal(buf.Bytes(), &isoFile); err != nil {
			err = errors.Annotate(err, "parsing isolated %q", iso).Err()
			return
		}

		if len(isoFile.Command) > 0 {
			cmd = isoFile.Command
			cwd = isoFile.RelativeCwd
			break
		}

		queue = append(isoFile.Includes, queue[1:]...)
	}

	return
}

// ConsolidateIsolateSources will, for Swarming tasks.
//
//   * Extract Cmd/Cwd from the Properties.CasInputs (if set)
//   * Combine the Properties.CasInputs with the UserPayload (if set) and
//     store the combined isolated in Properties.CasInputs.
func (jd *JobDefinition) ConsolidateIsolateSources(ctx context.Context, isoClient *isolatedclient.Client) error {
	if jd.GetSwarming() == nil {
		return nil
	}

	arc := mkArchiver(ctx, isoClient)

	for _, slc := range jd.GetSwarming().GetTask().GetTaskSlices() {
		if slc.Properties == nil {
			slc.Properties = &api.TaskProperties{}
		}
		if slc.Properties.CasInputs == nil {
			slc.Properties.CasInputs = jd.UserPayload
			return nil
		}

		// extract the cmd/cwd from the isolated, if they're set.
		//
		// This is an old feature of swarming/isolated where the isolated file can
		// contain directives for the swarming task.
		cmd, cwd, err := extractCmdCwdFromIsolated(
			ctx, isoClient, isolated.HexDigest(slc.Properties.CasInputs.Digest))
		if err != nil {
			return err
		}
		if len(cmd) > 0 {
			slc.Properties.Command = cmd
			slc.Properties.RelativeCwd = cwd
			// ExtraArgs is allowed to be set only if the Command is coming from the
			// isolated. However, now that we're explicitly setting the Command, we
			// must move ExtraArgs into Command.
			if len(slc.Properties.ExtraArgs) > 0 {
				slc.Properties.Command = append(slc.Properties.Command, slc.Properties.ExtraArgs...)
				slc.Properties.ExtraArgs = nil
			}
		}

		if jd.UserPayload == nil {
			continue
		}

		// TODO(maruel): Confirm the namespace here is compatible with arc's.
		// TODO(iannucci): use full CasTree object instead of just digests.
		h := isolated.GetHash(slc.Properties.CasInputs.Namespace)
		newHash, err := combineIsolateds(ctx, arc, h,
			isolated.HexDigest(jd.UserPayload.Digest),
			isolated.HexDigest(slc.Properties.CasInputs.Digest),
		)
		if err != nil {
			return errors.Annotate(err, "combining isolateds").Err()
		}

		slc.Properties.CasInputs = &api.CASTree{Digest: string(newHash)}
		return nil
	}
	return nil
}
