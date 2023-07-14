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
	"infra/cros/satlab/satlab/internal/commands"
	"infra/cros/satlab/satlab/internal/components/ufs"
	"infra/cros/satlab/satlab/internal/site"
	ufsModels "infra/unifiedfleet/api/v1/models"
	ufsApi "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

type deleteClient interface {
	DeleteMachineLSE(context.Context, *ufsApi.DeleteMachineLSERequest, ...grpc.CallOption) (*emptypb.Empty, error)
	GetMachineLSE(ctx context.Context, in *ufsApi.GetMachineLSERequest, opts ...grpc.CallOption) (*ufsModels.MachineLSE, error)
}

// DeleteDUTCmd is the implementation of the "satlab delete DUT" command.
var DeleteDUTCmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Delete a Satlab DUT",
	CommandRun: func() subcommands.CommandRun {
		c := &deleteDUT{}
		registerShivasFlags(c)
		return c
	},
}

// DeleteDUT holds the arguments that are needed for the delete DUT command.
type deleteDUT struct {
	shivasDeleteDUT
	// Satlab-specific fields, if any exist, go here.
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
		c.commonFlags.SatlabID, err = commands.GetDockerHostBoxIdentifier()
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
