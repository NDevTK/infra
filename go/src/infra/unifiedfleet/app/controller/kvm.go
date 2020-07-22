// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"strings"

	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/proto"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"
	ufsUtil "infra/unifiedfleet/app/util"
)

// CreateKVM creates a new kvm in datastore.
func CreateKVM(ctx context.Context, kvm *ufspb.KVM, rackName string) (*ufspb.KVM, error) {
	// TODO(eshwarn): Add logic for Chrome OS
	f := func(ctx context.Context) error {
		// 1. Validate the input
		if err := validateCreateKVM(ctx, kvm, rackName); err != nil {
			return err
		}

		// 2. Get rack to associate the kvm
		rack, err := GetRack(ctx, rackName)
		if err != nil {
			return err
		}

		// 3. Update the rack with new kvm information
		if err = addKVMToRack(ctx, rack, kvm.Name); err != nil {
			return err
		}

		// 4. Create a kvm entry
		// we use this func as it is a non-atomic operation and can be used to
		// run within a transaction to make it atomic. Datastore doesnt allow
		// nested transactions.
		if _, err = registration.BatchUpdateKVMs(ctx, []*ufspb.KVM{kvm}); err != nil {
			return errors.Annotate(err, "Unable to create kvm %s", kvm.Name).Err()
		}
		return nil
	}

	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to create kvm in datastore: %s", err)
		return nil, err
	}
	return kvm, nil
}

// UpdateKVM updates kvm in datastore.
func UpdateKVM(ctx context.Context, kvm *ufspb.KVM, rackName string) (*ufspb.KVM, error) {
	// TODO(eshwarn): Add logic for Chrome OS
	f := func(ctx context.Context) error {
		// 1. Validate the input
		if err := validateUpdateKVM(ctx, kvm, rackName); err != nil {
			return err
		}

		if rackName != "" {
			// 2. Get the old rack associated with kvm
			oldRack, err := getRackForKVM(ctx, kvm.Name)
			if err != nil {
				return err
			}

			// User is trying to associate this kvm with a different rack.
			if oldRack.Name != rackName {
				// 3. Get rack to associate the kvm
				rack, err := GetRack(ctx, rackName)
				if err != nil {
					return err
				}

				// 4. Remove the association between old rack and this kvm.
				if err = removeKVMFromRacks(ctx, []*ufspb.Rack{oldRack}, kvm.Name); err != nil {
					return err
				}

				// 5. Update the rack with new kvm information
				if err = addKVMToRack(ctx, rack, kvm.Name); err != nil {
					return err
				}
			}
		}

		// 6. Update kvm entry
		// we use this func as it is a non-atomic operation and can be used to
		// run within a transaction. Datastore doesnt allow nested transactions.
		if _, err := registration.BatchUpdateKVMs(ctx, []*ufspb.KVM{kvm}); err != nil {
			return errors.Annotate(err, "Unable to update kvm %s", kvm.Name).Err()
		}
		return nil
	}

	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to update kvm in datastore: %s", err)
		return nil, err
	}
	return kvm, nil
}

// GetKVM returns kvm for the given id from datastore.
func GetKVM(ctx context.Context, id string) (*ufspb.KVM, error) {
	return registration.GetKVM(ctx, id)
}

// ListKVMs lists the kvms
func ListKVMs(ctx context.Context, pageSize int32, pageToken string) ([]*ufspb.KVM, string, error) {
	return registration.ListKVMs(ctx, pageSize, pageToken)
}

// DeleteKVM deletes the kvm in datastore
//
// For referential data intergrity,
// Delete if this KVM is not referenced by other resources in the datastore.
// If there are any references, delete will be rejected and an error will be returned.
func DeleteKVM(ctx context.Context, id string) error {
	err := validateDeleteKVM(ctx, id)
	if err != nil {
		return err
	}
	return registration.DeleteKVM(ctx, id)
}

// ReplaceKVM replaces an old KVM with new KVM in datastore
//
// It does a delete of old kvm and create of new KVM.
// All the steps are in done in a transaction to preserve consistency on failure.
// Before deleting the old KVM, it will get all the resources referencing
// the old KVM. It will update all the resources which were referencing
// the old KVM(got in the last step) with new KVM.
// Deletes the old KVM.
// Creates the new KVM.
// This will preserve data integrity in the system.
func ReplaceKVM(ctx context.Context, oldKVM *ufspb.KVM, newKVM *ufspb.KVM) (*ufspb.KVM, error) {
	// TODO(eshwarn) : implement replace after user testing the tool
	return nil, nil
}

// validateDeleteKVM validates if a KVM can be deleted
//
// Checks if this KVM(KVMID) is not referenced by other resources in the datastore.
// If there are any other references, delete will be rejected and an error will be returned.
func validateDeleteKVM(ctx context.Context, id string) error {
	machines, err := registration.QueryMachineByPropertyName(ctx, "kvm_id", id, true)
	if err != nil {
		return err
	}
	racks, err := registration.QueryRackByPropertyName(ctx, "kvm_ids", id, true)
	if err != nil {
		return err
	}
	racklses, err := inventory.QueryRackLSEByPropertyName(ctx, "kvm_ids", id, true)
	if err != nil {
		return err
	}
	if len(machines) > 0 || len(racks) > 0 || len(racklses) > 0 {
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("KVM %s cannot be deleted because there are other resources which are referring this KVM.", id))
		if len(machines) > 0 {
			errorMsg.WriteString(fmt.Sprintf("\nMachines referring the KVM:\n"))
			for _, machine := range machines {
				errorMsg.WriteString(machine.Name + ", ")
			}
		}
		if len(racks) > 0 {
			errorMsg.WriteString(fmt.Sprintf("\nRacks referring the KVM:\n"))
			for _, rack := range racks {
				errorMsg.WriteString(rack.Name + ", ")
			}
		}
		if len(racklses) > 0 {
			errorMsg.WriteString(fmt.Sprintf("\nRackLSEs referring the KVM:\n"))
			for _, racklse := range racklses {
				errorMsg.WriteString(racklse.Name + ", ")
			}
		}
		logging.Errorf(ctx, errorMsg.String())
		return status.Errorf(codes.FailedPrecondition, errorMsg.String())
	}
	return nil
}

