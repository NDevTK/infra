// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package rename

import (
	"context"
	"testing"

	"infra/cmd/shivas/site"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"

	"github.com/golang/protobuf/proto"
	"github.com/maruel/subcommands"
)

func fakeRename(ctx context.Context, ic ufsAPI.FleetClient, name, newName string) (interface{}, error) {
	return &ufspb.Asset{}, nil
}

// printAsset prints the result of the operation
func fakePrint(asset proto.Message) {
	return
}

// TestGenGenericRenameCmdNamespace checks whether the validNSList correctly
// causes errors during execution of the command.
func TestGenGenericRenameCmdNamespace(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		validNSList []string
		wantErr     bool
	}{
		{
			"invalid ns",
			"browser",
			site.OSLikeNamespaces,
			true,
		},
		{
			"valid ns",
			"os",
			site.OSLikeNamespaces,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := GenGenericRenameCmd("fake", fakeRename, fakePrint, tt.validNSList, "")
			cmd := c.CommandRun().(*renameGeneric)
			cmd.name = "test"
			cmd.newName = "test"

			// have to set this way since namespace is unexported from envFlags
			cmd.GetFlags().Lookup("namespace").Value.Set(tt.namespace)

			ret_val := cmd.Run(&subcommands.DefaultApplication{}, []string{}, make(subcommands.Env))
			// we could also parse the error but this is less brittle and we
			// really just care about the presence of an error which directly
			// corresponds with the return value of running the cmd
			if (ret_val == 1) != tt.wantErr {
				t.Errorf("Got error: %t, want error: %t", ret_val != 0, tt.wantErr)
			}
		})
	}
}
