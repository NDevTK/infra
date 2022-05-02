// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/model/registration"
	ufsUtil "infra/unifiedfleet/app/util"
)

// CreateKVM creates a new kvm in datastore.
func CreateKVM(ctx context.Context, kvm *ufspb.KVM) (*ufspb.KVM, error) {
	// TODO(eshwarn): Add logic for Chrome OS
	f := func(ctx context.Context) error {
		hc := getKVMHistoryClient(kvm)
		hc.LogKVMChanges(nil, kvm)

		// Get rack to associate the kvm
		rack, err := GetRack(ctx, kvm.GetRack())
		if err != nil {
			return err
		}

		// Validate the input
		if err := validateCreateKVM(ctx, kvm, rack); err != nil {
			return err
		}

		// Fill the zone to kvm OUTPUT only fields for indexing
		kvm.Zone = rack.GetLocation().GetZone().String()
		kvm.ResourceState = ufspb.State_STATE_REGISTERED

		// Create a kvm entry
		// we use this func as it is a non-atomic operation and can be used to
		// run within a transaction to make it atomic. Datastore doesnt allow
		// nested transactions.
		if _, err = registration.BatchUpdateKVMs(ctx, []*ufspb.KVM{kvm}); err != nil {
			return errors.Annotate(err, "Unable to create kvm %s", kvm.Name).Err()
		}

		// Update state
		if err := hc.stUdt.updateStateHelper(ctx, ufspb.State_STATE_REGISTERED); err != nil {
			return err
		}
		return hc.SaveChangeEvents(ctx)
	}

	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to create kvm in datastore: %s", err)
		return nil, err
	}
	return kvm, nil
}

// UpdateKVM updates kvm in datastore.
func UpdateKVM(ctx context.Context, kvm *ufspb.KVM, mask *field_mask.FieldMask) (*ufspb.KVM, error) {
	// TODO(eshwarn): Add logic for Chrome OS
	f := func(ctx context.Context) error {
		hc := getKVMHistoryClient(kvm)

		// Get old/existing KVM
		oldKVM, err := registration.GetKVM(ctx, kvm.GetName())
		if err != nil {
			return errors.Annotate(err, "UpdateKVM - get kvm %s failed", kvm.GetName()).Err()
		}

		// Validate the input
		if err := validateUpdateKVM(ctx, oldKVM, kvm, mask); err != nil {
			return errors.Annotate(err, "UpdateKVM - validation failed").Err()
		}

		// Copy for logging
		oldKVMCopy := proto.Clone(oldKVM).(*ufspb.KVM)
		// Fill the zone to kvm OUTPUT only fields
		kvm.Zone = oldKVM.GetZone()

		// Partial update by field mask
		if mask != nil && len(mask.Paths) > 0 {
			kvm, err = processKVMUpdateMask(ctx, oldKVM, kvm, mask)
			if err != nil {
				return errors.Annotate(err, "UpdateKVM - processing update mask failed").Err()
			}
		} else {
			// This is for complete object input
			if kvm.GetRack() == "" {
				return status.Error(codes.InvalidArgument, "rack cannot be empty for updating a KVM")
			}
			if oldKVM.GetRack() != kvm.GetRack() {
				// User is trying to associate this kvm with a different rack.
				// Get rack to associate the kvm
				rack, err := GetRack(ctx, kvm.GetRack())
				if err != nil {
					return errors.Annotate(err, "UpdateKVM - get rack %s failed", kvm.GetRack()).Err()
				}

				// check permission for the new rack realm
				if err := ufsUtil.CheckPermission(ctx, ufsUtil.RegistrationsUpdate, rack.GetRealm()); err != nil {
					return err
				}
				// Fill the zone to kvm OUTPUT only fields
				kvm.Zone = rack.GetLocation().GetZone().String()
			}
		}

		// Update state
		if err := hc.stUdt.updateStateHelper(ctx, kvm.GetResourceState()); err != nil {
			return errors.Annotate(err, "Fail to update state to kvm %s", kvm.GetName()).Err()
		}

		// Update kvm entry
		// we use this func as it is a non-atomic operation and can be used to
		// run within a transaction. Datastore doesnt allow nested transactions.
		if _, err := registration.BatchUpdateKVMs(ctx, []*ufspb.KVM{kvm}); err != nil {
			return errors.Annotate(err, "UpdateKVM - unable to batch update kvm %s", kvm.Name).Err()
		}
		hc.LogKVMChanges(oldKVMCopy, kvm)
		return hc.SaveChangeEvents(ctx)
	}

	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		return nil, errors.Annotate(err, "UpdateKVM - failed to update kvm %s in datastore", kvm.Name).Err()
	}
	return kvm, nil
}

