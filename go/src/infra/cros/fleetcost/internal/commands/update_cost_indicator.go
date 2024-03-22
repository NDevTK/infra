// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/maruel/subcommands"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	prpc "go.chromium.org/luci/grpc/prpc"

	"infra/cmdsupport/cmdlib"
	fleetcostpb "infra/cros/fleetcost/api/models"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/site"
	"infra/cros/fleetcost/internal/utils"
)

var UpdateCostIndicatorCommand *subcommands.Command = &subcommands.Command{
	UsageLine: "update-ci [options...]",
	ShortDesc: "update a cost indicator",
	LongDesc:  "Update a cost indicator",
	CommandRun: func() subcommands.CommandRun {
		c := &updateCostIndicatorCommand{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.authFlags.RegisterIDTokenFlags(&c.Flags)
		c.commonFlags.Register(&c.Flags)
		c.Flags.StringVar(&c.name, "name", "", "name of cost indicator")
		c.Flags.Func("type", "name of cost indicator", func(name string) error {
			typ, err := utils.ToIndicatorType(name)
			if err != nil {
				return errors.Reason("type %s is invalid", name).Err()
			}
			c.typ = typ
			return nil
		})
		c.Flags.StringVar(&c.board, "board", "", "board")
		c.Flags.StringVar(&c.model, "model", "", "model")
		c.Flags.StringVar(&c.sku, "sku", "", "sku")
		c.Flags.Func("cost", "cost", func(value string) error {
			cost, err := utils.ToUSD(value)
			if err != nil {
				return errors.Reason("cost %q is invalid", value).Err()
			}
			c.cost = cost
			return nil
		})
		c.Flags.Func("cadence", "cost-cadence", func(value string) error {
			costCadence, err := utils.ToCostCadence(value)
			if err != nil {
				return errors.Reason("cost cadence %q is invalid", value).Err()
			}
			c.costCadence = costCadence
			return nil
		})
		c.Flags.Float64Var(&c.burnoutRate, "burnout", math.NaN(), "device burnout rate")
		c.Flags.Func("location", "where the device is located", func(value string) error {
			location, err := utils.ToLocation(value)
			if err != nil {
				return errors.Reason("location %q is invalid", value).Err()
			}
			c.location = location
			return nil
		})
		return c
	},
}

type updateCostIndicatorCommand struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	commonFlags site.CommonFlags
	name        string
	typ         fleetcostpb.IndicatorType
	board       string
	model       string
	sku         string
	cost        *money.Money
	costCadence fleetcostpb.CostCadence
	burnoutRate float64
	location    fleetcostpb.Location
	description string
}

// Run is the main entrypoint to update-ci.
func (c *updateCostIndicatorCommand) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if err := c.innerRun(ctx, a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *updateCostIndicatorCommand) getFieldMaskPaths() []string {
	var out []string
	if c.name != "" {
		out = append(out, "name")
	}
	if int(c.typ) != 0 {
		out = append(out, "type")
	}
	if c.board != "" {
		out = append(out, "board")
	}
	if c.model != "" {
		out = append(out, "model")
	}
	if c.cost != nil {
		out = append(out, "cost")
	}
	if int(c.costCadence) != 0 {
		out = append(out, "cost_cadence")
	}
	if c.burnoutRate != 0 {
		out = append(out, "burnout_rate")
	}
	if int(c.location) != 0 {
		out = append(out, "location")
	}
	if c.description != "" {
		out = append(out, "description")
	}
	return out
}

func (c *updateCostIndicatorCommand) innerRun(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	host, err := c.commonFlags.Host()
	if err != nil {
		return err
	}
	var httpClient *http.Client
	if !c.commonFlags.HTTP() {
		var err error
		httpClient, err = getSecureClient(ctx, host, c.authFlags)
		if err != nil {
			return err
		}
	}
	prpcClient := &prpc.Client{
		C:    httpClient,
		Host: host,
		Options: &prpc.Options{
			Insecure:      c.commonFlags.HTTP(),
			PerRPCTimeout: 30 * time.Second,
		},
	}
	fleetCostClient := fleetcostAPI.NewFleetCostPRPCClient(prpcClient)
	resp, err := fleetCostClient.UpdateCostIndicator(ctx, &fleetcostAPI.UpdateCostIndicatorRequest{
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:        c.name,
			Type:        c.typ,
			Board:       c.board,
			Model:       c.model,
			Cost:        c.cost,
			CostCadence: c.costCadence,
			BurnoutRate: c.burnoutRate,
			Location:    c.location,
			Description: c.description,
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: c.getFieldMaskPaths(),
		},
	})
	fmt.Printf("%#v\n", resp)
	return err
}