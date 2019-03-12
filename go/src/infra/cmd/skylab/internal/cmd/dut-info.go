// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/cmd/skylab/internal/site"
	"infra/libs/skylab/inventory"
)

// DutInfo subcommand: Get DUT inventory information
var DutInfo = &subcommands.Command{
	UsageLine: "dut-info [-json] [-short] HOSTNAME",
	ShortDesc: "print Device Under Test inventory information",
	LongDesc: `Print Device Under Test inventory information.

By default, information is printed in a human-readable format. This format may
change without prior notice.

If you need a stable output format, use -json, which dumps a JSON-encoded
serialization of the inventory.DeviceUnderTest protobuf.

The JSON-encoding for protobuf messages is described at
https://developers.google.com/protocol-buffers/docs/proto3#json

The protobuf definition of inventory.DeviceUnderTest is part of
https://chromium.googlesource.com/infra/infra/+/refs/heads/master/go/src/infra/
    libs/skylab/inventory/device.proto`,
	CommandRun: func() subcommands.CommandRun {
		c := &dutInfoRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)

		c.Flags.BoolVar(&c.asJSON, "json", false, "Print inventory as JSON-encoded protobuf.")
		c.Flags.BoolVar(&c.short, "short", false, `Print only the most frequently used information.
Incompatible with -json`)
		return c
	},
}

type dutInfoRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  envFlags

	asJSON bool
	short  bool
}

func (c *dutInfoRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		PrintError(a.GetErr(), errors.Annotate(err, "dut-info").Err())
		return 1
	}
	return 0
}

func (c *dutInfoRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if len(args) != 1 {
		return NewUsageError(c.Flags, "want 1 argument, have %d", len(args))
	}
	if c.short && c.asJSON {
		return NewUsageError(c.Flags, "can not use -short with -json")
	}

	ctx := cli.GetContext(a, c, env)
	hc, err := httpClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	e := c.envFlags.Env()
	ic := fleet.NewInventoryPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.AdminService,
		Options: site.DefaultPRPCOptions,
	})

	dut, err := getDutInfo(ctx, ic, args[0])
	if err != nil {
		return err
	}

	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()
	if c.asJSON {
		return printProtoJSON(bw, dut)
	}
	return printHumanizedInfo(bw, dut, c.short)
}

func getDutInfo(ctx context.Context, ic fleet.InventoryClient, hostname string) (*inventory.DeviceUnderTest, error) {
	resp, err := ic.GetDutInfo(ctx, &fleet.GetDutInfoRequest{Hostname: hostname})
	if err != nil {
		return nil, errors.Annotate(err, "get dutinfo for %s", hostname).Err()
	}
	var dut inventory.DeviceUnderTest
	if err := proto.Unmarshal(resp.Spec, &dut); err != nil {
		return nil, errors.Annotate(err, "get dutinfo for %s", hostname).Err()
	}
	return &dut, nil
}

// printHumanizedInfo prints dut in a human-readable format.
//
// If short is true, only the most-commonly used fields are printed.
func printHumanizedInfo(w io.Writer, dut *inventory.DeviceUnderTest, short bool) error {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	defer tw.Flush()

	c := dut.GetCommon()
	fmt.Fprintf(tw, "Hostname:\t%s\n", c.GetHostname())
	c.Hostname = nil
	fmt.Fprintf(tw, "Inventory Id:\t%s\n", c.GetId())
	c.Id = nil

	l := c.GetLabels()
	if l.GetModel() != "" {
		fmt.Fprintf(tw, "Model:\t%s\n", l.GetModel())
		l.Model = nil
	}
	if l.GetBoard() != "" {
		fmt.Fprintf(tw, "Board:\t%s\n", l.GetBoard())
		l.Board = nil
	}
	if l.GetReferenceDesign() != "" {
		fmt.Fprintf(tw, "ReferenceDesign:\t%s\n", l.GetReferenceDesign())
		l.ReferenceDesign = nil
	}
	if len(l.GetVariant()) > 0 {
		fmt.Fprintf(tw, "Variant:\t%s\n", strings.Join(l.GetVariant(), ", "))
		l.Variant = nil
	}

	c, sa := extractServoAttributes(c)
	if len(sa) > 0 {
		fmt.Fprintf(tw, "Servo attributes:\n")
		for k, v := range sa {
			fmt.Fprintf(tw, "\t%s\t%s\n", k, v)
		}
	}

	if short {
		return nil
	}

	fmt.Fprintf(tw, "\nOther inventory data:\n")
	// TODO(pprabhu) Use printProtoJSON once all protobuf fields have been made
	// optional. There is no clean way to skip printing Hostname and other
	// required fields currently.
	return printProtoText(tw, dut)
}

func printProtoText(w io.Writer, dut *inventory.DeviceUnderTest) error {
	return proto.MarshalText(w, dut)
}

func printProtoJSON(w io.Writer, dut *inventory.DeviceUnderTest) error {
	m := jsonpb.Marshaler{
		EnumsAsInts: false,
		Indent:      "\t",
	}
	return m.Marshal(w, dut)
}

var servoAttributeKeys = map[string]bool{
	"servo_host":   true,
	"servo_port":   true,
	"servo_serial": true,
}

func extractServoAttributes(c *inventory.CommonDeviceSpecs) (*inventory.CommonDeviceSpecs, map[string]string) {
	sa := make(map[string]string)
	others := make([]*inventory.KeyValue, 0, len(c.GetAttributes()))
	for _, kv := range c.GetAttributes() {
		if servoAttributeKeys[*kv.Key] {
			sa[*kv.Key] = *kv.Value
		} else {
			others = append(others, kv)
		}
	}
	c.Attributes = others
	return c, sa
}