// processKVMUpdateMask process update field mask to get only specific update
// fields and return a complete kvm object with updated and existing fields
func processKVMUpdateMask(ctx context.Context, oldKVM *ufspb.KVM, kvm *ufspb.KVM, mask *field_mask.FieldMask) (*ufspb.KVM, error) {
	// update the fields in the existing/old kvm
	for _, path := range mask.Paths {
		switch path {
		case "rack":
			if oldKVM.GetRack() != kvm.GetRack() {
				// User is trying to associate this kvm with a different rack.
				// Get rack to associate the kvm
				rack, err := GetRack(ctx, kvm.GetRack())
				if err != nil {
					return oldKVM, errors.Annotate(err, "UpdateKVM - get rack %s failed", kvm.GetRack()).Err()
				}
				// check permission for the new rack realm
				if err := ufsUtil.CheckPermission(ctx, ufsUtil.RegistrationsUpdate, rack.GetRealm()); err != nil {
					return oldKVM, err
				}
				oldKVM.Rack = kvm.GetRack()
				// Fill the zone to kvm OUTPUT only fields
				oldKVM.Zone = rack.GetLocation().GetZone().String()
			}
		case "resourceState":
			oldKVM.ResourceState = kvm.GetResourceState()
		case "platform":
			oldKVM.ChromePlatform = kvm.GetChromePlatform()
		case "macAddress":
			oldKVM.MacAddress = kvm.GetMacAddress()
		case "tags":
			oldKVM.Tags = mergeTags(oldKVM.GetTags(), kvm.GetTags())
		case "description":
			oldKVM.Description = kvm.GetDescription()
		}
	}
	// return existing/old kvm with new updated values
	return oldKVM, nil
}

// DeleteKVMHost deletes the host of a kvm in datastore.
func DeleteKVMHost(ctx context.Context, kvmName string) error {
	f := func(ctx context.Context) error {
		hc := getKVMHistoryClient(&ufspb.KVM{Name: kvmName})
		if err := hc.netUdt.deleteDHCPHelper(ctx); err != nil {
			return err
		}
		if err := hc.stUdt.updateStateHelper(ctx, ufspb.State_STATE_REGISTERED); err != nil {
			return errors.Annotate(err, "Fail to update state to kvm %s", kvmName).Err()
		}
		return hc.SaveChangeEvents(ctx)
	}

	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to delete the kvm host: %s", err)
		return err
	}
	return nil
}

// UpdateKVMHost updates the kvm host in datastore.
func UpdateKVMHost(ctx context.Context, kvm *ufspb.KVM, nwOpt *ufsAPI.NetworkOption) error {
	f := func(ctx context.Context) error {
		hc := getKVMHistoryClient(kvm)
		// 1. Validate the input
		if err := validateUpdateKVMHost(ctx, kvm, nwOpt.GetVlan(), nwOpt.GetIp()); err != nil {
			return err
		}
		// 2. Verify if the hostname is already set with IP. if yes, remove the current dhcp.
		if err := hc.netUdt.deleteDHCPHelper(ctx); err != nil {
			return err
		}

		// 3. Find free ip, set IP and DHCP config
		if _, err := hc.netUdt.addHostHelper(ctx, nwOpt.GetVlan(), nwOpt.GetIp(), kvm.GetMacAddress()); err != nil {
			return err
		}

		if err := hc.stUdt.updateStateHelper(ctx, ufspb.State_STATE_DEPLOYING); err != nil {
			return errors.Annotate(err, "Fail to update state to kvm %s", kvm.GetName()).Err()
		}
		return hc.SaveChangeEvents(ctx)
	}

	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to assign IP to the kvm: %s", err)
		return err
	}
	return nil
}

