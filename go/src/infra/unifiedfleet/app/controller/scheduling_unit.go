// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/util"
)

// CreateSchedulingUnit creates a new SchedulingUnit in datastore.
func CreateSchedulingUnit(ctx context.Context, su *ufspb.SchedulingUnit) (*ufspb.SchedulingUnit, error) {
	f := func(ctx context.Context) error {
		if err := validateCreateSchedulingUnit(ctx, su); err != nil {
			return err
		}
		if _, err := inventory.BatchUpdateSchedulingUnits(ctx, []*ufspb.SchedulingUnit{su}); err != nil {
			return err
		}
		hc := &HistoryClient{}
		hc.logSchedulingUnitChanges(nil, su)
		return hc.SaveChangeEvents(ctx)
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		return nil, errors.Annotate(err, "unable to create SchedulingUnit %s", su.Name).Err()
	}
	return su, nil
}

// UpdateSchedulingUnit updates existing SchedulingUnit in datastore.
func UpdateSchedulingUnit(ctx context.Context, su *ufspb.SchedulingUnit, mask *field_mask.FieldMask) (*ufspb.SchedulingUnit, error) {
	f := func(ctx context.Context) error {
		// Get old/existing SchedulingUnit for logging and partial update.
		oldsu, err := inventory.GetSchedulingUnit(ctx, su.GetName())
		if err != nil {
			return err
		}
		// Validate the input.
		if err := validateUpdateSchedulingUnit(ctx, oldsu, su, mask); err != nil {
			return err
		}
		// Copy for logging.
		oldsuCopy := oldsu
		// Partial update by field mask.
		if mask != nil && len(mask.Paths) > 0 {
			// Validate partial update field mask.
			if err := validateSchedulingUnitUpdateMask(ctx, su, mask); err != nil {
				return err
			}
			// Clone oldsu for logging as the oldsu will be updated with new values.
			oldsuCopy = proto.Clone(oldsu).(*ufspb.SchedulingUnit)
			// Process the field mask to get updated values.
			su, err = processSchedulingUnitUpdateMask(ctx, oldsu, su, mask)
			if err != nil {
				return err
			}
		}
		if _, err := inventory.BatchUpdateSchedulingUnits(ctx, []*ufspb.SchedulingUnit{su}); err != nil {
			return err
		}
		hc := &HistoryClient{}
		hc.logSchedulingUnitChanges(oldsuCopy, su)
		return hc.SaveChangeEvents(ctx)
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		return nil, errors.Annotate(err, "failed to update SchedulingUnit %s in datastore", su.Name).Err()
	}
	return su, nil
}

// GetSchedulingUnit returns SchedulingUnit for the given id from datastore.
func GetSchedulingUnit(ctx context.Context, id string) (*ufspb.SchedulingUnit, error) {
	return inventory.GetSchedulingUnit(ctx, id)
}

// DeleteSchedulingUnit deletes the given SchedulingUnit in datastore.
func DeleteSchedulingUnit(ctx context.Context, id string) error {
	f := func(ctx context.Context) error {
		// Get the SchedulingUnit for logging.
		su, err := inventory.GetSchedulingUnit(ctx, id)
		if err != nil {
			return err
		}
		if err := inventory.DeleteSchedulingUnit(ctx, id); err != nil {
			return err
		}
		hc := &HistoryClient{}
		hc.logSchedulingUnitChanges(su, nil)
		return hc.SaveChangeEvents(ctx)
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		return errors.Annotate(err, "failed to delete SchedulingUnit %s in datastore", id).Err()
	}
	return nil
}

// ListSchedulingUnits lists the SchedulingUnits in datastore.
func ListSchedulingUnits(ctx context.Context, pageSize int32, pageToken, filter string, keysOnly bool) ([]*ufspb.SchedulingUnit, string, error) {
	var filterMap map[string][]interface{}
	var err error
	if filter != "" {
		filterMap, err = getFilterMap(filter, inventory.GetSchedulingUnitIndexedFieldName)
		if err != nil {
			return nil, "", errors.Annotate(err, "failed to read filter for listing SchedulingUnits").Err()
		}
	}
	filterMap = resetSchedulingUnitTypeFilter(filterMap)
	return inventory.ListSchedulingUnits(ctx, pageSize, pageToken, filterMap, keysOnly)
}

// validateCreateSchedulingUnit validates if a SchedulingUnit can be created.
func validateCreateSchedulingUnit(ctx context.Context, su *ufspb.SchedulingUnit) error {
	// Check if SchedulingUnit already exists.
	if err := resourceAlreadyExists(ctx, []*Resource{GetSchedulingUnitResource(su.Name)}, nil); err != nil {
		return err
	}
	// Check if the DUTs/MachineLSEs not found.
	if err := checkIfMachineLSEsExists(ctx, su.GetMachineLSEs()); err != nil {
		return err
	}
	// Check if DUTs/MachineLSEs already used in other SchedulingUnit or specified more than once
	seenDuts := make(map[string]bool)
	for _, lse := range su.GetMachineLSEs() {
		if seenDuts[lse] {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("DUT %s was specified more than once", lse))
		}
		seenDuts[lse] = true

		schedulingUnits, err := inventory.QuerySchedulingUnitByPropertyNames(ctx, map[string]string{"machinelses": lse}, true)
		if err != nil {
			return errors.Annotate(err, "failed to query SchedulingUnit for machinelses %s", lse).Err()
		}
		if len(schedulingUnits) > 0 {
			return status.Errorf(codes.FailedPrecondition, fmt.Sprintf("DUT %s is already associated with SchedulingUnit %s.", lse, schedulingUnits[0].GetName()))
		}
	}
	return nil
}

