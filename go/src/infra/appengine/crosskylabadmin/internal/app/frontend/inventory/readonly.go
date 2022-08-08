// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package inventory implements the fleet.Inventory service end-points of
// corsskylabadmin.
package inventory

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/grpc/grpcutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	dssv "infra/appengine/crosskylabadmin/internal/app/frontend/datastore/stableversion"
	"infra/appengine/crosskylabadmin/internal/app/frontend/datastore/stableversion/satlab"
	"infra/appengine/crosskylabadmin/internal/ufs"
	"infra/libs/skylab/common/heuristics"
	"infra/libs/skylab/inventory"
)

const beagleboneServo = "beaglebone_servo"

// GetStableVersion implements the method from fleet.InventoryServer interface
func (is *ServerImpl) GetStableVersion(ctx context.Context, req *fleet.GetStableVersionRequest) (resp *fleet.GetStableVersionResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()

	ic, err := is.newInventoryClient(ctx)
	if err != nil {
		logging.Errorf(ctx, "Failed to create inventory client: %s", err.Error())
		logging.Infof(ctx, "Fall back to legacy flow")
		ic = nil
	}
	return getStableVersionImpl(ctx, ic, req.GetBuildTarget(), req.GetModel(), req.GetHostname(), req.GetSatlabInformationalQuery())
}

// getSatlabStableVersion gets a stable version for a satlab device.
//
// It returns a full response if there's no error, and a boolean ok which determines whether the error should cause
// the request to fail or not.
func getSatlabStableVersion(ctx context.Context, ic inventoryClient, buildTarget string, model string, hostname string) (resp *fleet.GetStableVersionResponse, ok bool, e error) {
	logging.Infof(ctx, "using satlab flow board:%q model:%q host:%q", buildTarget, model, hostname)

	if hostnameID := satlab.MakeSatlabStableVersionID(hostname, "", ""); hostnameID != "" {
		entry, err := satlab.GetSatlabStableVersionEntryByRawID(ctx, hostnameID)
		switch {
		case err == nil:
			reason := fmt.Sprintf("looked up satlab device using id %q", hostnameID)
			resp := &fleet.GetStableVersionResponse{
				CrosVersion:     entry.OS,
				FirmwareVersion: entry.FW,
				FaftVersion:     entry.FWImage,
				Reason:          reason,
			}
			return resp, true, nil
		case datastore.IsErrNoSuchEntity(err):
			// Do nothing. If there is no override for the hostname, it is correct to
			// move on to the next case, checking by board & model.
		default:
			return nil, false, status.Errorf(codes.NotFound, "get satlab: %s", err)
		}
	}

	if hostname != "" && (buildTarget == "" || model == "") {
		logging.Infof(ctx, "looking up inventory info for DUT host:%q board:%q model:%q in order to get board and model info", hostname, buildTarget, model)
		dut, err := getDUT(ctx, ic, hostname)
		if err != nil {
			return nil, false, status.Errorf(codes.NotFound, "get satlab: processing dut %q: %s", hostname, err)
		}

		buildTarget = dut.GetCommon().GetLabels().GetBoard()
		model = dut.GetCommon().GetLabels().GetModel()
	}

	if boardModelID := satlab.MakeSatlabStableVersionID("", buildTarget, model); boardModelID != "" {
		entry, err := satlab.GetSatlabStableVersionEntryByRawID(ctx, boardModelID)
		switch {
		case err == nil:
			reason := fmt.Sprintf("looked up satlab device using id %q", boardModelID)

			resp := &fleet.GetStableVersionResponse{
				CrosVersion:     entry.OS,
				FirmwareVersion: entry.FW,
				FaftVersion:     entry.FWImage,
				Reason:          reason,
			}

			return resp, true, nil
		case datastore.IsErrNoSuchEntity(err):
			// Do nothing.
		default:
			return nil, false, status.Errorf(codes.NotFound, "get satlab: lookup by board/model %q: %s", boardModelID, err)
		}
	}

	return nil, true, status.Error(codes.Aborted, "get satlab: falling back %s")
}

