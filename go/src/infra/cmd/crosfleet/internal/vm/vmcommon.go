// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vm

import (
	"context"
	"flag"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"go.chromium.org/chromiumos/config/go/test/api"

	croscommon "infra/cros/cmd/common_lib/common"
	"infra/vm_leaser/client"
)

// sanitizeForLabel replaces all unsupported characters with _ to be compatible
// with labels on GCP.
func sanitizeForLabel(str string) string {
	re := regexp.MustCompile(`[^a-z0-9-]`)
	return re.ReplaceAllString(str, "_")
}

// listLeases lists all active leases for the current user.
func listLeases(vmLeaser *client.Client, ctx context.Context) ([]*api.VM, error) {
	resp, err := vmLeaser.VMLeaserClient.ListLeases(ctx, &api.ListLeasesRequest{
		Parent: fmt.Sprintf("projects/%s", croscommon.GceProject),
		Filter: fmt.Sprintf("labels.client:crosfleet AND labels.leased-by:%s AND status:RUNNING", sanitizeForLabel(vmLeaser.Email)),
	})

	vms := resp.GetVms()
	for _, vm := range vms {
		// GceRegion is in full URL format https://www.googleapis.com/compute/v1/projects/chromeos-gce-tests/zones/us-west1-b
		splits := strings.Split(vm.GetGceRegion(), "/")
		vm.GceRegion = splits[len(splits)-1]
	}

	return resp.GetVms(), err
}

// printVMList pretty prints the list of VMs to a io.Writer.
func printVMList(vms []*api.VM, w io.Writer) {
	tw := tabwriter.NewWriter(w, 1, 1, 2, ' ', 0)
	fmt.Fprintln(tw, "Name\tZone\tIP Address\tSSH Port\tTime remaining\t")
	for _, vm := range vms {
		remainTime := ""
		if vm.GetExpirationTime() != nil {
			expiry := vm.GetExpirationTime().AsTime()
			remainTime = expiry.Sub(time.Now()).Round(time.Second).String()
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t\n", vm.GetId(), vm.GetGceRegion(), vm.GetAddress().GetHost(), vm.GetAddress().GetPort(), remainTime)
	}
	tw.Flush()
}

// envFlags contains parameters to config environment for "vm" subcommands.
type envFlags struct {
	env string
}

// Registers env flags.
func (c *envFlags) register(f *flag.FlagSet) {
	f.StringVar(&c.env, "env", "prod", "Environment of vm_leaser server. Choose from: prod, staging, local")
}

// getClientConfig returns vm_leaser client config based on flags.
func (c *envFlags) getClientConfig() (*client.Config, error) {
	switch c.env {
	case "prod":
		return client.ProdConfig(), nil
	case "staging":
		return client.StagingConfig(), nil
	case "local":
		return client.LocalConfig(), nil
	}
	return nil, fmt.Errorf("invalid environment: %s", c.env)
}
