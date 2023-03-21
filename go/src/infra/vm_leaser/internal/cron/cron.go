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
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/runtime/paniccatcher"
	"go.chromium.org/luci/server"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"

	"infra/vm_leaser/internal/frontend"
)

// Default VM Leaser cron parameters
const (
	// TODO(justinsuen): See isInstanceExpired comment.
	//
	// Filter for listing expired instances
	expiredVMFilter string = "(name eq ^vm-.*) (status eq RUNNING)"
)

// waitFunc takes a list of GCE operations and executes them.
//
// waitFunc blocks until the operations are complete, polling regularly and
// returns the statuses.
var waitFunc = func(ctx context.Context, ops []*compute.Operation) error {
	var resultErr *multierror.Error
	for _, op := range ops {
		if err := op.Wait(ctx); err != nil {
			resultErr = multierror.Append(resultErr, err)
		}
	}
	return resultErr.ErrorOrNil()
}
var wait = waitFunc

// RegisterCronServer initializes the VM Leaser cron server.
func RegisterCronServer(srv *server.Server) {
	srv.RunInBackground("vm_leaser.cron", func(ctx context.Context) {
		// releaseExpiredVMs every five minutes. GCP takes about 2 minutes to kill
		// instances.
		Run(ctx, 5*time.Minute, releaseExpiredVMs)
	})
}

// Run runs f repeatedly, until the context is cancelled.
//
// This method runs f based on minInterval time interval.
func Run(ctx context.Context, minInterval time.Duration, f func(context.Context) error) {
	defer logging.Warningf(ctx, "Exiting cron")

	// call calls the provided cron method f
	//
	// If call catches a panic, the cron run will stop once the whole context is
	// cancelled.
	call := func(ctx context.Context) error {
		defer paniccatcher.Catch(func(p *paniccatcher.Panic) {
			logging.Errorf(ctx, "Caught panic: %s\n%s", p.Reason, p.Stack)
		})
		return f(ctx)
	}

	for {
		start := clock.Now(ctx)
		if err := call(ctx); err != nil {
			logging.Errorf(ctx, "Iteration failed: %s", err)
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
func releaseExpiredVMs(ctx context.Context) error {
	logging.Debugf(ctx, "Releasing expired VMs")
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient: %v", err)
	}
	defer instancesClient.Close()

	var ops []*compute.Operation
	var errors *multierror.Error

	it, err := listInstances(ctx, instancesClient, "chrome-fleet-vm-leaser-cr-exp", frontend.DefaultRegion)
	if err != nil {
		return err
	}

	// Iterate through each instance and check the expiry for deletion.
	for {
		instance, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		expired, err := isInstanceExpired(ctx, instance, time.Now().Unix())
		if err != nil {
			break
		}
		if expired {
			logging.Infof(ctx, "Scheduling %s for deletion.\n", instance.GetName(), instance.GetMetadata().GetItems())
			op, err := deleteInstance(ctx, instancesClient, instance.GetName(), "chrome-fleet-vm-leaser-cr-exp", frontend.DefaultRegion)
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("failed deleting VM instance %s: %v", instance.GetName(), err))
				continue
			}
			ops = append(ops, op)
		}
	}

	if err := wait(ctx, ops); err != nil {
		errors = multierror.Append(errors, fmt.Errorf("failed waiting for VM instance to be deleted: %v", err))
	}

	logging.Infof(ctx, "Done.")
	return errors.ErrorOrNil()
}

// listInstances lists filtered instances created in a project and zone.
func listInstances(ctx context.Context, c *compute.InstancesClient, project, zone string) (*compute.InstanceIterator, error) {
	req := &computepb.ListInstancesRequest{
		Project: project,
		Zone:    zone,
		Filter:  proto.String(expiredVMFilter),
	}
	return c.List(ctx, req), nil
}

// deleteInstance creates a delete operation for a given instance name.
func deleteInstance(ctx context.Context, c *compute.InstancesClient, instanceName, project, zone string) (*compute.Operation, error) {
	req := &computepb.DeleteInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: instanceName,
	}
	op, err := c.Delete(ctx, req)
	if err != nil {
		return nil, err
	}
	return op, nil
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
