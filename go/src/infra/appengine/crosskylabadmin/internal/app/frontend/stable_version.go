// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"go.chromium.org/chromiumos/infra/proto/go/lab_platform"
	authclient "go.chromium.org/luci/auth"
	gitilesApi "go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/grpc/grpcutil"
	"go.chromium.org/luci/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/config"
	dssv "infra/appengine/crosskylabadmin/internal/app/frontend/datastore/stableversion"
	"infra/appengine/crosskylabadmin/internal/app/frontend/datastore/stableversion/satlab"
	"infra/appengine/crosskylabadmin/internal/ufs"
	"infra/cros/stableversion"
	"infra/libs/git"
	"infra/libs/skylab/common/heuristics"
	"infra/libs/skylab/inventory"
)

// StableVersionGitClientFactory is a constructor for a git client pointed at the source of truth
// for the stable version information
type StableVersionGitClientFactory func(c context.Context) (git.ClientInterface, error)

// ServerImpl implements the fleet.InventoryServer interface.
type ServerImpl struct {
	// StableVersionGitClientFactory
	StableVersionGitClientFactory StableVersionGitClientFactory
}

type getStableVersionRecordsResult struct {
	cros     map[string]string
	faft     map[string]string
	firmware map[string]string
}

const beagleboneServo = "beaglebone_servo"

// DumpStableVersionToDatastore takes stable version info from the git repo where it lives
// and dumps it to datastore
func (is *ServerImpl) DumpStableVersionToDatastore(ctx context.Context, in *fleet.DumpStableVersionToDatastoreRequest) (*fleet.DumpStableVersionToDatastoreResponse, error) {
	client, err := is.newStableVersionGitClient(ctx)
	if err != nil {
		logging.Errorf(ctx, "get git client: %s", err)
		return nil, errors.Annotate(err, "get git client").Err()
	}
	return dumpStableVersionToDatastoreImpl(ctx, client.GetFile)
}

// GetStableVersion implements the method from fleet.InventoryServer interface
func (is *ServerImpl) GetStableVersion(ctx context.Context, req *fleet.GetStableVersionRequest) (resp *fleet.GetStableVersionResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()

	return getStableVersionImpl(ctx, req.GetBuildTarget(), req.GetModel(), req.GetHostname(), req.GetSatlabInformationalQuery())
}

