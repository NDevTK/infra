// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"fmt"
	"time"

	ufsapi "infra/unifiedfleet/api/v1/rpc"
	ufsutil "infra/unifiedfleet/app/util"

	schedukepb "go.chromium.org/chromiumos/config/go/test/scheduling"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/gcloud/googleoauth"
	"google.golang.org/grpc/metadata"
)

// listLeasesFromScheduke sends a request to Scheduke to list all requested and
// in-flight leases for the current user.
func listLeasesFromScheduke(ctx context.Context, flags *authcli.Flags) ([]*schedukepb.TaskWithState, error) {
	user, err := getUserEmail(ctx, flags)
	if err != nil {
		return nil, err
	}

	sc, err := NewLocalSchedukeClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := sc.ReadTaskStates(nil, []string{user}, nil)
	if err != nil {
		return nil, err
	}
	return resp.Tasks, nil
}

// leaseDeviceFromScheduke sends a lease request to Scheduke and waits for the
// request to be fulfilled before returning a device name.
func leaseDeviceFromScheduke(ctx context.Context, flags *authcli.Flags, dims map[string][]string, mins int64) (string, error) {
	user, err := getUserEmail(ctx, flags)
	if err != nil {
		return "", err
	}

	sc, err := NewLocalSchedukeClient(ctx)
	if err != nil {
		return "", err
	}

	t := time.Now()
	req := leaseRequest(dims, mins, user, t)
	scheduleResp, err := sc.ScheduleExecution(req)
	if err != nil {
		return "", err
	}
	leaseID, ok := scheduleResp.GetIds()[schedukeTaskKey]
	if !ok {
		return "", fmt.Errorf("respose %v from Scheduke did not include an ID for the requested lease", scheduleResp)
	}

	leaseIDsList := []int64{leaseID}
	for {
		time.Sleep(pollingInterval)

		resp, err := sc.ReadTaskStates(leaseIDsList, nil, nil)
		if err != nil {
			return "", fmt.Errorf("error polling Scheduke for lease status: %w", err)
		}
		if numTasks := len(resp.GetTasks()); numTasks != 1 {
			return "", fmt.Errorf("response %v from Scheduke returned %d tasks (expected exactly 1)", resp, numTasks)
		}

		taskWithState := resp.GetTasks()[0]
		switch taskWithState.GetState() {
		case schedukepb.TaskState_LAUNCHED:
			return taskWithState.GetDeviceName(), nil
		case schedukepb.TaskState_CANCELED:
			return "", fmt.Errorf("lease %d was unexpectedly cancelled", leaseID)
		case schedukepb.TaskState_EXPIRED:
			return "", fmt.Errorf("lease %d expired without being fulfilled", leaseID)
		case schedukepb.TaskState_COMPLETED:
			return "", fmt.Errorf("lease already launched and completed; consider requesting a longer lease")
		}
	}
}

// getDeviceInfo calls UFS to add information about the given device in-place.
func addDeviceInfo(ctx context.Context, di *DeviceInfo, uc ufsapi.FleetClient) error {
	ctx = ufsCTX(ctx)
	var err error
	di.LabSetup, err = uc.GetMachineLSE(ctx, &ufsapi.GetMachineLSERequest{
		Name: ufsutil.AddPrefix(ufsutil.MachineLSECollection, di.Name),
	})
	if err != nil {
		return err
	}
	// Only attempt to retrieve information about the device's machine if its
	// lab setup contains a machine name.
	machineNames := di.LabSetup.GetMachines()
	if len(machineNames) == 0 || machineNames[0] == "" {
		return nil
	}
	di.Machine, err = uc.GetMachine(ctx, &ufsapi.GetMachineRequest{
		Name: ufsutil.AddPrefix(ufsutil.MachineCollection, machineNames[0]),
	})
	if err != nil {
		return err
	}
	return nil
}

// ufsCTX adds an "os" namespace to the given context, which
// is required for API calls to UFS.
func ufsCTX(ctx context.Context) context.Context {
	osMetadata := metadata.Pairs(ufsutil.Namespace, ufsutil.OSNamespace)
	return metadata.NewOutgoingContext(ctx, osMetadata)
}

// getUserEmail parses the given auth flags and returns the email of the
// authenticated crosfleet user.
func getUserEmail(ctx context.Context, flags *authcli.Flags) (string, error) {
	opts, err := flags.Options()
	if err != nil {
		return "", nil
	}
	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, opts)
	tempToken, err := authenticator.GetAccessToken(time.Minute)
	if err != nil {
		return "", err
	}
	authInfo, err := googleoauth.GetTokenInfo(ctx, googleoauth.TokenInfoParams{
		AccessToken: tempToken.AccessToken,
	})
	if err != nil {
		return "", err
	}
	if authInfo.Email == "" {
		return "", fmt.Errorf("no email found for the current user")
	}
	return authInfo.Email, nil
}