// getStableVersionImpl returns all the stable versions associated with a given buildTarget and model
// NOTE: hostname is explicitly allowed to be "". If hostname is "", then no hostname was provided in the GetStableVersion RPC call
// ALSO NOTE: If the hostname is "", then we assume that the device is not a satlab device and therefore we should not fall back to satlab.
func getStableVersionImpl(ctx context.Context, ic inventoryClient, buildTarget string, model string, hostname string, satlabInformationalQuery bool) (*fleet.GetStableVersionResponse, error) {
	logging.Infof(ctx, "getting stable version for buildTarget: %s and model: %s", buildTarget, model)

	wantSatlab := heuristics.LooksLikeSatlabDevice(hostname) || satlabInformationalQuery

	if wantSatlab {
		resp, ok, err := getSatlabStableVersion(ctx, ic, buildTarget, model, hostname)
		switch {
		case err == nil:
			return resp, nil
		case ok:
			// Do nothing and fall back
		default:
			return nil, err
		}
	}

	if hostname == "" {
		if buildTarget == "" || model == "" {
			return nil, status.Errorf(codes.FailedPrecondition, "search criteria must be provided.")
		}
		logging.Infof(ctx, "hostname not provided, using buildTarget (%s) and model (%s)", buildTarget, model)
		out, err := getStableVersionImplNoHostname(ctx, buildTarget, model)
		if err == nil {
			msg := "looked up board %q and model %q"
			if wantSatlab {
				msg = "wanted satlab, falling back to board %q and model %q"
			}
			maybeSetReason(out, fmt.Sprintf(msg, buildTarget, model))
			return out, nil
		}
		return out, status.Errorf(codes.NotFound, "get stable version impl: %s", err)
	}

	// Default case, not a satlab device.
	logging.Infof(ctx, "hostname (%s) provided, ignoring user-provided buildTarget (%s) and model (%s)", hostname, buildTarget, model)
	out, err := getStableVersionImplWithHostname(ctx, ic, hostname)
	if err == nil {
		msg := "looked up non-satlab device hostname %q"
		if wantSatlab {
			msg = "falling back to non-satlab path for device %q"
		}
		maybeSetReason(out, fmt.Sprintf(msg, hostname))
		return out, nil
	}
	return out, status.Errorf(codes.NotFound, "get stable version impl: %s", err)
}

// getStableVersionImplNoHostname returns stableversion information given a buildTarget and model
// TODO(gregorynisbet): Consider under what circumstances an error leaving this function
// should be considered transient or non-transient.
// If the dut in question is a beaglebone servo, then failing to get the firmware version
// is non-fatal.
func getStableVersionImplNoHostname(ctx context.Context, buildTarget string, model string) (*fleet.GetStableVersionResponse, error) {
	logging.Infof(ctx, "getting stable version for buildTarget: (%s) and model: (%s)", buildTarget, model)
	var err error
	out := &fleet.GetStableVersionResponse{}

	out.CrosVersion, err = dssv.GetCrosStableVersion(ctx, buildTarget, model)
	if err != nil {
		return nil, errors.Annotate(err, "getStableVersionImplNoHostname").Err()
	}
	out.FaftVersion, err = dssv.GetFaftStableVersion(ctx, buildTarget, model)
	if err != nil {
		logging.Infof(ctx, "faft stable version does not exist: %#v", err)
	} else {
		logging.Infof(ctx, "Got faft stable version %s from datastore", out.FaftVersion)
	}
	// successful early exit if we have a beaglebone servo
	if buildTarget == beagleboneServo || model == beagleboneServo {
		maybeSetReason(out, "looks like beaglebone")
		return out, nil
	}
	out.FirmwareVersion, err = dssv.GetFirmwareStableVersion(ctx, buildTarget, model)
	if err != nil {
		logging.Infof(ctx, "firmware version does not exist: %#v", err)
	}
	return out, nil
}

// getStableVersionImplWithHostname return stable version information given just a hostname
// TODO(gregorynisbet): Consider under what circumstances an error leaving this function
// should be considered transient or non-transient.
func getStableVersionImplWithHostname(ctx context.Context, ic inventoryClient, hostname string) (*fleet.GetStableVersionResponse, error) {
	var err error

	// If the DUT in question is a labstation or a servo (i.e. is a servo host), then it does not have
	// its own servo host.
	if looksLikeServo(hostname) {
		return getStableVersionImplNoHostname(ctx, beagleboneServo, "")
	}

	dut, err := getDUT(ctx, ic, hostname)
	if err != nil {
		return nil, errors.Annotate(err, "failed to get DUT %q", dut).Err()
	}

	buildTarget := dut.GetCommon().GetLabels().GetBoard()
	model := dut.GetCommon().GetLabels().GetModel()

	out, err := getStableVersionImplNoHostname(ctx, buildTarget, model)
	if err != nil {
		return nil, errors.Annotate(err, "failed to get stable version info").Err()
	}

	if heuristics.LooksLikeLabstation(hostname) {
		return out, nil
	}
	servoHostHostname, err := getServoHostHostname(dut)
	if err != nil {
		// Some DUTs, particularly High Touch Lab DUTs legitimately do not have servos.
		// See b/162030132 for context.
		logging.Infof(ctx, "failed to get servo host for %q", hostname)
		return out, nil
	}
	if looksLikeFakeServo(servoHostHostname) {
		logging.Infof(ctx, "concluded servo hostname is fake %q", servoHostHostname)
		return out, nil
	}
	servoStableVersion, err := getCrosVersionFromServoHost(ctx, ic, servoHostHostname)
	if err != nil {
		return nil, errors.Annotate(err, "getting cros version from servo host %q", servoHostHostname).Err()
	}
	out.ServoCrosVersion = servoStableVersion

	return out, nil
}

