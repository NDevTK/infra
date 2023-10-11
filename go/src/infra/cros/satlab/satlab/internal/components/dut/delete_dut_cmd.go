// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"fmt"
	"os"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	"infra/cros/satlab/common/satlabcommands"
	"infra/cros/satlab/common/services/ufs"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/executor"
	ufsModels "infra/unifiedfleet/api/v1/models"
	ufsApi "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

type deleteClient interface {
	DeleteAsset(context.Context, *ufsApi.DeleteAssetRequest, ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteMachineLSE(context.Context, *ufsApi.DeleteMachineLSERequest, ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteRack(context.Context, *ufsApi.DeleteRackRequest, ...grpc.CallOption) (*emptypb.Empty, error)
	GetMachineLSE(ctx context.Context, in *ufsApi.GetMachineLSERequest, opts ...grpc.CallOption) (*ufsModels.MachineLSE, error)
}

// DeleteDUTCmd is the implementation of the "satlab delete DUT" command.
var DeleteDUTCmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Delete a Satlab DUT",
	CommandRun: func() subcommands.CommandRun {
		c := &deleteDUT{}
		registerShivasFlags(c)
		c.Flags.BoolVar(&c.full, "full", false, "whether to use a full/cascading delete for DUTs")
		return c
	},
}

// DeleteDUT holds the arguments that are needed for the delete DUT command.
type deleteDUT struct {
	shivasDeleteDUT
	// Satlab-specific fields, if any exist, go here.
	full bool
}

// Run attempts to delete a DUT and returns an exit status.
func (c *deleteDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// InnerRun is the implementation of the delete command.
func (c *deleteDUT) innerRun(a subcommands.Application, positionalArgs []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	ctx = utils.SetupContext(ctx, c.envFlags.GetNamespace())

	if c.commonFlags.SatlabID == "" {
		var err error
		c.commonFlags.SatlabID, err = satlabcommands.GetDockerHostBoxIdentifier(ctx, &executor.ExecCommander{})
		if err != nil {
			return errors.Annotate(err, "get dut").Err()
		}
	}

	// No flags need to be annotated with the satlab prefix for delete dut.
	// However, the positional arguments need to have the satlab prefix
	// prepended.
	for i, item := range positionalArgs {
		positionalArgs[i] = site.MaybePrepend(site.Satlab, c.commonFlags.SatlabID, item)
	}

	ufs, err := ufs.NewUFSClient(ctx, c.envFlags.GetUFSService(), &c.authFlags)
	if err != nil {
		return errors.Annotate(err, "get dut").Err()
	}

	return innerRunWithClients(ctx, c, positionalArgs, ufs)
}

// innerRunWithClients is intended to contain testable business logic we can
// pass a custom client into.
func innerRunWithClients(ctx context.Context, c *deleteDUT, dutNames []string, ufs deleteClient) error {
	// fetch DUTs to print them. Purely cosmetic, so errors will be ignored.
	duts := getAllDuts(ctx, dutNames, ufs)
	fmt.Printf("\nDUT(s) before deletion:")
	printMachineLSEs(duts)

	pass, fail := deleteAllDuts(ctx, dutNames, ufs)
	fmt.Printf("\nSuccessfully deleted DUT(s):\n")
	fmt.Printf("%v\n", pass)
	fmt.Printf("\nFailed to delete DUT(s):\n")
	fmt.Printf("%v\n", fail)

	if c.full {
		// Delete all assets for DUTs. If the DUT still exists (due to a
		// failure when deleting), the DeleteAsset RPC will return an error,
		// so we can be relatively sloppy when finding which assets to delete.
		assetsToDelete := []string{}
		for _, dut := range duts {
			fmt.Printf("Attempting to delete assets: %v for dut: %s\n", dut.Machines, dut.Name)
			assetsToDelete = append(assetsToDelete, dut.Machines...)
		}

		pass, fail := deleteAllAssets(ctx, assetsToDelete, ufs)
		fmt.Printf("\nSuccessfully deleted Assets(s):\n")
		fmt.Printf("%v\n", pass)
		fmt.Printf("Failed to delete Assets(s):\n")
		fmt.Printf("%v\n\n", fail)

		// Delete all racks. Similarly, if a rack still has assets associated
		// with it, the RPC will fail, so we can give a best effort attempt and
		// tell the user the RPC failed if there is some issue.
		//
		// In theory this is just `satlab-<id>-rack`, but it's easy enough to
		// use the actual rack that `GetMachineLSE` reports.
		racksToDelete := []string{}
		for _, dut := range duts {
			fmt.Printf("Attempting to delete rack: %s for dut: %s\n", dut.Rack, dut.Name)
			racksToDelete = append(racksToDelete, dut.Rack)
		}
		pass, fail = deleteAllRacks(ctx, racksToDelete, ufs)
		fmt.Printf("\nSuccessfully deleted Racks(s):\n")
		fmt.Printf("%v\n", pass)
		fmt.Printf("Failed to delete Racks(s):\n")
		fmt.Printf("%v\n\n", fail)
	}

	return nil
}

// getAllDuts fetches all DUTs with name in names.
//
// Should eventually be replaced with BatchGet or ConcurrentGet methods but
// since the caller will only be using this with a low # of DUTs is acceptable
// for now.
func getAllDuts(ctx context.Context, names []string, ufs deleteClient) []*ufsModels.MachineLSE {
	machineLSEs := []*ufsModels.MachineLSE{}
	for _, n := range names {
		m, err := ufs.GetMachineLSE(ctx, &ufsApi.GetMachineLSERequest{
			Name: ufsUtil.AddPrefix(ufsUtil.MachineLSECollection, n),
		})
		if err != nil {
			fmt.Printf("error fetching DUT: %s", err)
		} else {
			machineLSEs = append(machineLSEs, m)
		}
	}

	return machineLSEs
}

// deleteAllDuts deletes all DUTs with certain names. Returns an two arrays
// with the names that have been deleted successfully and unsuccessfully.
func deleteAllDuts(ctx context.Context, names []string, ufs deleteClient) ([]string, []string) {
	success := []string{}
	fail := []string{}

	for _, dut := range names {
		_, err := ufs.DeleteMachineLSE(ctx, &ufsApi.DeleteMachineLSERequest{
			Name: ufsUtil.AddPrefix(ufsUtil.MachineLSECollection, dut),
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error()+" => "+dut)
			fail = append(fail, dut)
		} else {
			success = append(success, dut)
		}
	}

	return success, fail
}

// deleteAllAssets deletes all Assets with certain names. Returns an two arrays
// with the names that have been deleted successfully and unsuccessfully.
func deleteAllAssets(ctx context.Context, names []string, ufs deleteClient) ([]string, []string) {
	success := []string{}
	fail := []string{}

	for _, assetName := range names {
		_, err := ufs.DeleteAsset(ctx, &ufsApi.DeleteAssetRequest{
			Name: ufsUtil.AddPrefix(ufsUtil.AssetCollection, assetName),
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error()+" => "+assetName)
			fail = append(fail, assetName)
		} else {
			success = append(success, assetName)
		}
	}

	return success, fail
}

// deleteAllRacks deletes all Racks with certain names. Returns an two arrays
// with the names that have been deleted successfully and unsuccessfully.
func deleteAllRacks(ctx context.Context, names []string, ufs deleteClient) ([]string, []string) {
	success := []string{}
	fail := []string{}

	for _, rackName := range names {
		_, err := ufs.DeleteRack(ctx, &ufsApi.DeleteRackRequest{
			Name: ufsUtil.AddPrefix(ufsUtil.RackCollection, rackName),
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error()+" => "+rackName)
			fail = append(fail, rackName)
		} else {
			success = append(success, rackName)
		}
	}

	return success, fail
}

func printMachineLSEs(machineLSEs []*ufsModels.MachineLSE) {
	for i, m := range machineLSEs {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		PrintProtoJSON(m)
		if i < len(machineLSEs)-1 {
			fmt.Printf(",\n")
		}
	}
}

// PrintProtoJSON prints proto as a JSON object.
func PrintProtoJSON(pm proto.Message) {
	m := protojson.MarshalOptions{
		Indent:          "\t",
		EmitUnpopulated: true,
	}
	fmt.Print(m.Format(pm))
}
