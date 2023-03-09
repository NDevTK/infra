// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	invlab "go.chromium.org/chromiumos/infra/proto/go/lab"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"

	invapibq "infra/appengine/cros/lab_inventory/api/bigquery"
	iv2ds "infra/cros/lab_inventory/datastore"
	iv2pr "infra/libs/fleet/protos"
	iv2pr2 "infra/libs/fleet/protos/go"
	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/controller"
	"infra/unifiedfleet/app/model/registration"
	"infra/unifiedfleet/app/util"
)

var macRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:\.-]){5}([0-9A-Fa-f]{2})$`)

// List of fields to be ignored when comparing a machine object to another.
// Field names here should reflect *.proto not generated *.pb.go
var machineCmpIgnoreFields = []protoreflect.Name{
	protoreflect.Name("update_time"),
}

// List of fields to be ignored when comparing a Asset object to another.
// Field names here should reflect *.proto not generated *.pb.go
var assetCmpIgnoreFields = []protoreflect.Name{
	protoreflect.Name("update_time"),
	// Don't care about info, We can update it from HaRT directly
	protoreflect.Name("info"),
}

func checkRackExists(ctx context.Context, rack string) error {
	// It's possible that an asset's rack is empty because
	// a. we cannot parse rack from its hostname, e.g. chromeos1-...jetstream-host5
	// b. the asset is not scanned/doesn't exist in HaRT
	if rack == "" {
		return nil
	}
	return controller.ResourceExist(ctx, []*controller.Resource{controller.GetRackResource(rack)}, nil)
}

func registerRacksForAsset(ctx context.Context, asset *ufspb.Asset) error {
	l := asset.GetLocation()
	rack := &ufspb.Rack{
		Name: l.GetRack(),
		Location: &ufspb.Location{
			Aisle:       l.GetAisle(),
			Row:         l.GetRow(),
			Rack:        l.GetRack(),
			RackNumber:  l.GetRackNumber(),
			BarcodeName: l.GetRack(),
			Zone:        l.GetZone(),
		},
		Description:   "Added from IV2 by SyncAssetsFromIV2",
		ResourceState: ufspb.State_STATE_SERVING,
		Realm:         util.ToUFSRealm(l.GetZone().String()),
	}
	logging.Infof(ctx, "Add rack: %v", rack)
	_, err := controller.RackRegistration(ctx, rack)
	return err
}

// SyncMachinesFromAssets updates machines table from assets table
//
// Checks all the DUT and Labstation assets and creates/updates machines if required.
func SyncMachinesFromAssets(ctx context.Context) error {
	// In UFS write to 'os' namespace
	var err error
	ctx, err = util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	if err != nil {
		return err
	}
	logging.Infof(ctx, "SyncMachinesFromAssets")
	assets, err := registration.GetAllAssets(ctx)
	if err != nil {
		return err
	}
	for _, asset := range assets {
		// Store DUTs and Labstations as machines
		if asset.GetType() == ufspb.AssetType_DUT || asset.GetType() == ufspb.AssetType_LABSTATION {
			// Create rack when creating machines
			if err := checkRackExists(ctx, asset.GetLocation().GetRack()); err != nil {
				if err := registerRacksForAsset(ctx, asset); err != nil {
					logging.Warningf(ctx, "Unable to create rack %s: %s", asset.GetLocation().GetRack(), err.Error())
					continue
				}
			}
			aMachine := controller.CreateMachineFromAsset(asset)
			if aMachine == nil {
				continue
			}
			ufsMachine, err := controller.GetMachine(ctx, asset.GetName())
			if err != nil && util.IsNotFoundError(err) {
				// Create a new machine
				_, err := controller.MachineRegistration(ctx, aMachine)
				if err != nil {
					logging.Warningf(ctx, "Unable to create machine %v %v", aMachine, err)
				}
			} else if ufsMachine != nil && !Compare(aMachine, ufsMachine) {
				// Serial number, Hwid, Sku of UFS machine is updated by SSW in
				// UpdateDutMeta https://source.corp.google.com/chromium_infra/go/src/infra/unifiedfleet/app/controller/machine.go;l=182
				// Dont rely on Hart for Serial number, Hwid, Sku and
				// macaddress. Copy back original values.
				aMachine.SerialNumber = ufsMachine.GetSerialNumber()
				aMachine.GetChromeosMachine().Hwid = ufsMachine.GetChromeosMachine().GetHwid()
				aMachine.GetChromeosMachine().Sku = ufsMachine.GetChromeosMachine().GetSku()
				aMachine.GetChromeosMachine().MacAddress = ufsMachine.GetChromeosMachine().GetMacAddress()
				_, err := controller.UpdateMachine(ctx, aMachine, nil)
				if err != nil {
					logging.Warningf(ctx, "Failed to update machine %v %v", aMachine, err)
				}
			}
		}
	}
	return nil
}

// GetAllAssets retrieves all the asset data from inventory-V2
func GetAllAssets(ctx context.Context, client *datastore.Client) ([]*iv2pr.ChopsAsset, error) {
	var assetEntities []*iv2ds.AssetEntity

	k, err := client.GetAll(ctx, datastore.NewQuery(iv2ds.AssetEntityName), &assetEntities)
	if err != nil {
		return nil, err
	}
	logging.Debugf(ctx, "Found %v assetEntities", len(assetEntities))

	assets := make([]*iv2pr.ChopsAsset, 0, len(assetEntities))
	for idx, a := range assetEntities {
		// Add key to the asset. GetAll doesn't update keys but
		// returns []keys in order
		a.ID = k[idx].Name
		asset, err := a.ToChopsAsset()
		if err != nil {
			logging.Warningf(ctx, "Unable to parse %v: %v", a.ID, err)
		}
		assets = append(assets, asset)
	}
	return assets, nil
}

// GetAllAssetInfo retrieves all the asset info data from inventory-V2
func GetAllAssetInfo(ctx context.Context, client *datastore.Client) (map[string]*iv2pr2.AssetInfo, error) {
	var assetInfoEntities []*iv2ds.AssetInfoEntity

	_, err := client.GetAll(ctx, datastore.NewQuery(iv2ds.AssetInfoEntityKind), &assetInfoEntities)
	if err != nil {
		return nil, err
	}
	logging.Debugf(ctx, "Found %v assetInfoEntities", len(assetInfoEntities))

	assetInfos := make(map[string]*iv2pr2.AssetInfo, len(assetInfoEntities))
	for _, a := range assetInfoEntities {
		assetInfos[a.Info.GetAssetTag()] = &a.Info
	}
	return assetInfos, nil
}

// GetAssetToHostnameMap gets the asset tag to hostname mapping from
// assets_in_swarming BQ table
func GetAssetToHostnameMap(ctx context.Context, client *bigquery.Client) (map[string]string, error) {
	type mapping struct {
		AssetTag string
		HostName string
	}
	//TODO(anushruth): Get table name, dataset and project from config
	q := client.Query(`
		SELECT a_asset_tag AS AssetTag, s_host_name AS HostName FROM ` +
		"`cros-lab-inventory.inventory.assets_in_swarming`")
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}
	// Read the first mapping as TotalRows is not populated until first
	// call to Next()
	var d mapping
	err = it.Next(&d)
	assetsToHostname := make(map[string]string, int(it.TotalRows))
	assetsToHostname[d.AssetTag] = d.HostName

	for {
		err := it.Next(&d)
		if err == iterator.Done {
			break
		}
		if err != nil {
			logging.Warningf(ctx, "Failed to read a row from BQ: %v", err)
		}
		assetsToHostname[d.AssetTag] = d.HostName
	}
	logging.Debugf(ctx, "Found hostnames for %v devices", len(assetsToHostname))
	return assetsToHostname, nil
}

// Compare does protobuf comparison between both inputs
func Compare(iv2Machine, ufsMachine *ufspb.Machine) bool {
	// Ignoring fields not required for comparison
	opts1 := protocmp.IgnoreFields(iv2Machine, machineCmpIgnoreFields...)
	// See: https://developers.google.com/protocol-buffers/docs/reference/go/faq#deepequal
	opts2 := protocmp.Transform()
	return cmp.Equal(iv2Machine, ufsMachine, opts1, opts2)
}

// Cmp does protobuf comparison between both inputs
func Cmp(iv2Asset, ufsAsset *ufspb.Asset) bool {
	opts1 := protocmp.IgnoreFields(ufsAsset, assetCmpIgnoreFields...)
	opts2 := protocmp.Transform()
	return cmp.Equal(iv2Asset, ufsAsset, opts1, opts2)
}

// createAssetsFromChopsAsset returns Asset proto constructed from ChopsAsset and AssetInfo proto
func createAssetsFromChopsAsset(asset *iv2pr.ChopsAsset, assetinfo *iv2pr2.AssetInfo, hostname string) (*ufspb.Asset, error) {
	a := &ufspb.Asset{
		Name: asset.GetId(),
		Location: &ufspb.Location{
			Aisle:       asset.GetLocation().GetAisle(),
			Row:         asset.GetLocation().GetRow(),
			Shelf:       asset.GetLocation().GetShelf(),
			Position:    asset.GetLocation().GetPosition(),
			BarcodeName: hostname,
		},
	}
	if assetinfo != nil {
		a.Info = &ufspb.AssetInfo{
			AssetTag:           assetinfo.GetAssetTag(),
			SerialNumber:       assetinfo.GetSerialNumber(),
			CostCenter:         assetinfo.GetCostCenter(),
			GoogleCodeName:     assetinfo.GetGoogleCodeName(),
			Model:              assetinfo.GetModel(),
			BuildTarget:        assetinfo.GetBuildTarget(),
			ReferenceBoard:     assetinfo.GetReferenceBoard(),
			EthernetMacAddress: assetinfo.GetEthernetMacAddress(),
			Sku:                assetinfo.GetSku(),
			Phase:              assetinfo.GetPhase(),
		}
	}

	a.Location.Zone = util.LabToZone(asset.GetLocation().GetLab())
	if a.Location.Zone == ufspb.Zone_ZONE_CROS_GOOGLER_DESK && hostname == "" {
		a.Location.BarcodeName = asset.GetLocation().GetLab()
	}
	// Construct rack name as `chromeos[$zone]`-row`$row`-rack`$rack`
	loc := asset.GetLocation()
	var r strings.Builder
	if loc.GetLab() == "" {
		return nil, errors.Reason("Cannot create an asset without zone").Err()
	}
	r.WriteString(loc.GetLab())
	if row := loc.GetRow(); row != "" {
		r.WriteString("-row")
		r.WriteString(row)
	}
	if rack := loc.GetRack(); rack != "" {
		r.WriteString("-rack")
		r.WriteString(rack)
		a.Location.RackNumber = rack
	} else {
		// Avoid setting Rack to zone name, e.g. chromeos2
		r.WriteString("-norack")
	}
	a.Location.Rack = r.String()
	if assetinfo != nil && assetinfo.GetGoogleCodeName() != "" {
		// Convert the model to all lowercase for compatibility with rest of the data
		a.Model = strings.ToLower(assetinfo.GetGoogleCodeName())
	}
	// Device can be one of DUT, Labstation, Servo, etc,.
	if a.Model == "" {
		// Some servos are recorded using their ethernet mac address
		if macRegex.MatchString(a.GetName()) {
			a.Type = ufspb.AssetType_SERVO
		} else {
			a.Type = ufspb.AssetType_UNDEFINED
		}
	} else if strings.Contains(a.Model, "labstation") {
		a.Type = ufspb.AssetType_LABSTATION
	} else if strings.Contains(a.Model, "servo") {
		a.Type = ufspb.AssetType_SERVO
	} else {
		// The asset is a DUT if it has model info and isn't a labstation or servo.
		a.Type = ufspb.AssetType_DUT
	}
	return a, nil
}

// DeviceDataToBQDeviceMsgs converts a sequence of devices data into messages that can be committed to bigquery.
func DeviceDataToBQDeviceMsgs(ctx context.Context, devicesData []*DeviceData) []proto.Message {
	labconfigs := make([]proto.Message, len(devicesData))
	for i, data := range devicesData {
		if data.Device == nil || data.UpdateTime == nil {
			logging.Errorf(ctx, "deviceData Device or UpdateTime is nil")
			continue
		}
		var hostname string
		if data.Device.GetDut() != nil {
			hostname = data.Device.GetDut().GetHostname()
		} else {
			hostname = data.Device.GetLabstation().GetHostname()
		}
		labconfigs[i] = &invapibq.LabInventory{
			Id:          data.Device.GetId().GetValue(),
			Hostname:    hostname,
			Device:      data.Device,
			UpdatedTime: data.UpdateTime,
		}
		fmt.Println(labconfigs[i])
	}
	return labconfigs
}

// DeviceData holds the invV2 Device and updatetime(of MachineLSE)
type DeviceData struct {
	Device     *invlab.ChromeOSDevice
	UpdateTime *timestamppb.Timestamp
}

// DutStateData holds the invV2 DutState and updatetime
type DutStateData struct {
	DutState   *invlab.DutState
	UpdateTime *timestamppb.Timestamp
}
