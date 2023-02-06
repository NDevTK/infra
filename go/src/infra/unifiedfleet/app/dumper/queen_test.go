// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"
	"fmt"
	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/frontend/fake"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"
	"infra/unifiedfleet/app/util"
	"testing"

	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/gae/service/datastore"
)

func testingContext() context.Context {
	c := gaetesting.TestingContextWithAppID("dev~infra-unified-fleet-system")
	c = gologger.StdConfig.Use(c)
	c = logging.SetLevel(c, logging.Error)
	c = config.Use(c, &config.Config{})
	c = fake.FakePubsubClientInterface(c)
	datastore.GetTestable(c).Consistent(true)
	return c
}

func mockMachineLSE(ctx context.Context, name string) (*ufspb.MachineLSE, error) {
	machine1 := &ufspb.Machine{
		Name: fmt.Sprintf("machine-%s", name),
	}
	_, merr := registration.CreateMachine(ctx, machine1)
	if merr != nil {
		return nil, fmt.Errorf("Error creating machine: %s", merr)
	}

	m, err := inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
		Name: fmt.Sprintf("machinelse-%s", name),
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating machineLSE: %s", err)
	}

	return m, nil
}

func TestPushToDroneQueen(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	osCtx, _ := util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	partnerCtx, _ := util.SetupDatastoreNamespace(ctx, util.OSPartnerNamespace)

	osMachineLSE, err := mockMachineLSE(osCtx, "os")
	partnerMachineLSE, err := mockMachineLSE(partnerCtx, "partner")

	_, err = inventory.GetMachineLSE(osCtx, osMachineLSE.Name)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	_, err = inventory.GetMachineLSE(osCtx, partnerMachineLSE.Name)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
}