// getServoHostHostname gets the servo host hostname associated with a dut
// for instance, a labstation is a servo host.
func getServoHostHostname(dut *inventory.DeviceUnderTest) (string, error) {
	attrs := dut.GetCommon().GetAttributes()
	if len(attrs) == 0 {
		return "", errors.Reason("attributes for dut with hostname %q is unexpectedly empty", dut.GetCommon().GetHostname()).Err()
	}
	for _, item := range attrs {
		key := item.GetKey()
		value := item.GetValue()
		if key == "servo_host" {
			if value == "" {
				return "", errors.Reason("\"servo_host\" attribute unexpectedly has value \"\" for hostname %q", dut.GetCommon().GetHostname()).Err()
			}
			return value, nil
		}
	}
	return "", errors.Reason("no \"servo_host\" attribute for hostname %q", dut.GetCommon().GetHostname()).Err()
}

// getDUTOverrideForTests is an override for tests only.
//
// Do not set this variable for any other purpose.
var getDUTOverrideForTests func(context.Context, string) (*inventory.DeviceUnderTest, error) = nil

// getDUT returns the DUT associated with a particular hostname from datastore
func getDUT(ctx context.Context, ic inventoryClient, hostname string) (*inventory.DeviceUnderTest, error) {
	if getDUTOverrideForTests != nil {
		return getDUTOverrideForTests(ctx, hostname)
	}

	// Call UFS directly to get DUT info, if fails, falling back to use the old workflow
	dutV1, err := ufs.GetDutV1(ctx, hostname)
	if err != nil {
		logging.Infof(ctx, "getDUT: fail to get DUT info from UFS for host %s: %s", hostname, err)
	} else {
		return dutV1, err
	}

	logging.Infof(ctx, "getDUT: fallback to get DUT info from Inv2 for host %s", hostname)
	if ic == nil {
		return nil, errors.Reason("Inventory Client cannot be nil").Err()
	}
	resp, _, err := ic.getDutInfo(ctx, &fleet.GetDutInfoRequest{
		Hostname: hostname,
	})
	if err != nil {
		return nil, errors.Annotate(err, "getting serialized DUT by hostname for %q", hostname).Err()
	}
	dut := &inventory.DeviceUnderTest{}
	if err := proto.Unmarshal(resp, dut); err != nil {
		return nil, errors.Annotate(err, "unserializing DUT for hostname %q", hostname).Err()
	}
	return dut, nil
}

// This is a heuristic to check if something is a servo and might be wrong.
func looksLikeServo(hostname string) bool {
	return strings.Contains(hostname, "servo")
}

// looksLikeFakeServo is a heuristic to check if a given hostname is an obviously
// fake entry such as an empty string or dummy_host or FAKE_SERVO_HOST or similar
func looksLikeFakeServo(hostname string) bool {
	h := strings.ToLower(hostname)
	return h == "" || strings.Contains(h, "dummy") || strings.Contains(h, "fake")
}

// getCrosVersionFromServoHost returns the cros version associated with a particular servo host
// hostname : hostname of the servo host (e.g. labstation)
// NOTE: If hostname is "localhost", task is for Satlab Containerized servod.
// NOTE: If hostname is "", this indicates the absence of a relevant servo host. This can happen if the DUT in question is already a labstation, for instance.
// NOTE: The cros version will be empty "" if the labstation does not exist. Because we don't re-image labstations as part of repair, the absence of a stable CrOS version for a labstation is not an error.
func getCrosVersionFromServoHost(ctx context.Context, ic inventoryClient, hostname string) (string, error) {
	if hostname == "" || hostname == "localhost" {
		logging.Infof(ctx, "Skipping getting cros version. Servo host hostname is %q", hostname)
		return "", nil
	}
	if heuristics.LooksLikeLabstation(hostname) {
		dut, err := getDUT(ctx, ic, hostname)
		if err != nil {
			logging.Infof(ctx, "get labstation dut info; %s", err)
			return "", nil
		}
		buildTarget := dut.GetCommon().GetLabels().GetBoard()
		model := dut.GetCommon().GetLabels().GetModel()
		if buildTarget == "" {
			return "", errors.Reason("no buildTarget for hostname %q", hostname).Err()
		}
		out, err := dssv.GetCrosStableVersion(ctx, buildTarget, model)
		if err != nil {
			return "", errors.Annotate(err, "getting labstation stable version").Err()
		}
		return out, nil
	}
	if looksLikeServo(hostname) {
		// TODO(gregorynisbet): getting the stable version of a beaglebone servo is dependent on the fallback
		// behavior
		out, err := dssv.GetCrosStableVersion(ctx, beagleboneServo, beagleboneServo)
		if err != nil {
			return "", errors.Annotate(err, "getting beaglebone servo stable version").Err()
		}
		return out, nil
	}
	return "", errors.Reason("unrecognized hostname %q is not a labstation or beaglebone servo", hostname).Err()
}

// maybeSetReason sets the reason on a stable version response if the response is non-nil and the reason is "".
func maybeSetReason(resp *fleet.GetStableVersionResponse, msg string) {
	if resp != nil && resp.GetReason() == "" {
		resp.Reason = msg
	}
}
