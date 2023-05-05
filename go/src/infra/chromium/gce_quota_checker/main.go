// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/iterator"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
)

const USAGE = `
$ ./gce_quota_checker [--verbose] [--project some-project-ID] path/to/vms.cfg path/to/another/vms.cfg

This will sum up all deployments in the specified vms.cfg files for the
specified GCP project, then check those against the allowed quotas for the
project. This checks:
- VM instances per region
- instanced per VPC network
- in-use IPs per region
- CPUs per region (N1)
- N2 CPUs per region
- HDD per region
- SSD per region
- local SSD per region + family

Hints:
- region == "us-east1"
- zone == "us-east1-d"
- family == "n1", "n2", "e2", etc

NOTE: need to run 'gcloud auth application-default login' locally first.
`

func parseFlags() (string, bool, []string) {
	gcpProject := flag.String("project", "google.com:chromecompute", "ID of the project to get quota for.")
	isVerbose := flag.Bool("verbose", false, "Prints full quota usage.")
	flag.Usage = func() {
		fmt.Printf("%v\n", USAGE)
		os.Exit(1)
	}
	flag.Parse()
	paths := flag.Args()
	if len(paths) == 0 {
		flag.Usage()
	}
	return *gcpProject, *isVerbose, paths
}

func main() {
	ctx := context.Background()
	project, _, _ := parseFlags()

	var regions []string
	n2CpusPerRegionQuota := make(map[string]int32)
	cpusPerRegionQuota := make(map[string]int32)
	iPsPerRegionQuota := make(map[string]int32)
	instancesPerRegionQuota := make(map[string]int32)
	hddPerRegionQuota := make(map[string]int64)
	remoteSSDPerRegionQuota := make(map[string]int64)

	// Get regions and their quotas. Simply calling ListRegion will get
	// a variety of quota info for each region, which includes most of
	// what we care about.
	c, err := compute.NewRegionsRESTClient(ctx)
	if err != nil {
		log.Fatalln("Error init'ing gcloud client:", err)
	}
	defer c.Close()
	req := &computepb.ListRegionsRequest{
		Project: project,
	}
	it := c.List(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalln("Error listing regions:", err)
		}
		regions = append(regions, *resp.Name)
		for _, quota := range resp.Quotas {
			if *quota.Metric == "N2_CPUS" {
				n2CpusPerRegionQuota[*resp.Name] = int32(*quota.Limit)
			} else if *quota.Metric == "CPUS" {
				cpusPerRegionQuota[*resp.Name] = int32(*quota.Limit)
			} else if *quota.Metric == "IN_USE_ADDRESSES" {
				iPsPerRegionQuota[*resp.Name] = int32(*quota.Limit)
			} else if *quota.Metric == "INSTANCES" {
				instancesPerRegionQuota[*resp.Name] = int32(*quota.Limit)
			} else if *quota.Metric == "DISKS_TOTAL_GB" {
				hddPerRegionQuota[*resp.Name] = int64(*quota.Limit)
			} else if *quota.Metric == "SSD_TOTAL_GB" {
				remoteSSDPerRegionQuota[*resp.Name] = int64(*quota.Limit)
			}
		}
	}
	// TODO: Finish the rest.
}
