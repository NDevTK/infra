// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/internal/app/model/inventory"
	ufsUtil "infra/unifiedfleet/util"
)

// GetMachineLSEDeployment returns the deployment record for the given id.
func GetMachineLSEDeployment(ctx context.Context, id string) (*ufspb.MachineLSEDeployment, error) {
	return inventory.GetMachineLSEDeployment(ctx, id)
}

// BatchGetMachineLSEDeployments returns a batch of deployment records.
func BatchGetMachineLSEDeployments(ctx context.Context, ids []string) ([]*ufspb.MachineLSEDeployment, error) {
	return inventory.BatchGetMachineLSEDeployments(ctx, ids)
}

// ListMachineLSEDeployments returns a batch of deployment records by filters
func ListMachineLSEDeployments(ctx context.Context, pageSize int32, pageToken, filter string, keysOnly bool) ([]*ufspb.MachineLSEDeployment, string, error) {
	var filterMap map[string][]interface{}
	var err error
	if filter != "" {
		filterMap, err = getFilterMap(filter, inventory.GetDeploymentIndexedFieldName)
		if err != nil {
			return nil, "", errors.Annotate(err, "Failed to read filter for listing deployment records").Err()
		}
	}
	return inventory.ListMachineLSEDeployments(ctx, pageSize, pageToken, filterMap, keysOnly)
}

// UpdateMachineLSEDeployment updates a machine lse deployment to datastore
func UpdateMachineLSEDeployment(ctx context.Context, dr *ufspb.MachineLSEDeployment, mask *field_mask.FieldMask) (*ufspb.MachineLSEDeployment, error) {
	f := func(ctx context.Context) error {
		hc := &HistoryClient{}

		// Get old/existing deployment record for logging and partial update.
		resp, err := inventory.GetMachineLSEDeployment(ctx, dr.GetSerialNumber())
		if err != nil {
			logging.Infof(ctx, "no existing deployment record for serial number %s, continue update", dr.GetSerialNumber())
		}
		var oldDr *ufspb.MachineLSEDeployment
		var oldDrCopy *ufspb.MachineLSEDeployment
		if resp != nil {
			oldDr = resp
			oldDrCopy = proto.Clone(oldDr).(*ufspb.MachineLSEDeployment)
		}

		// Partial update by field mask.
		if oldDr != nil && mask != nil && len(mask.Paths) > 0 {
			if err := validateDeploymentUpdateMask(mask); err != nil {
				return err
			}
			// Process the field mask to get updated values.
			dr, err = processDeploymentUpdateMask(ctx, oldDr, dr, mask)
			if err != nil {
				return errors.Annotate(err, "processing update mask failed").Err()
			}
		}
		if dr.GetHostname() == "" {
			dr.Hostname = ufsUtil.GetHostnameWithNoHostPrefix(dr.GetSerialNumber())
		}

		logging.Infof(ctx, "The deployment record to update is %#v", dr)
		if _, err := inventory.UpdateMachineLSEDeployments(ctx, []*ufspb.MachineLSEDeployment{dr}); err != nil {
			return errors.Annotate(err, "unable to update new deployment record: %s (%s)", dr.GetHostname(), dr.GetSerialNumber()).Err()
		}
		hc.LogMachineLSEDeploymentChanges(oldDrCopy, dr)
		return hc.SaveChangeEvents(ctx)
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		return nil, errors.Annotate(err, "UpdateMachineLSEDeployment").Err()
	}
	logging.Infof(ctx, "Successfully update deployment record serial number %q (%q)", dr.GetSerialNumber(), dr.GetDeploymentIdentifier())
	return dr, nil
}

// processDeploymentUpdateMask processes update field mask to get only specific update
func processDeploymentUpdateMask(ctx context.Context, oldCs *ufspb.MachineLSEDeployment, cs *ufspb.MachineLSEDeployment, mask *field_mask.FieldMask) (*ufspb.MachineLSEDeployment, error) {
	// Update the fields in the existing/old object
	for _, path := range mask.Paths {
		switch path {
		case "hostname":
			oldCs.Hostname = cs.Hostname
		case "deployment_identifier":
			oldCs.DeploymentIdentifier = cs.GetDeploymentIdentifier()
		case "deployment_env":
			oldCs.DeploymentEnv = cs.GetDeploymentEnv()
		case "configs_to_push":
			oldCs.ConfigsToPush = cs.GetConfigsToPush()
		}
	}
	// Return existing/old with new updated values.
	return oldCs, nil
}

// validateDeploymentUpdateMask validates the update mask for deployment record partial update.
func validateDeploymentUpdateMask(mask *field_mask.FieldMask) error {
	if mask != nil {
		// Validate the give field mask.
		for _, path := range mask.Paths {
			switch path {
			case "serial_number":
				return status.Error(codes.InvalidArgument, "serial number cannot be updated")
			case "hostname":
			case "deployment_identifier":
			case "deployment_env":
			case "configs_to_push":
				// Valid fields, nothing to validate.
			default:
				return status.Errorf(codes.InvalidArgument, "unsupported update mask path %q", path)
			}
		}
	}
	return nil
}
