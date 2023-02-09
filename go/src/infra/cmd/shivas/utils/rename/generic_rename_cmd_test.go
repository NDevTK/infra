// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package rename

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"

	"infra/cmd/shivas/site"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

func fakeRename(ctx context.Context, ic ufsAPI.FleetClient, name, newName string) (interface{}, error) {
	return &ufspb.Asset{}, nil
}

// printAsset prints the result of the operation
func fakePrint(asset proto.Message) {
	return
}

// TestGenGenericRenameCmdNamespace checks whether the generated command can
// fetch the appropriate value
func TestGenGenericRenameCmdNamespace(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		validNSList []string
		wantNS      string
		wantErr     bool
	}{
		{
			"invalid ns",
			ufsUtil.BrowserNamespace,
			site.OSLikeNamespaces,
			ufsUtil.BrowserNamespace,
			true,
		},
		{
			"valid ns",
			ufsUtil.OSNamespace,
			site.OSLikeNamespaces,
			ufsUtil.OSNamespace,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := GenGenericRenameCmd("fake", fakeRename, fakePrint, tt.validNSList, "partner-os")
			cmd := c.CommandRun().(*renameGeneric)

			// have to set this way since namespace is unexported from envFlags
			cmd.GetFlags().Set("namespace", tt.namespace)

			ns, err := cmd.getNamespace()
			if ns != tt.wantNS {
				t.Errorf("wrong namespace. expected: %s, got %s", tt.wantNS, ns)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("expected err: %t, got err: %t", tt.wantErr, (err != nil))
			}
		})
	}
}
