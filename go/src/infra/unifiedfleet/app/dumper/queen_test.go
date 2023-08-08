// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.chromium.org/luci/gae/impl/memory"
	"google.golang.org/grpc"

	dronequeenapi "infra/appengine/drone-queen/api"
	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/util"
)

// testingContext sets up an environment in memory datastore
func testingContext() context.Context {
	c := context.Background()
	return memory.UseWithAppID(c, ("dev~infra-unified-fleet-system"))
}

// addMachineLSE registers a machine
func addMachineLSE(ctx context.Context, name string) (*ufspb.MachineLSE, error) {
	m, err := inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
		Name: fmt.Sprintf("machinelse-%s", name),
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating machineLSE: %s", err)
	}

	return m, nil
}

// Singleton client
var (
	client = &stubDroneQueenClientImpl{}
)

// stubDroneQueenClientImpl is a stub implementation of drone queen client used for testing
type stubDroneQueenClientImpl struct {
	lastDeclareDUTsCall *dronequeenapi.DeclareDutsRequest
}

// DeclareDuts is a noop that stores the last call in the stub
func (c *stubDroneQueenClientImpl) DeclareDuts(ctx context.Context, in *dronequeenapi.DeclareDutsRequest, opts ...grpc.CallOption) (*dronequeenapi.DeclareDutsResponse, error) {
	c.lastDeclareDUTsCall = in
	return &dronequeenapi.DeclareDutsResponse{}, nil
}

// getStubSingletonDroneQueenClient returns the stub singleton client
func getStubSingletonDroneQueenClient(ctx context.Context) (dronequeenapi.InventoryProviderClient, error) {
	return client, nil
}

// TestPushToDroneQueenNamespaces creates DUTs in three namespaces, and then
// verifies correct behavior wrt which namespaces we push DUTs for
func TestPushToDroneQueenNamesapces(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	droneQueenGenerator = getStubSingletonDroneQueenClient

	osCtx, _ := util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	partnerCtx, _ := util.SetupDatastoreNamespace(ctx, util.OSPartnerNamespace)
	browserCtx, _ := util.SetupDatastoreNamespace(ctx, util.BrowserNamespace)

	osMachine, _ := addMachineLSE(osCtx, "os")
	partnerMachine, _ := addMachineLSE(partnerCtx, "partner")
	_, _ = addMachineLSE(browserCtx, "browser")

	// only want os, partner machines to be pushed
	want := &dronequeenapi.DeclareDutsRequest{
		AvailableDuts: []*dronequeenapi.DeclareDutsRequest_Dut{
			{Name: osMachine.Name},
			{Name: partnerMachine.Name},
		},
	}

	err := pushToDroneQueen(ctx)
	if err != nil {
		t.Errorf("err when pushing to drone queen: %s", err)
	}

	if diff := cmp.Diff(client.lastDeclareDUTsCall, want, cmpopts.IgnoreUnexported(dronequeenapi.DeclareDutsRequest_Dut{}, dronequeenapi.DeclareDutsRequest{})); diff != "" {
		t.Errorf("Call to drone queen had unexpected diff:\n%s", diff)
	}
}
