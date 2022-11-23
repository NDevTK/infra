// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package experimental

import (
	"context"
	"fmt"
	"strconv"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cmd/shivas/site"
	"infra/cmdsupport/cmdlib"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// DumpNlyteCmd dumps updated entries from Nlyte to UFS.
var DumpNlyteCmd = &subcommands.Command{
	UsageLine: "nlyte ...",
	ShortDesc: "Dump nlyte updates",
	LongDesc: `Dump nlyte updates from the provided json file.
Example:
shivas nlyte -f testing.json`,
	CommandRun: func() subcommands.CommandRun {
		c := &dumpNlyte{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.updatedEntryFile, "f", "", "Path to a file containing AssetAndHosts specification in JSON format.")

		c.outputFlags.Register(&c.Flags)
		return c
	},
}

type dumpNlyte struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	updatedEntryFile string

	outputFlags site.OutputFlags
}

func (c *dumpNlyte) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *dumpNlyte) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return nil
}

// parseRecord parses the input record and update the input Asset with parsed info.
func (c *dumpNlyte) parseRecord(ctx context.Context, record *ufspb.AssetAndHostInfo, asset *ufspb.Asset) error {
	if record.GetAssetName() == "" {
		return fmt.Errorf("Failed to parse record: Empty AssetName\n")
	}

	// TODO: (b/260134576) Use field masks to specify the fields to update as that is a better way to update.
	// Add more field masks as we increase the amount of tracked field in Nlyte.
	asset.Location.Position = strconv.Itoa(int(record.GetCabinetUNumber()))
	// TODO: (b/260134131) Rack implementation right now is just a place holder because we don't have row and col info in Nlyte yet.
	asset.Location.Rack = fmt.Sprintf("rack-%s", strconv.Itoa(int(record.GetCabinetAssetId())))
	asset.Location.RackId = record.GetCabinetAssetId()
	asset.Location.LabId = record.GetLocationGroup().GetLocationGroupId()
	asset.Location.FullLocationName = record.GetLocationGroup().GetFullLocationName()

	for _, customField := range record.GetAssetInfo().GetCustomFields() {
		if customField.GetFieldName() == "Asset Type" {
			if val, ok := ufspb.AssetType_value[customField.GetFieldStringValue()]; ok {
				asset.Type = ufspb.AssetType(val)
				continue
			}
			return fmt.Errorf("Failed to parse record: Invalid asset type: %s\n", customField.GetFieldStringValue())
		} else if customField.GetFieldName() == "Zone" {
			if val, ok := ufspb.Zone_value[customField.GetFieldStringValue()]; ok {
				asset.Location.Zone = ufspb.Zone(val)
				asset.Realm = ufsUtil.ToUFSRealm(ufspb.Zone(val).String())
				continue
			}
			return fmt.Errorf("Failed to parse record: Invalid zone type: %s\n", customField.GetFieldStringValue())
		}
	}

	return nil
}

// addAssetToUFS attempts to add given asset to UFS. Returns updated asset and error if any
func (c *dumpNlyte) addAssetToUFS(ctx context.Context, ic ufsAPI.FleetClient, req *ufsAPI.CreateAssetRequest) (*ufspb.Asset, error) {
	if req.Asset == nil {
		return nil, cmdlib.NewQuietUsageError(c.Flags, "Failed to add asset: Invalid input, Missing asset to add")
	}
	if req.Asset.Location == nil {
		return nil, cmdlib.NewQuietUsageError(c.Flags, "Failed to add asset %s: Invalid input, Missing any location information", req.Asset.GetName())
	}
	if req.Asset.Location.Rack == "" {
		return nil, cmdlib.NewQuietUsageError(c.Flags, "Failed to add asset %s: Invalid input, Missing rack", req.Asset.GetName())
	}
	if req.Asset.Location.Zone == ufspb.Zone_ZONE_UNSPECIFIED {
		return nil, cmdlib.NewQuietUsageError(c.Flags, "Failed to add asset %s: Invalid zone", req.Asset.GetName())
	}
	ufsAsset, err := ic.CreateAsset(ctx, req)
	if ufsAsset != nil {
		// Remove the prefix from the asset returned by UFS
		ufsAsset.Name = ufsUtil.RemovePrefix(ufsAsset.Name)
	}
	return ufsAsset, err
}
