// Copyright 2022 The LUCI Authors.
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

package satlab

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/gae/service/datastore"

	"infra/appengine/crosskylabadmin/api/fleet/v1"

	// See https://bugs.chromium.org/p/chromium/issues/detail?id=1242998 for details.
	// TODO(gregorynisbet): Remove this once new behavior is default.
	_ "go.chromium.org/luci/gae/service/datastore/crbug1242998safeget"
)

const SatlabStableVersionKind = "satlab_stable_version"

// SatlabStableVersionEntry is a stable version entry.
type SatlabStableVersionEntry struct {
	_kind string `gae:"$kind,satlab_stable_version"`
	// ID has the following format. Earlier entries have precedence over later ones.
	//
	//   hostname     -- for a version scoped to a specific device.
	//   board|model  -- for a version scoped to just the board and the model.
	//
	ID string `gae:"$id"`
	// Base64Req stores a base64-encoded SetSatlabStableVersionRequest.
	// This request can be used to reconstitute this entire entry if required.
	Base64Req string `gae:"base64_req"`
	OS        string `gae:"os"`
	FW        string `gae:"fw"`
	FWImage   string `gae:"fw_image"`
}

// MakeSatlabStableVersionEntry creates a stable version entry from a stable version request.
func MakeSatlabStableVersionEntry(req *fleet.SetSatlabStableVersionRequest) (*SatlabStableVersionEntry, error) {
	if req == nil {
		return nil, errors.Reason("make satlab stable version: request cannot be nil").Err()
	}
	hostname := ""
	board := ""
	model := ""
	switch v := req.GetStrategy().(type) {
	case *fleet.SetSatlabStableVersionRequest_SatlabBoardAndModelStrategy:
		s := v.SatlabBoardAndModelStrategy
		board = s.GetBoard()
		model = s.GetModel()
	case *fleet.SetSatlabStableVersionRequest_SatlabHostnameStrategy:
		hostname = v.SatlabHostnameStrategy.GetHostname()
	}
	base64Req := ""
	bytes, err := proto.Marshal(req)
	if err == nil {
		base64Req = base64.StdEncoding.EncodeToString(bytes)
	}
	return &SatlabStableVersionEntry{
		ID:        makeSatlabStableVersionID(hostname, board, model),
		Base64Req: base64Req,
		OS:        req.GetCrosVersion(),
		FW:        req.GetFirmwareVersion(),
		FWImage:   req.GetFirmwareImage(),
	}, nil
}

// PutSatlabStableVersionEntry puts a single SatlabStableVersionEntry in datastore.
func PutSatlabStableVersionEntry(ctx context.Context, entry *SatlabStableVersionEntry) error {
	if err := datastore.Put(ctx, entry); err != nil {
		return errors.Annotate(err, "put satlab stable version entry").Err()
	}
	return nil
}

// GetSatlabStableVersionEntryByID uses the ID to look up a satlab stable entry.
func GetSatlabStableVersionEntryByID(ctx context.Context, req *fleet.GetStableVersionRequest) (*SatlabStableVersionEntry, error) {
	if req == nil {
		return nil, errors.Reason("get satlab stable version entry by id: request cannot be nil").Err()
	}
	id := makeSatlabStableVersionID(req.GetHostname(), req.GetBuildTarget(), req.GetModel())
	entry := &SatlabStableVersionEntry{ID: id}
	if err := datastore.Get(ctx, entry); err != nil {
		return nil, errors.Annotate(err, "get satlab stable version entry").Err()
	}
	return entry, nil
}

// MakeSatlabStableVersionID takes a hostname, board, and model and combines them into an ID.
func makeSatlabStableVersionID(hostname string, board string, model string) string {
	if hostname != "" {
		return hostname
	}
	if model != "" && board != "" {
		return fmt.Sprintf("%s|%s", board, model)
	}
	return ""
}
