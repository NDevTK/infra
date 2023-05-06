// Copyright 2023 The Chromium Authors
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
$ ./gce_quota_checker [--project some-project-ID] path/to/vms.cfg path/to/another/vms.cfg

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

type quotaVals struct {
	max  int64
	used int64
}

type regionQuotas struct {
	// One-per instance quotas
	instancesQuota quotaVals
	ipsQuota       quotaVals

	// CPUs
	cpusQuota   quotaVals
	n2CpusQuota quotaVals

	// Disk
	hddQuota       quotaVals
	remoteSSDQuota quotaVals
	// GCE tracks this one across both region and VM family (n1, e2, etc).
	localSSDPerFamilyQuota map[string]*quotaVals
}

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

	quotasPerRegion := make(map[string]*regionQuotas)

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
		quotas := regionQuotas{localSSDPerFamilyQuota: make(map[string]*quotaVals)}
		for _, quota := range resp.Quotas {
			if *quota.Metric == "N2_CPUS" {
				quotas.n2CpusQuota = quotaVals{max: int64(*quota.Limit)}
			} else if *quota.Metric == "CPUS" {
				quotas.cpusQuota = quotaVals{max: int64(*quota.Limit)}
			} else if *quota.Metric == "IN_USE_ADDRESSES" {
				quotas.ipsQuota = quotaVals{max: int64(*quota.Limit)}
			} else if *quota.Metric == "INSTANCES" {
				quotas.instancesQuota = quotaVals{max: int64(*quota.Limit)}
			} else if *quota.Metric == "DISKS_TOTAL_GB" {
				quotas.hddQuota = quotaVals{max: int64(*quota.Limit)}
			} else if *quota.Metric == "SSD_TOTAL_GB" {
				quotas.remoteSSDQuota = quotaVals{max: int64(*quota.Limit)}
			}
		}
		quotasPerRegion[*resp.Name] = &quotas
	}

	// TODO: Finish the rest.
}