// getSatlabStableVersion gets a stable version for a satlab device.
//
// It returns a full response if there's no error, and a boolean ok which determines whether the error should cause
// the request to fail or not.
func getSatlabStableVersion(ctx context.Context, buildTarget string, model string, hostname string) (resp *fleet.GetStableVersionResponse, ok bool, e error) {
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
		dut, err := getDUT(ctx, hostname)
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
func getStableVersionImpl(ctx context.Context, buildTarget string, model string, hostname string, satlabInformationalQuery bool) (*fleet.GetStableVersionResponse, error) {
	logging.Infof(ctx, "getting stable version for buildTarget: %s and model: %s", buildTarget, model)

	wantSatlab := heuristics.LooksLikeSatlabDevice(hostname) || satlabInformationalQuery

	if wantSatlab {
		resp, ok, err := getSatlabStableVersion(ctx, buildTarget, model, hostname)
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
	out, err := getStableVersionImplWithHostname(ctx, hostname)
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
func getStableVersionImplWithHostname(ctx context.Context, hostname string) (*fleet.GetStableVersionResponse, error) {
	var err error

	dut, err := getDUT(ctx, hostname)
	if err != nil {
		return nil, errors.Annotate(err, "failed to get DUT %q", dut).Err()
	}

	buildTarget := dut.GetCommon().GetLabels().GetBoard()
	model := dut.GetCommon().GetLabels().GetModel()

	out, err := getStableVersionImplNoHostname(ctx, buildTarget, model)
	if err != nil {
		return nil, errors.Annotate(err, "failed to get stable version info").Err()
	}
	return out, nil
}

// getDUTOverrideForTests is an override for tests only.
//
// Do not set this variable for any other purpose.
var getDUTOverrideForTests func(context.Context, string) (*inventory.DeviceUnderTest, error) = nil

// getDUT returns the DUT associated with a particular hostname from datastore
func getDUT(ctx context.Context, hostname string) (*inventory.DeviceUnderTest, error) {
	if getDUTOverrideForTests != nil {
		return getDUTOverrideForTests(ctx, hostname)
	}

	// Call UFS directly to get DUT info, if fails, falling back to use the old workflow
	dutV1, err := ufs.GetDutV1(ctx, hostname)
	if err != nil {
		logging.Infof(ctx, "getDUT: fail to get DUT info from UFS for host %s: %s", hostname, err)
		return nil, err
	}
	return dutV1, err
}

// looksLikeServod is a heuristic to detect whether a servod entry.
// Historically, these used "localhost" as a hostname.
// Currently, they look like satlab-0⬛⬛⬛⬛⬛⬛⬛⬛⬛-host1-docker_servod.
//
// The suffix can only be "docker_servod".
// I am intentionally keeping the number of supported suffixes small so that misnamed devices are surfaced in a reasonably
// intuitive way.
// See b/187895178 comment #13 for details.
func validateServod(hostname string) error {
	if hostname == "localhost" {
		return nil
	}
	if strings.Contains(hostname, "docker_servod") {
		return nil
	}
	// TODO(gregorynisbet): Consider removing this special case. Formerly, there was a `hostname == ""` check at the sole call site.
	if hostname == "" {
		return nil
	}
	// Detect common errors and give a helpful error message to our users.
	// Any error that isn't validateServodFallbackError indicates that we should not fall back.
	if strings.Contains(hostname, "docker-servod") {
		return errors.New(`validate servod: use "docker_servod" with an underscore, not "docker-servod" with a hyphen`)
	}
	return validateServodFallbackError
}

// validateServodFallbackError indicates that we should fallback.
var validateServodFallbackError = errors.New("validate servod: should fall back")

// maybeSetReason sets the reason on a stable version response if the response is non-nil and the reason is "".
func maybeSetReason(resp *fleet.GetStableVersionResponse, msg string) {
	if resp != nil && resp.GetReason() == "" {
		resp.Reason = msg
	}
}

func (is *ServerImpl) newStableVersionGitClient(ctx context.Context) (git.ClientInterface, error) {
	if is.StableVersionGitClientFactory != nil {
		return is.StableVersionGitClientFactory(ctx)
	}
	hc, err := getAuthenticatedHTTPClient(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "newStableVersionGitClient").Err()
	}
	return getStableVersionGitClient(ctx, hc)
}

// dumpStableVersionToDatastoreImpl takes some way of getting a file and a context and writes to datastore
func dumpStableVersionToDatastoreImpl(ctx context.Context, getFile func(context.Context, string) (string, error)) (*fleet.DumpStableVersionToDatastoreResponse, error) {
	contents, err := getFile(ctx, config.Get(ctx).StableVersionConfig.StableVersionDataPath)
	if err != nil {
		logging.Errorf(ctx, "fetch file: %s", err)
		return nil, errors.Annotate(err, "fetch file").Err()
	}
	stableVersions, err := parseStableVersions(contents)
	if err != nil {
		logging.Errorf(ctx, "parse json: %s", err)
		return nil, errors.Annotate(err, "parse json").Err()
	}
	m := getStableVersionRecords(ctx, stableVersions)
	// TODO(gregorynisbet): Walk the board;models that exist and bring them up to date one-by-one
	// inside one transaction per key instead of discarding the result.
	keys, err := getAllBoardModels(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "dump stable version to datastore implementation").Err()
	}

	// allKeys are the keys from datastore and the incoming map. The incoming map is more
	// authoritative than the contents of datastore.
	allKeys := make(map[string]bool)
	for k := range keys {
		allKeys[k] = true
	}
	for _, versionMap := range []map[string]string{m.cros, m.firmware, m.faft} {
		for k := range versionMap {
			allKeys[k] = true
		}
	}

	// nil and "No Such Entity" are both acceptable get responses.
	// The latter unambiguously indicates that we succesfully determined that an item does not exist.
	isAcceptableGetResponse := func(e error) bool {
		return err == nil || datastore.IsErrNoSuchEntity(err)
	}

	for boardModel := range allKeys {
		if err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			cros := &dssv.CrosStableVersionEntity{ID: boardModel}
			if err := datastore.Get(ctx, cros); !isAcceptableGetResponse(err) {
				return errors.Annotate(err, "dump stable version to datastore implementation: put cros stable version %q", boardModel).Err()
			}
			if err := cros.ImposeVersion(ctx, m.cros[boardModel]); err != nil {
				return errors.Annotate(err, "dump stable version to datastore implementation %q", boardModel).Err()
			}
			firmware := &dssv.FirmwareStableVersionEntity{ID: boardModel}
			if err := datastore.Get(ctx, firmware); !isAcceptableGetResponse(err) {
				return errors.Annotate(err, "dump stable version to datastore implementation: put firmware stable version %q", boardModel).Err()
			}
			if err := firmware.ImposeVersion(ctx, m.firmware[boardModel]); err != nil {
				return errors.Annotate(err, "dump stable version to datastore implementation %q", boardModel).Err()
			}
			faft := &dssv.FaftStableVersionEntity{ID: boardModel}
			if err := datastore.Get(ctx, faft); !isAcceptableGetResponse(err) {
				return errors.Annotate(err, "dump stable version to datastore implementation: put faft stable version for %q", boardModel).Err()
			}
			if err := faft.ImposeVersion(ctx, m.faft[boardModel]); err != nil {
				return errors.Annotate(err, "dump stable version to datastore implementation for %q", boardModel).Err()
			}
			return nil
		}, nil); err != nil {
			return nil, err
		}
	}

	logging.Infof(ctx, "successfully wrote stable versions")
	return &fleet.DumpStableVersionToDatastoreResponse{}, nil
}

