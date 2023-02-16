// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/maruel/subcommands"

	"infra/cmd/vmlab/internal/config"

	"infra/libs/vmlab"
	"infra/libs/vmlab/api"
)

var LeaseCmd = &subcommands.Command{
	UsageLine: "lease",
	ShortDesc: "lease a virtual machine from given inputs",
	CommandRun: func() subcommands.CommandRun {
		c := &leaseRun{}
		c.leaseFlags.register(&c.Flags)
		c.gcloudBackendFlags.register(&c.Flags)
		return c
	},
}

type leaseFlags struct {
	configName      string
	expireAfter     int
	swarmingBotName string
}

type gcloudBackendFlags struct {
	gceImageProject string
	gceImageName    string
}

func (c *leaseFlags) register(f *flag.FlagSet) {
	f.StringVar(&c.configName, "config", "", "Config name to use. cts-prototype is the only valid option at this momenet.")
	f.IntVar(&c.expireAfter, "expire-after", -1, "Created VM instance should expire and be destroyed after given seconds. Not all backends support this.")
	f.StringVar(&c.swarmingBotName, "swarming-bot-name", "", "Name of the swarming bot name. You can use the same same to do bulk cleanup for some backends.")
}

func (c *gcloudBackendFlags) register(f *flag.FlagSet) {
	f.StringVar(&c.gceImageProject, "gce-image-project", "", "GCE image project, example: betty-cloud-prototype")
	f.StringVar(&c.gceImageName, "gce-image-name", "", "GCE image name, example: betty-arc-r-release-r112-15350-0-0")
}

type leaseRun struct {
	subcommands.CommandRunBase
	leaseFlags
	gcloudBackendFlags
}

func (c *leaseRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if c.leaseFlags.configName == "" {
		fmt.Fprintln(os.Stderr, "Config name must be set.")
		return 1
	}
	createConfig, ok := config.ConfigMapping[c.leaseFlags.configName]
	if !ok {
		fmt.Fprintln(os.Stderr, "cannot find config.")
		return 1
	}
	if createConfig.ProviderId == api.ProviderId_GCLOUD {
		if c.gcloudBackendFlags.gceImageProject == "" {
			fmt.Fprintln(os.Stderr, "gce-image-project must be set.")
		}
		if c.gcloudBackendFlags.gceImageName == "" {
			fmt.Fprintln(os.Stderr, "gce-image-name must be set.")
		}
	}
	tags := map[string]string{}
	if c.swarmingBotName != "" {
		tags["swarming-bot-name"] = c.swarmingBotName
	}
	ins, err := vmlab.NewInstanceApi(createConfig.ProviderId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create instance: %v", err)
	}
	createdInstance, err := ins.Create(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create instance: %v", err)
	}
	fmt.Printf("%v", tags)
	fmt.Printf("%v", createdInstance)
	return 0
}