// GetKVM returns kvm for the given id from datastore.
func GetKVM(ctx context.Context, id string) (*ufspb.KVM, error) {
	return registration.GetKVM(ctx, id)
}

// BatchGetKVMs returns a batch of kvms based on ids.
func BatchGetKVMs(ctx context.Context, ids []string) ([]*ufspb.KVM, error) {
	return registration.BatchGetKVM(ctx, ids)
}

// ListKVMs lists the kvms
func ListKVMs(ctx context.Context, pageSize int32, pageToken, filter string, keysOnly bool) ([]*ufspb.KVM, string, error) {
	var filterMap map[string][]interface{}
	var err error
	if filter != "" {
		filterMap, err = getFilterMap(filter, registration.GetKVMIndexedFieldName)
		if err != nil {
			return nil, "", errors.Annotate(err, "Failed to read filter for listing kvms").Err()
		}
	}
	filterMap = resetStateFilter(filterMap)
	filterMap = resetZoneFilter(filterMap)
	return registration.ListKVMs(ctx, pageSize, pageToken, filterMap, keysOnly)
}

// DeleteKVM deletes the kvm in datastore
func DeleteKVM(ctx context.Context, id string) error {
	return deleteKVMHelper(ctx, id, true)
}

func deleteKVMHelper(ctx context.Context, id string, inTransaction bool) error {
	f := func(ctx context.Context) error {
		hc := getKVMHistoryClient(&ufspb.KVM{Name: id})

		// Get kvm
		kvm, err := registration.GetKVM(ctx, id)
		if err != nil {
			return errors.Annotate(err, "Unable to get KVM").Err()
		}

		// Validate input
		if err := validateDeleteKVM(ctx, kvm); err != nil {
			return errors.Annotate(err, "Validation failed - unable to delete kvm %s", id).Err()
		}

		// Delete the kvm
		if err := registration.DeleteKVM(ctx, id); err != nil {
			return errors.Annotate(err, "Delete failed - unable to delete kvm %s", id).Err()
		}

		// Update state
		hc.stUdt.deleteStateHelper(ctx)

		// Delete ip configs
		if err := hc.netUdt.deleteDHCPHelper(ctx); err != nil {
			return err
		}
		hc.LogKVMChanges(kvm, nil)
		return hc.SaveChangeEvents(ctx)
	}
	if inTransaction {
		if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
			logging.Errorf(ctx, "Failed to delete kvm in datastore: %s", err)
			return err
		}
		return nil
	}
	return f(ctx)
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

func getKVMHistoryClient(kvm *ufspb.KVM) *HistoryClient {
	return &HistoryClient{
		stUdt: &stateUpdater{
			ResourceName: ufsUtil.AddPrefix(ufsUtil.KVMCollection, kvm.Name),
		},
		netUdt: &networkUpdater{
			Hostname: kvm.Name,
		},
	}
}