func parseStableVersions(contents string) (*lab_platform.StableVersions, error) {
	var stableVersions lab_platform.StableVersions
	if err := jsonpb.Unmarshal(strings.NewReader(contents), &stableVersions); err != nil {
		return nil, errors.Annotate(err, "unmarshal stableversions json").Err()
	}
	return &stableVersions, nil
}

// getStableVersionRecords takes a StableVersions proto and produces a structure containing maps from
// key names (buildTarget or buildTarget+model) to stable version strings
func getStableVersionRecords(ctx context.Context, stableVersions *lab_platform.StableVersions) getStableVersionRecordsResult {
	cros := make(map[string]string)
	faft := make(map[string]string)
	firmware := make(map[string]string)
	for _, item := range stableVersions.GetCros() {
		buildTarget := item.GetKey().GetBuildTarget().GetName()
		model := item.GetKey().GetModelId().GetValue()
		version := item.GetVersion()
		key, err := stableversion.JoinBuildTargetModel(buildTarget, model)
		if err != nil {
			logging.Infof(ctx, "buildTarget and/or model contains invalid sequence: %s", err)
			continue
		}
		cros[key] = version
	}
	for _, item := range stableVersions.GetFirmware() {
		buildTarget := item.GetKey().GetBuildTarget().GetName()
		model := item.GetKey().GetModelId().GetValue()
		version := item.GetVersion()
		key, err := stableversion.JoinBuildTargetModel(buildTarget, model)
		if err != nil {
			logging.Infof(ctx, "buildTarget and/or model contains invalid sequence: %s", err)
			continue
		}
		firmware[key] = version
	}
	for _, item := range stableVersions.GetFaft() {
		buildTarget := item.GetKey().GetBuildTarget().GetName()
		model := item.GetKey().GetModelId().GetValue()
		version := item.GetVersion()
		key, err := stableversion.JoinBuildTargetModel(buildTarget, model)
		if err != nil {
			logging.Infof(ctx, "buildTarget and/or model contains invalid sequence: %s", err)
			continue
		}
		faft[key] = version
	}
	return getStableVersionRecordsResult{
		cros:     cros,
		faft:     faft,
		firmware: firmware,
	}
}

func getAuthenticatedHTTPClient(ctx context.Context) (*http.Client, error) {
	transport, err := auth.GetRPCTransport(ctx, auth.AsSelf, auth.WithScopes(authclient.OAuthScopeEmail, gitilesApi.OAuthScope))
	if err != nil {
		return nil, errors.Annotate(err, "new authenticated http client").Err()
	}
	return &http.Client{Transport: transport}, nil
}

func getStableVersionGitClient(ctx context.Context, hc *http.Client) (git.ClientInterface, error) {
	cfg := config.Get(ctx)
	s := cfg.StableVersionConfig
	if s == nil {
		return nil, fmt.Errorf("DumpStableVersionToDatastore: app config does not have StableVersionConfig")
	}
	client, err := git.NewClient(ctx, hc, s.GerritHost, s.GitilesHost, s.Project, s.Branch)
	if err != nil {
		return nil, errors.Annotate(err, "get git client").Err()
	}
	return client, nil
}

// getAllBoardModels gets all the keys of the form board;model that currently exist in datastore.
func getAllBoardModels(ctx context.Context) (map[string]bool, error) {
	out := make(map[string]bool)
	if err := datastore.Run(ctx, datastore.NewQuery(dssv.CrosStableVersionKind), func(ent *dssv.CrosStableVersionEntity) {
		out[ent.ID] = true
	}); err != nil {
		return nil, errors.Annotate(err, "get all board models").Err()
	}
	if err := datastore.Run(ctx, datastore.NewQuery(dssv.FirmwareStableVersionKind), func(ent *dssv.FirmwareStableVersionEntity) {
		out[ent.ID] = true
	}); err != nil {
		return nil, errors.Annotate(err, "get all board models").Err()
	}
	if err := datastore.Run(ctx, datastore.NewQuery(dssv.FaftStableVersionKind), func(ent *dssv.FaftStableVersionEntity) {
		out[ent.ID] = true
	}); err != nil {
		return nil, errors.Annotate(err, "get all board models").Err()
	}
	return out, nil
}
