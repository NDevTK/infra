// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"context"
	"fmt"
	"strconv"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	multierror "github.com/hashicorp/go-multierror"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/runtime/paniccatcher"
	"go.chromium.org/luci/server"

	"infra/vm_leaser/internal/constants"
	"infra/vm_leaser/internal/controller"
)

// Options are server options for the cron server
type Options struct {
	GcpProjects []string
	ServiceEnv  *string
}

// RegisterCronServer initializes the VM Leaser cron server.
func RegisterCronServer(srv *server.Server, opts Options) {
	for _, p := range opts.GcpProjects {
		projName := p // variable scoping
		cronName := fmt.Sprint("vm_leaser.vm_cleanup:", projName)
		srv.RunInBackground(cronName, func(ctx context.Context) {
			// releaseExpiredVMs every five minutes. GCP takes about 2 minutes to kill
			// instances.
			Run(ctx, 5*time.Minute, projName, releaseExpiredVMs)
		})
	}
}

// Run runs f repeatedly, until the context is cancelled.
//
// This method runs f based on minInterval time interval.
func Run(ctx context.Context, minInterval time.Duration, gcpProject string, f func(context.Context, string) error) {
	defer logging.Warningf(ctx, "exiting cron")

	// call calls the provided cron method f
	//
	// If call catches a panic, the cron run will stop once the whole context is
	// cancelled.
	call := func(ctx context.Context) error {
		defer paniccatcher.Catch(func(p *paniccatcher.Panic) {
			logging.Errorf(ctx, "caught panic: %s\n%s", p.Reason, p.Stack)
		})
		return f(ctx, gcpProject)
	}

	for {
		start := clock.Now(ctx)
		if err := call(ctx); err != nil {
			logging.Errorf(ctx, "iteration failed: %s", err)
		}

		// Ensure minInterval between iterations.
		if sleep := minInterval - clock.Since(ctx, start); sleep > 0 {
			select {
			case <-time.After(sleep):
			case <-ctx.Done():
				return
			}
		}
	}
}

// releaseExpiredVMs releases VMs based on their expiration times.
func releaseExpiredVMs(ctx context.Context, gcpProject string) error {
	logging.Debugf(ctx, "releaseExpiredVMs: releasing expired VMs for GCP project: %v", gcpProject)
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient: %v", err)
	}
	defer instancesClient.Close()

	var errors *multierror.Error

	allZones := []string{}
	for _, subZones := range constants.AllQuotaZones {
		for _, subZone := range subZones {
			allZones = append(allZones, subZone)
		}
	}

	// Loop through all quota zones for expired instances.
	for _, zone := range allZones {
		logging.Debugf(ctx, "releaseExpiredVMs: processing zone %v", zone)
		instances, err := controller.ListInstances(ctx, instancesClient, &api.ListLeasesRequest{
			Parent: fmt.Sprintf("projects/%s/zones/%s", gcpProject, zone),
		})
		if err != nil {
			return err
		}

		// Iterate through each instance and check the expiry for deletion.
		for _, in := range instances {
			expired, err := isInstanceExpired(ctx, in, time.Now().Unix())
			if err != nil {
				break
			}
			if expired {
				logging.Infof(ctx, "releaseExpiredVMs: scheduling %s in zone %s for deletion", in.GetName(), zone, in.GetMetadata().GetItems())
				err := controller.DeleteInstance(ctx, instancesClient, &api.ReleaseVMRequest{
					LeaseId:    in.GetName(),
					GceProject: gcpProject,
					GceRegion:  zone,
				})
				if err != nil {
					errors = multierror.Append(errors, fmt.Errorf("failed to schedule VM instance for deletion %s: %v", in.GetName(), err))
					continue
				}
			}
		}
	}

	logging.Infof(ctx, "done")
	return errors.ErrorOrNil()
}

// TODO(justinsuen): This implementation is a workaround since b/35164571 and
// b/120255780 blocks adding metadata filtering directly to the GCE list filter.
//
// isInstanceExpired checks the expiration time of an instance.
//
// isInstanceExpired manually gets the metadata fields of an instance and checks
// it against a deletion time. If the deletion time is greater, then that means
// the instance is expired.
func isInstanceExpired(ctx context.Context, instance *computepb.Instance, deletionTime int64) (bool, error) {
	var err error
	expirationTime := deletionTime
	for _, m := range instance.GetMetadata().GetItems() {
		if m.GetKey() == "expiration_time" {
			expirationTime, err = strconv.ParseInt(m.GetValue(), 10, 64)
			if err != nil {
				return false, fmt.Errorf("failed to convert expiration time: %v", err)
			}
			break
		}
	}

	// Expiration time must have been before the current time. If it was not set,
	// then it would equal deletion time here.
	if expirationTime < deletionTime {
		return true, nil
	}
	return false, nil
}