// validateUpdateSchedulingUnit validates if an existing SchedulingUnit can be updated.
func validateUpdateSchedulingUnit(ctx context.Context, oldsu *ufspb.SchedulingUnit, su *ufspb.SchedulingUnit, mask *field_mask.FieldMask) error {
	// Check if resources does not exist.
	if err := ResourceExist(ctx, []*Resource{GetSchedulingUnitResource(su.Name)}, nil); err != nil {
		return err
	}
	// Check if the DUTs/MachineLSEs not found.
	if err := checkIfMachineLSEsExists(ctx, su.GetMachineLSEs()); err != nil {
		return err
	}
	// Check if DUTs/MachineLSEs already used in other SchedulingUnit or specified more than once
	seenDuts := make(map[string]bool)
	for _, lse := range su.GetMachineLSEs() {
		if seenDuts[lse] {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("DUT %s was specified more than once", lse))
		}
		seenDuts[lse] = true

		schedulingUnits, err := inventory.QuerySchedulingUnitByPropertyNames(ctx, map[string]string{"machinelses": lse}, true)
		if err != nil {
			return errors.Annotate(err, "failed to query SchedulingUnit for machinelses %s", lse).Err()
		}
		for _, schedulingUnit := range schedulingUnits {
			if schedulingUnit.GetName() != su.GetName() {
				return status.Errorf(codes.FailedPrecondition, fmt.Sprintf("DUT %s is already associated with SchedulingUnit %s.", lse, schedulingUnit.GetName()))
			}
		}
	}
	return nil
}

// processSchedulingUnitUpdateMask processes update field mask to get only specific update
// fields and return a complete SchedulingUnit object with updated and existing fields.
func processSchedulingUnitUpdateMask(ctx context.Context, oldSu *ufspb.SchedulingUnit, su *ufspb.SchedulingUnit, mask *field_mask.FieldMask) (*ufspb.SchedulingUnit, error) {
	// Update the fields in the existing/old SchedulingUnit.
	for _, path := range mask.Paths {
		switch path {
		case "pools":
			oldSu.Pools = mergeTags(oldSu.GetPools(), su.GetPools())
		case "pools.remove":
			oldPools := oldSu.GetPools()
			for _, lse := range su.GetPools() {
				oldPools = util.RemoveStringEntry(oldPools, lse)
			}
			oldSu.Pools = oldPools
		case "machinelses":
			oldSu.MachineLSEs = mergeTags(oldSu.GetMachineLSEs(), su.GetMachineLSEs())
		case "machinelses.remove":
			oldMachineLSEs := oldSu.GetMachineLSEs()
			for _, lse := range su.GetMachineLSEs() {
				oldMachineLSEs = util.RemoveStringEntry(oldMachineLSEs, lse)
			}
			oldSu.MachineLSEs = oldMachineLSEs
		case "tags":
			oldSu.Tags = mergeTags(oldSu.GetTags(), su.GetTags())
		case "tags.remove":
			oldTags := oldSu.GetTags()
			for _, lse := range su.GetTags() {
				oldTags = util.RemoveStringEntry(oldTags, lse)
			}
			oldSu.Tags = oldTags
		case "type":
			oldSu.Type = su.GetType()
		case "description":
			oldSu.Description = su.GetDescription()
		case "primary-dut":
			oldSu.PrimaryDut = su.GetPrimaryDut()
		case "expose-type":
			oldSu.ExposeType = su.GetExposeType()
		}
	}
	if oldSu.GetPrimaryDut() != "" {
		// Check primary dut exists in SU machinelses
		hasPrimaryDut := false
		for _, dut := range oldSu.GetMachineLSEs() {
			if dut == oldSu.GetPrimaryDut() {
				hasPrimaryDut = true
				break
			}
		}
		if !hasPrimaryDut {
			return oldSu, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("Primary dut %s is associated with SchedulingUnit's machinelses %s.", oldSu.GetPrimaryDut(), oldSu.GetMachineLSEs()))
		}
	}
	// Return existing/old SchedulingUnit with new updated values.
	return oldSu, nil
}

// validateSchedulingUnitUpdateMask validates the update mask for SchedulingUnit partial update.
func validateSchedulingUnitUpdateMask(ctx context.Context, su *ufspb.SchedulingUnit, mask *field_mask.FieldMask) error {
	if mask != nil {
		// Validate the give field mask.
		for _, path := range mask.Paths {
			switch path {
			case "name":
				return status.Error(codes.InvalidArgument, "name cannot be updated, delete and create a SchedulingUnit instead")
			case "update_time":
				return status.Error(codes.InvalidArgument, "update_time cannot be updated, it is a output only field")
			case "pools":
			case "pools.remove":
			case "tags":
			case "tags.remove":
			case "type":
			case "machinelses":
			case "machinelses.remove":
			case "description":
			case "primary-dut":
			case "expose-type":
				// Valid fields, nothing to validate.
			default:
				return status.Errorf(codes.InvalidArgument, "unsupported update mask path %q", path)
			}
		}
	}
	return nil
}

func checkIfMachineLSEsExists(ctx context.Context, lseNames []string) error {
	var resourcesNotfound []*Resource
	for _, lseName := range lseNames {
		resourcesNotfound = append(resourcesNotfound, GetMachineLSEResource(lseName))
	}
	if err := ResourceExist(ctx, resourcesNotfound, nil); err != nil {
		return err
	}
	return nil
}