// validateCreateKVM validates if a kvm can be created
//
// check if the kvm already exists
// check if the rack and resources referenced by kvm does not exist
func validateCreateKVM(ctx context.Context, kvm *ufspb.KVM, rackName string) error {
	// 1. Check if kvm already exists
	if err := resourceAlreadyExists(ctx, []*Resource{GetKVMResource(kvm.Name)}, nil); err != nil {
		return err
	}

	// Aggregate resource to check if rack does not exist
	resourcesNotFound := []*Resource{GetRackResource(rackName)}
	// Aggregate resource to check if resources referenced by the kvm does not exist
	if chromePlatformID := kvm.GetChromePlatform(); chromePlatformID != "" {
		resourcesNotFound = append(resourcesNotFound, GetChromePlatformResource(chromePlatformID))
	}
	// 2. Check if resources does not exist
	return ResourceExist(ctx, resourcesNotFound, nil)
}

// validateUpdateKVM validates if a kvm can be updated
//
// check if kvm, rack and resources referenced kvm does not exist
func validateUpdateKVM(ctx context.Context, kvm *ufspb.KVM, rackName string) error {
	// Aggregate resource to check if kvm does not exist
	resourcesNotFound := []*Resource{GetKVMResource(kvm.Name)}
	// Aggregate resource to check if rack does not exist
	if rackName != "" {
		resourcesNotFound = append(resourcesNotFound, GetRackResource(rackName))
	}
	// Aggregate resource to check if resources referenced by the kvm does not exist
	if chromePlatformID := kvm.GetChromePlatform(); chromePlatformID != "" {
		resourcesNotFound = append(resourcesNotFound, GetChromePlatformResource(chromePlatformID))
	}
	// Check if resources does not exist
	return ResourceExist(ctx, resourcesNotFound, nil)
}

// addKVMToRack adds the kvm info to the rack and updates
// the rack in datastore.
// Must be called within a transaction as BatchUpdateRacks is a non-atomic operation
func addKVMToRack(ctx context.Context, rack *ufspb.Rack, kvmName string) error {
	if rack == nil {
		return status.Errorf(codes.FailedPrecondition, "Rack is nil")
	}
	if rack.GetChromeBrowserRack() == nil {
		errorMsg := fmt.Sprintf("Rack %s is not a browser rack", rack.Name)
		return status.Errorf(codes.FailedPrecondition, errorMsg)
	}
	kvms := []string{kvmName}
	if rack.GetChromeBrowserRack().GetKvms() != nil {
		kvms = rack.GetChromeBrowserRack().GetKvms()
		kvms = append(kvms, kvmName)
	}
	rack.GetChromeBrowserRack().Kvms = kvms
	_, err := registration.BatchUpdateRacks(ctx, []*ufspb.Rack{rack})
	if err != nil {
		return errors.Annotate(err, "Unable to update rack %s with kvm %s information", rack.Name, kvmName).Err()
	}
	return nil
}

// getRackForKVM return rack associated with the kvm.
func getRackForKVM(ctx context.Context, kvmName string) (*ufspb.Rack, error) {
	racks, err := registration.QueryRackByPropertyName(ctx, "kvm_ids", kvmName, false)
	if err != nil {
		return nil, errors.Annotate(err, "Unable to query rack for kvm %s", kvmName).Err()
	}
	if racks == nil || len(racks) == 0 {
		errorMsg := fmt.Sprintf("No rack associated with the kvm %s. Data discrepancy error.\n", kvmName)
		return nil, status.Errorf(codes.Internal, errorMsg)
	}
	if len(racks) > 1 {
		errorMsg := fmt.Sprintf("More than one rack associated the kvm %s. Data discrepancy error.\n", kvmName)
		return nil, status.Errorf(codes.Internal, errorMsg)
	}
	return racks[0], nil
}

// removeKVMFromRacks removes the kvm info from racks and
// updates the racks in datastore.
// Must be called within a transaction as BatchUpdateRacks is a non-atomic operation
func removeKVMFromRacks(ctx context.Context, racks []*ufspb.Rack, id string) error {
	for _, rack := range racks {
		if rack.GetChromeBrowserRack() == nil {
			errorMsg := fmt.Sprintf("Rack %s is not a browser rack", rack.Name)
			return status.Errorf(codes.FailedPrecondition, errorMsg)
		}
		kvms := rack.GetChromeBrowserRack().GetKvms()
		kvms = ufsUtil.RemoveStringEntry(kvms, id)
		rack.GetChromeBrowserRack().Kvms = kvms
	}
	_, err := registration.BatchUpdateRacks(ctx, racks)
	if err != nil {
		return errors.Annotate(err, "Unable to remove kvm information %s from rack", id).Err()
	}
	return nil
}
