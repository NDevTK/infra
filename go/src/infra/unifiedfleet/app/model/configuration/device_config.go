// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configuration

import (
	"context"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/errors"

	ufsdevice "infra/unifiedfleet/api/v1/models/chromeos/device"
	ufsds "infra/unifiedfleet/app/model/datastore"
)

// DeviceConfigKind is the name of the device config entity kind in datastore.
const DeviceConfigKind string = "DeviceConfig"

// DeviceConfigEntity is a datastore entity that tracks a DeviceConfig.
type DeviceConfigEntity struct {
	_kind        string `gae:"$kind,DeviceConfig"`
	ID           string `gae:"$id"`
	DeviceConfig []byte `gae:",noindex"`
	Updated      time.Time
	Realm        string `gae:"realm"` // Realm for this entity
}

// GetProto returns the unmarshaled DeviceConfig.
func (e *DeviceConfigEntity) GetProto() (proto.Message, error) {
	var p ufsdevice.Config
	if err := proto.Unmarshal(e.DeviceConfig, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// RealmAssignerFunc holds logic for associating a `DeviceConfig` with a realm.
type RealmAssignerFunc func(*ufsdevice.Config) string

// BlankRealmAssigner is a RealmAssignerFunc for situations where associating a
// realm is not import, ex. fetching an entity.
func BlankRealmAssigner(d *ufsdevice.Config) string {
	return ""
}

// newDeviceConfigEntityFunc generates a `datastore.NewFunc` that adds a realm
// to the entity based on the `realmAssigner` passed in
//
// This pattern is necessary as the upstream DeviceConfig proto won't have a
// `realm` field, but we need it in the entity. Because each namespace may
// desire a separate realm mapping, we can't hardcode this logic.
func newDeviceConfigEntityFunc(realmAssigner RealmAssignerFunc) ufsds.NewFunc {
	return func(context context.Context, pm proto.Message) (ufsds.FleetEntity, error) {
		p := pm.(*ufsdevice.Config)

		// configs must have model and platform
		if p.Id.GetModelId().GetValue() == "" || p.Id.GetPlatformId().GetValue() == "" {
			return nil, errors.Reason("invalid device config, platform id and model id must be populated").Err()
		}

		idString := GetDeviceConfigIDStr(p.GetId())

		dc, err := proto.Marshal(p)
		if err != nil {
			return nil, errors.Annotate(err, "failed to marshal proto: %s", p).Err()
		}
		return &DeviceConfigEntity{
			ID:           idString,
			DeviceConfig: dc,
			Realm:        realmAssigner(p),
			Updated:      time.Now().UTC(),
		}, nil
	}
}

// GetDeviceConfig fetches a single device config.
func GetDeviceConfig(ctx context.Context, cfgID *ufsdevice.ConfigId) (*ufsdevice.Config, error) {
	pm, err := ufsds.Get(ctx, &ufsdevice.Config{Id: cfgID}, newDeviceConfigEntityFunc(BlankRealmAssigner))
	if err == nil {
		return pm.(*ufsdevice.Config), err
	}
	return nil, err
}

// BatchUpdateDeviceConfigs upserts all configs. The `realmAssigner` determines
// the logic for adding realms to these configs, and can be different in
// different namespaces.
func BatchUpdateDeviceConfigs(ctx context.Context, configs []*ufsdevice.Config, realmAssigner RealmAssignerFunc) ([]*ufsdevice.Config, error) {
	protos := make([]proto.Message, len(configs))
	for i, cfg := range configs {
		protos[i] = cfg
	}
	_, err := ufsds.PutAll(ctx, protos, newDeviceConfigEntityFunc(realmAssigner), true)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// GetDeviceConfigIDStr returns a string as device config short name.
func GetDeviceConfigIDStr(cfgid *ufsdevice.ConfigId) string {
	var platformID, modelID, variantID string
	if v := cfgid.GetPlatformId(); v != nil {
		platformID = strings.ToLower(v.GetValue())
	}
	if v := cfgid.GetModelId(); v != nil {
		modelID = strings.ToLower(v.GetValue())
	}
	if v := cfgid.GetVariantId(); v != nil {
		variantID = strings.ToLower(v.GetValue())
	}
	return strings.Join([]string{platformID, modelID, variantID}, ".")
}
