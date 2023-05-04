// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package external

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.chromium.org/chromiumos/infra/proto/go/device"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc"

	invV2Api "infra/appengine/cros/lab_inventory/api/v1"
	ufsdevice "infra/unifiedfleet/api/v1/models/chromeos/device"
	"infra/unifiedfleet/app/model/configuration"
)

type fakeInventoryClient struct {
	GetDeviceConfigResp    *device.Config
	GetDeviceConfigErr     bool
	DeviceConfigExistsResp *invV2Api.DeviceConfigsExistsResponse
	DeviceConfigExistsErr  bool
}

func (ic *fakeInventoryClient) DeviceConfigsExists(ctx context.Context, in *invV2Api.DeviceConfigsExistsRequest, opts ...grpc.CallOption) (*invV2Api.DeviceConfigsExistsResponse, error) {
	if ic.DeviceConfigExistsErr {
		return nil, errors.New("error fetching device configs")
	}
	return ic.DeviceConfigExistsResp, nil
}

func (ic *fakeInventoryClient) GetDeviceConfig(ctx context.Context, in *invV2Api.GetDeviceConfigRequest, opts ...grpc.CallOption) (*device.Config, error) {
	if ic.GetDeviceConfigErr {
		return nil, errors.New("error fetching device config")
	}
	return ic.GetDeviceConfigResp, nil
}

// makeDevCfgForTesting creates a basic DeviceConfig. These configs have no
// guarantee to make sense at a domain level, but can be used to verify code
// behavior.
func makeDevCfgForTesting(board, model, variant string, tams []string) *ufsdevice.Config {
	return &ufsdevice.Config{
		Id:  configuration.GetConfigID(board, model, variant),
		Tam: tams,
	}
}

// TestGetDeviceConfig tests behavior of the dual read client. The testing
// environment is seeded with a single device config in datastore. It also
// has a flexible inventory client which can return device configs. This allows
// all combinations of device config existence to be tested
func TestGetDeviceConfig(t *testing.T) {
	tests := []struct {
		name    string
		invResp *device.Config
		invErr  bool
		cfgID   *ufsdevice.ConfigId
		want    *ufsdevice.Config
		wantErr bool
	}{
		{
			name:    "config in UFS",
			invResp: nil,
			invErr:  true,
			cfgID:   configuration.GetConfigID("zork", "gumboz", ""),
			want:    makeDevCfgForTesting("zork", "gumboz", "", []string{"test@google.com"}),
			wantErr: false,
		},
		{
			name: "config in inventory and UFS", // same board/model but inventory has different TAM
			invResp: &device.Config{
				Id: &device.ConfigId{
					PlatformId: &device.PlatformId{Value: "zork"},
					ModelId:    &device.ModelId{Value: "gumboz"},
					VariantId:  &device.VariantId{Value: ""},
				},
				Tam: []string{"inventory@google.com"},
			},
			invErr:  false,
			cfgID:   configuration.GetConfigID("zork", "gumboz", ""),
			want:    makeDevCfgForTesting("zork", "gumboz", "", []string{"inventory@google.com"}),
			wantErr: false,
		},
		{
			name:    "config nowhere",
			invResp: nil,
			invErr:  true,
			cfgID:   configuration.GetConfigID("other", "device", ""),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// setup datastore + populate it
			ctx := memory.UseWithAppID(context.Background(), ("gae-test"))
			datastore.GetTestable(ctx).Consistent(true)
			devCfg := makeDevCfgForTesting("zork", "gumboz", "", []string{"test@google.com"})
			_, err := configuration.BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{devCfg}, configuration.BlankRealmAssigner)
			if err != nil {
				t.Errorf("error setting up test data")
			}

			// setup inventory and dual read clients
			ic := &fakeInventoryClient{
				GetDeviceConfigResp: tt.invResp,
				GetDeviceConfigErr:  tt.invErr,
			}
			c := &DualDeviceConfigClient{
				inventoryClient: ic,
			}

			got, err := c.GetDeviceConfig(ctx, tt.cfgID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DualDeviceConfigClient.GetDeviceConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(ufsdevice.Config{}, ufsdevice.ConfigId{}, ufsdevice.PlatformId{}, ufsdevice.ModelId{}, ufsdevice.VariantId{})); diff != "" {
				t.Errorf("unexpected diff: %s", diff)
			}
		})
	}
}