// validateDeleteKVM validates if a KVM can be deleted
//
// Checks if this KVM(KVMID) is not referenced by other resources in the datastore.
// If there are any other references, delete will be rejected and an error will be returned.
func validateDeleteKVM(ctx context.Context, kvm *ufspb.KVM) error {
	rack, err := registration.GetRack(ctx, kvm.GetRack())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "rack %s not found", kvm.GetRack())
	}
	// Check permission
	if err := ufsUtil.CheckPermission(ctx, ufsUtil.RegistrationsDelete, rack.GetRealm()); err != nil {
		return err
	}
	machines, err := registration.QueryMachineByPropertyName(ctx, "kvm_id", kvm.GetName(), true)
	if err != nil {
		return err
	}
	if len(machines) > 0 {
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("KVM %s cannot be deleted because there are other resources which are referring this KVM.", kvm.GetName()))
		if len(machines) > 0 {
			errorMsg.WriteString("\nMachines referring the KVM:\n")
			for _, machine := range machines {
				errorMsg.WriteString(machine.Name + ", ")
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
func validateCreateKVM(ctx context.Context, kvm *ufspb.KVM, rack *ufspb.Rack) error {
	// Check permission
	if err := ufsUtil.CheckPermission(ctx, ufsUtil.RegistrationsCreate, rack.GetRealm()); err != nil {
		return err
	}
	// Check if kvm already exists
	if err := resourceAlreadyExists(ctx, []*Resource{GetKVMResource(kvm.Name)}, nil); err != nil {
		return err
	}
	if err := validateMacAddress(ctx, kvm.GetName(), kvm.GetMacAddress()); err != nil {
		return err
	}
	// Aggregate resource to check if resources referenced by the kvm does not exist
	if chromePlatformID := kvm.GetChromePlatform(); chromePlatformID != "" {
		return ResourceExist(ctx, []*Resource{GetChromePlatformResource(chromePlatformID)}, nil)
	}
	return nil
}

// validateUpdateKVM validates if a kvm can be updated
//
// check if kvm, rack and resources referenced kvm does not exist
func validateUpdateKVM(ctx context.Context, oldKvm *ufspb.KVM, kvm *ufspb.KVM, mask *field_mask.FieldMask) error {
	rack, err := registration.GetRack(ctx, oldKvm.GetRack())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "rack %s not found", oldKvm.GetRack())
	}
	// Check permission
	if err := ufsUtil.CheckPermission(ctx, ufsUtil.RegistrationsUpdate, rack.GetRealm()); err != nil {
		return err
	}
	// Aggregate resource to check if kvm does not exist
	var resourcesNotFound []*Resource
	// Aggregate resource to check if rack does not exist
	if kvm.GetRack() != "" {
		resourcesNotFound = append(resourcesNotFound, GetRackResource(kvm.GetRack()))
	}
	// Aggregate resource to check if resources referenced by the kvm does not exist
	if chromePlatformID := kvm.GetChromePlatform(); chromePlatformID != "" {
		resourcesNotFound = append(resourcesNotFound, GetChromePlatformResource(chromePlatformID))
	}
	// check if resources does not exist
	if err := ResourceExist(ctx, resourcesNotFound, nil); err != nil {
		return err
	}

	return validateKVMUpdateMask(ctx, kvm, mask)
}

// validateKVMUpdateMask validates the update mask for kvm update
func validateKVMUpdateMask(ctx context.Context, kvm *ufspb.KVM, mask *field_mask.FieldMask) error {
	if mask != nil {
		// validate the give field mask
		for _, path := range mask.Paths {
			switch path {
			case "name":
				return status.Error(codes.InvalidArgument, "validateUpdateKVM - name cannot be updated, delete and create a new kvm instead")
			case "update_time":
				return status.Error(codes.InvalidArgument, "validateUpdateKVM - update_time cannot be updated, it is a Output only field")
			case "macAddress":
				if err := validateMacAddress(ctx, kvm.GetName(), kvm.GetMacAddress()); err != nil {
					return err
				}
			case "rack":
				if kvm.GetRack() == "" {
					return status.Error(codes.InvalidArgument, "rack cannot be empty for updating a KVM")
				}
			case "platform":
			case "description":
			case "tags":
			case "resourceState":
				// valid fields, nothing to validate.
			default:
				return status.Errorf(codes.InvalidArgument, "validateUpdateKVM - unsupported update mask path %q", path)
			}
		}
	}
	if err := validateMacAddress(ctx, kvm.GetName(), kvm.GetMacAddress()); err != nil {
		return err
	}
	return nil
}

// validateUpdateKVMHost validates if a host can be added to a kvm
func validateUpdateKVMHost(ctx context.Context, kvm *ufspb.KVM, vlanName, ipv4Str string) error {
	// during partial update, kvm object may not have rack info, so we get the old kvm to get the rack
	// to check the permission
	oldKvm, err := registration.GetKVM(ctx, kvm.GetName())
	if err != nil {
		return err
	}
	rack, err := registration.GetRack(ctx, oldKvm.GetRack())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "rack %s not found", oldKvm.GetRack())
	}
	// Check permission
	if err := ufsUtil.CheckPermission(ctx, ufsUtil.RegistrationsUpdate, rack.GetRealm()); err != nil {
		return err
	}
	if kvm.GetMacAddress() == "" {
		return errors.New("mac address of kvm hasn't been specified")
	}
	if ipv4Str != "" {
		return nil
	}
	// Check if resources does not exist
	return ResourceExist(ctx, []*Resource{GetKVMResource(kvm.Name), GetVlanResource(vlanName)}, nil)
}
