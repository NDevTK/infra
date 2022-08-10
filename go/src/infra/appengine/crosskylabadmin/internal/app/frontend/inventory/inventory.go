// Copyright 2018 The LUCI Authors.
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
	"net/http"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"go.chromium.org/chromiumos/infra/proto/go/lab_platform"
	authclient "go.chromium.org/luci/auth"
	gitilesApi "go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/config"
	dataSV "infra/appengine/crosskylabadmin/internal/app/frontend/datastore/stableversion"
	"infra/cros/stableversion"
	"infra/libs/git"
)

// TrackerFactory is a constructor for a TrackerServer object.
type TrackerFactory func() fleet.TrackerServer

// StableVersionGitClientFactory is a constructor for a git client pointed at the source of truth
// for the stable version information
type StableVersionGitClientFactory func(c context.Context) (git.ClientInterface, error)

// ServerImpl implements the fleet.InventoryServer interface.
type ServerImpl struct {
	// TrackerServerFactory is a required factory function for creating a tracker object.
	//
	// TODO(pprabhu) Move tracker/tasker to individual sub-packages and inject
	// dependencies directly (instead of factory functions).
	TrackerFactory TrackerFactory

	// StableVersionGitClientFactory
	StableVersionGitClientFactory StableVersionGitClientFactory
}

type getStableVersionRecordsResult struct {
	cros     map[string]string
	faft     map[string]string
	firmware map[string]string
}

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
	merr := errors.NewMultiError()
	if err := dataSV.PutManyCrosStableVersion(ctx, m.cros); err != nil {
		merr = append(merr, errors.Annotate(err, "put cros stable version").Err())
	}
	if err := dataSV.PutManyFirmwareStableVersion(ctx, m.firmware); err != nil {
		merr = append(merr, errors.Annotate(err, "put firmware stable version").Err())
	}
	if err := dataSV.PutManyFaftStableVersion(ctx, m.faft); err != nil {
		merr = append(merr, errors.Annotate(err, "put firmware stable version").Err())
	}
	if len(merr) != 0 {
		logging.Errorf(ctx, "error writing stable versions: %s", merr)
		return nil, merr
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
