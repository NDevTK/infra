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
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/proto"
	ufsds "infra/unifiedfleet/app/model/datastore"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"
	"infra/unifiedfleet/app/util"
)

// MachineRegistration creates a new machine, new nic and a new drac in datastore.
func MachineRegistration(ctx context.Context, machine *ufspb.Machine) (*ufspb.Machine, error) {
	nics := machine.GetChromeBrowserMachine().GetNicObjects()
	drac := machine.GetChromeBrowserMachine().GetDracObject()
	f := func(ctx context.Context) error {
		hc := getMachineHistoryClient(machine)
		var err error
		// Validate input
		if err = validateMachineRegistration(ctx, machine); err != nil {
			return errors.Annotate(err, "MachineRegistration - validation failed").Err()
		}

		// Create nics
		if nics != nil {
			for _, nic := range nics {
				// Fill the machine/rack/lab to nic OUTPUT only fields
				nic.Machine = machine.GetName()
				nic.Rack = machine.GetLocation().GetRack()
				nic.Lab = machine.GetLocation().GetLab().String()
			}

			// we use this func as it is a non-atomic operation and can be used to
			// run within a transaction to make it atomic. Datastore doesnt allow
			// nested transactions.
			if _, err = registration.BatchUpdateNics(ctx, nics); err != nil {
				return errors.Annotate(err, "MachineRegistration - unable to batch update nics").Err()
			}
			for _, nic := range nics {
				hc.LogNicChanges(nil, nic)
			}

			// we dont store nic details inside a machine object
			// we have a separate nic table indexed by machine name
			machine.GetChromeBrowserMachine().NicObjects = nil
		}

		// Create drac
		if drac != nil {
			// Fill the machine/rack/lab to drac OUTPUT only fields
			drac.Machine = machine.GetName()
			drac.Rack = machine.GetLocation().GetRack()
			drac.Lab = machine.GetLocation().GetLab().String()

			// we use this func as it is a non-atomic operation and can be used to
			// run within a transaction to make it atomic. Datastore doesnt allow
			// nested transactions.
			if _, err = registration.BatchUpdateDracs(ctx, []*ufspb.Drac{drac}); err != nil {
				return errors.Annotate(err, "MachineRegistration - unable to batch update drac").Err()
			}
			hc.LogDracChanges(nil, drac)
			// we dont store drac details inside a machine object
			// We have a separate drac table indexed by machine name
			machine.GetChromeBrowserMachine().DracObject = nil
		}

		// Create the machine
		// we use this func as it is a non-atomic operation and can be used to
		// run within a transaction to make it atomic. Datastore doesnt allow
		// nested transactions.
		if _, err = registration.BatchUpdateMachines(ctx, []*ufspb.Machine{machine}); err != nil {
			return errors.Annotate(err, "MachineRegistration - unable to batch update machine").Err()
		}
		hc.LogMachineChanges(nil, machine)

		if machine.GetChromeBrowserMachine() != nil {
			// We fill the machine object with newly created nics/drac from nic/drac table
			// This will have all the extra information for nics/drac(machine name, updated time.. etc)
			machine.GetChromeBrowserMachine().NicObjects = nics
			machine.GetChromeBrowserMachine().DracObject = drac
		}

		// Update machine state
		if err := hc.stUdt.updateStateHelper(ctx, ufspb.State_STATE_REGISTERED); err != nil {
			return err
		}
		return hc.SaveChangeEvents(ctx)
	}

	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		return nil, errors.Annotate(err, "MachineRegistration - failed to create machine/nics/drac in datastore").Err()
	}
	return machine, nil
}

// UpdateMachine updates machine in datastore.
//
// Checks if the resources referenced by the Machine input already exists
// in the system before updating a Machine
func UpdateMachine(ctx context.Context, machine *ufspb.Machine, mask *field_mask.FieldMask) (*ufspb.Machine, error) {
	var oldMachine *ufspb.Machine
	var err error
	f := func(ctx context.Context) error {
		hc := getMachineHistoryClient(machine)
		// Validate input
		if err := validateUpdateMachine(ctx, machine, mask); err != nil {
			return errors.Annotate(err, "UpdateMachine - validation failed").Err()
		}

		if machine.GetChromeBrowserMachine() != nil {
			// We dont update the nics and dracs in UpdateMachine call.
			// We dont store nics/drac object inside Machine object in Machine table.
			// nics/drac objects are stored in their separate tables
			// user has to use nic/drac CRUD apis to update nic/drac
			machine.GetChromeBrowserMachine().NicObjects = nil
			machine.GetChromeBrowserMachine().DracObject = nil
		}

		// Get the existing/old machine
		oldMachine, err = registration.GetMachine(ctx, machine.GetName())
		if err != nil {
			return errors.Annotate(err, "UpdateMachine - get machine %s failed", machine.GetName()).Err()
		}

		// Do not let updating from browser to os or vice versa change for machine.
		if oldMachine.GetChromeBrowserMachine() != nil && machine.GetChromeosMachine() != nil {
			return status.Error(codes.InvalidArgument, "UpdateMachine - cannot update a browser machine to os machine. Please delete the browser machine and create a new os machine")
		}
		if oldMachine.GetChromeosMachine() != nil && machine.GetChromeBrowserMachine() != nil {
			return status.Error(codes.InvalidArgument, "UpdateMachine - cannot update an os machine to browser machine. Please delete the os machine and create a new browser machine")
		}

		// Partial update by field mask
		if mask != nil && len(mask.Paths) > 0 {
			machine, err = processMachineUpdateMask(ctx, oldMachine, machine, mask)
			if err != nil {
				return errors.Annotate(err, "UpdateMachine - processing update mask failed").Err()
			}
		} else if machine.GetLocation().GetRack() != oldMachine.GetLocation().GetRack() ||
			machine.GetLocation().GetLab() != oldMachine.GetLocation().GetLab() {
			// this check is for json input with complete update machine
			// Check if machine lab/rack information is changed/updated
			indexMap := map[string]string{
				"lab": machine.GetLocation().GetLab().String(), "rack": machine.GetLocation().GetRack()}
			if err = updateIndexingForMachineResources(ctx, oldMachine, indexMap); err != nil {
				return errors.Annotate(err, "UpdateMachine - update lab and rack indexing failed").Err()
			}
		}

		// update the machine
		// we use this func as it is a non-atomic operation and can be used to
		// run within a transaction to make it atomic. Datastore doesnt allow
		// nested transactions.
		if _, err := registration.BatchUpdateMachines(ctx, []*ufspb.Machine{machine}); err != nil {
			return errors.Annotate(err, "UpdateMachine - unable to batch update machine %s", machine.Name).Err()
		}
		hc.LogMachineChanges(oldMachine, machine)
		return hc.SaveChangeEvents(ctx)
	}

	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		return nil, errors.Annotate(err, "UpdateMachine - failed to update machine %s in datastore", machine.Name).Err()
	}
	if oldMachine.GetChromeBrowserMachine() != nil {
		// We fill the machine object with its nics/drac from nic/drac table
		setMachine(ctx, machine)
	}
	return machine, nil
}

// processMachineUpdateMask process update field mask to get only specific update
// fields and return a complete machine object with updated and existing fields
func processMachineUpdateMask(ctx context.Context, oldMachine *ufspb.Machine, machine *ufspb.Machine, mask *field_mask.FieldMask) (*ufspb.Machine, error) {
	// update the fields in the existing nic
	for _, path := range mask.Paths {
		switch path {
		case "lab":
			if err := updateIndexingForMachineResources(ctx, oldMachine, map[string]string{"lab": machine.GetLocation().GetLab().String()}); err != nil {
				return oldMachine, errors.Annotate(err, "processMachineUpdateMask - failed to update lab indexing").Err()
			}
			if oldMachine.GetLocation() == nil {
				oldMachine.Location = &ufspb.Location{}
			}
			oldMachine.GetLocation().Lab = machine.GetLocation().GetLab()
		case "rack":
			if err := updateIndexingForMachineResources(ctx, oldMachine, map[string]string{"rack": machine.GetLocation().GetRack()}); err != nil {
				return oldMachine, errors.Annotate(err, "processMachineUpdateMask - failed to update rack indexing").Err()
			}
			if oldMachine.GetLocation() == nil {
				oldMachine.Location = &ufspb.Location{}
			}
			oldMachine.GetLocation().Rack = machine.GetLocation().GetRack()
		case "platform":
			if oldMachine.GetChromeBrowserMachine() == nil {
				oldMachine.Device = &ufspb.Machine_ChromeBrowserMachine{
					ChromeBrowserMachine: &ufspb.ChromeBrowserMachine{},
				}
			}
			oldMachine.GetChromeBrowserMachine().ChromePlatform = machine.GetChromeBrowserMachine().GetChromePlatform()
		case "kvm":
			if oldMachine.GetChromeBrowserMachine() == nil {
				oldMachine.Device = &ufspb.Machine_ChromeBrowserMachine{
					ChromeBrowserMachine: &ufspb.ChromeBrowserMachine{},
				}
			}
			if oldMachine.GetChromeBrowserMachine().GetKvmInterface() == nil {
				oldMachine.GetChromeBrowserMachine().KvmInterface = &ufspb.KVMInterface{}
			}
			oldMachine.GetChromeBrowserMachine().GetKvmInterface().Kvm = machine.GetChromeBrowserMachine().GetKvmInterface().GetKvm()
		case "ticket":
			if oldMachine.GetChromeBrowserMachine() == nil {
				oldMachine.Device = &ufspb.Machine_ChromeBrowserMachine{
					ChromeBrowserMachine: &ufspb.ChromeBrowserMachine{},
				}
			}
			oldMachine.GetChromeBrowserMachine().DeploymentTicket = machine.GetChromeBrowserMachine().GetDeploymentTicket()
		case "tags":
			oldMachine.Tags = mergeTags(oldMachine.GetTags(), machine.GetTags())
		}
	}
	// return existing/old machine with new updated values
	return oldMachine, nil
}

// updateIndexingForMachineResources updates indexing for machinelse/nic/drac tables
// can be used inside a transaction
func updateIndexingForMachineResources(ctx context.Context, machine *ufspb.Machine, indexMap map[string]string) error {
	// get MachineLSEs for indexing
	machinelses, err := inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", machine.GetName(), false)
	if err != nil {
		return errors.Annotate(err, "updateIndexing - failed to query machinelses/hosts for machine %s", machine.GetName()).Err()
	}
	var nics []*ufspb.Nic
	var dracs []*ufspb.Drac
	// update indexing for nic/drac table only for chrome browser machines.
	if machine.GetChromeBrowserMachine() != nil {
		var err error
		// get Nics for indexing
		nics, err = registration.QueryNicByPropertyName(ctx, "machine", machine.GetName(), false)
		if err != nil {
			return errors.Annotate(err, "updateIndexing - failed to query nics for machine %s", machine.GetName()).Err()
		}
		// get Dracs for indexing
		dracs, err = registration.QueryDracByPropertyName(ctx, "machine", machine.GetName(), false)
		if err != nil {
			return errors.Annotate(err, "updateIndexing - failed to query dracs for machine %s", machine.GetName()).Err()
		}
	}

	for k, v := range indexMap {
		// These are output only fields used for indexing machinelse/drac/nic table
		switch k {
		case "rack":
			for _, machinelse := range machinelses {
				machinelse.Rack = v
			}
			for _, nic := range nics {
				nic.Rack = v
			}
			for _, drac := range dracs {
				drac.Rack = v
			}
		case "lab":
			for _, machinelse := range machinelses {
				machinelse.Lab = v
			}
			for _, nic := range nics {
				nic.Lab = v
			}
			for _, drac := range dracs {
				drac.Lab = v
			}
		}
	}
	if _, err := inventory.BatchUpdateMachineLSEs(ctx, machinelses); err != nil {
		return errors.Annotate(err, "updateIndexing - unable to batch update machinelses").Err()
	}
	// update indexing for nic/drac table only for chrome browser machines.
	if machine.GetChromeBrowserMachine() != nil {
		if _, err := registration.BatchUpdateNics(ctx, nics); err != nil {
			return errors.Annotate(err, "updateIndexing - unable to batch update nics").Err()
		}
		if _, err := registration.BatchUpdateDracs(ctx, dracs); err != nil {
			return errors.Annotate(err, "updateIndexing - unable to batch update dracs").Err()
		}
	}
	return nil
}

// GetMachine returns machine for the given id from datastore.
func GetMachine(ctx context.Context, id string) (*ufspb.Machine, error) {
	machine, err := registration.GetMachine(ctx, id)
	if err != nil {
		return nil, err
	}
	setMachine(ctx, machine)
	return machine, nil
}

// ListMachines lists the machines
func ListMachines(ctx context.Context, pageSize int32, pageToken, filter string, keysOnly bool) ([]*ufspb.Machine, string, error) {
	var filterMap map[string][]interface{}
	var err error
	if filter != "" {
		filterMap, err = getFilterMap(filter, registration.GetMachineIndexedFieldName)
		if err != nil {
			return nil, "", errors.Annotate(err, "ListMachines - failed to read filter for listing machines").Err()
		}
	}
	return registration.ListMachines(ctx, pageSize, pageToken, filterMap, keysOnly)
}

// GetAllMachines returns all machines in datastore.
func GetAllMachines(ctx context.Context) (*ufsds.OpResults, error) {
	return registration.GetAllMachines(ctx)
}

// DeleteMachine deletes the machine and its associated nics and drac in datastore
//
// For referential data intergrity,
// Delete if this Machine is not referenced by other resources in the datastore.
// If there are any references, delete will be rejected and an error will be returned.
func DeleteMachine(ctx context.Context, id string) error {
	f := func(ctx context.Context) error {
		hc := getMachineHistoryClient(&ufspb.Machine{Name: id})
		// 1. Get the machine
		machine, err := registration.GetMachine(ctx, id)
		if status.Code(err) == codes.Internal {
			return errors.Annotate(err, "DeleteMachine - failed to get machine %s", id).Err()
		}
		if machine == nil {
			return status.Errorf(codes.NotFound, ufsds.NotFound)
		}

		// 2. Check if any other resource references this machine.
		if err = validateDeleteMachine(ctx, id); err != nil {
			return errors.Annotate(err, "DeleteMachine - validation failed").Err()
		}

		//Only for a browser machine
		if machine.GetChromeBrowserMachine() != nil {
			// 3. Delete the nics
			// get Nics for machine - only keys are sufficient
			nics, err := registration.QueryNicByPropertyName(ctx, "machine", machine.GetName(), true)
			if err != nil {
				return errors.Annotate(err, "DeleteMachine - failed to query nics for machine %s", machine.GetName()).Err()
			}
			if nics != nil {
				var nicIDs []string
				for _, nic := range nics {
					nicIDs = append(nicIDs, nic.GetName())
					nicHc := getNicHistoryClient(nic)
					nicHc.LogNicChanges(nic, nil)
					if err = nicHc.SaveChangeEvents(ctx); err != nil {
						return err
					}
				}
				// we use this func as it is a non-atomic operation and can be used to
				// run within a transaction to make it atomic. Datastore doesnt allow
				// nested transactions.
				if err = registration.BatchDeleteNics(ctx, nicIDs); err != nil {
					return errors.Annotate(err, "DeleteMachine - failed to batch delete nics for machine %s", machine.GetName()).Err()
				}
			}

			// 4. Delete the drac
			// get Dracs for machine - only keys are sufficient
			dracs, err := registration.QueryDracByPropertyName(ctx, "machine", machine.GetName(), true)
			if err != nil {
				return errors.Annotate(err, "DeleteMachine - failed to query dracs for machine %s", machine.GetName()).Err()
			}
			if dracs != nil && len(dracs) > 0 {
				dracHc := getDracHistoryClient(dracs[0])
				if err = registration.DeleteDrac(ctx, dracs[0].GetName()); err != nil {
					return errors.Annotate(err, "DeleteMachine - failed to delete drac %s for machine %s", dracs[0].GetName(), machine.GetName()).Err()
				}
				if err := dracHc.netUdt.deleteDHCPHelper(ctx); err != nil {
					return errors.Annotate(err, "DeleteMachine - unable to delete ip configs for drac %s", dracs[0].GetName()).Err()
				}
				dracHc.LogDracChanges(dracs[0], nil)
				if err = dracHc.SaveChangeEvents(ctx); err != nil {
					return err
				}
			}
		}

		// 5. Delete the state
		hc.stUdt.deleteStateHelper(ctx)
		if err := registration.DeleteMachine(ctx, id); err != nil {
			return err
		}
		hc.LogMachineChanges(machine, nil)
		// 6. Delete the machine
		return hc.SaveChangeEvents(ctx)
	}

	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		return errors.Annotate(err, "DeleteMachine - failed to delete machine and its associated nics and drac in datastore").Err()
	}
	return nil
}

// ImportMachines creates or updates a batch of machines in datastore
func ImportMachines(ctx context.Context, machines []*ufspb.Machine, pageSize int) (*ufsds.OpResults, error) {
	deleteNonExistingMachines(ctx, machines, pageSize)
	allRes := make(ufsds.OpResults, 0)
	for i := 0; ; i += pageSize {
		end := util.Min(i+pageSize, len(machines))
		res, err := registration.ImportMachines(ctx, machines[i:end])
		allRes = append(allRes, *res...)
		if err != nil {
			return &allRes, err
		}
		if i+pageSize >= len(machines) {
			break
		}
	}
	return &allRes, nil
}

func deleteNonExistingMachines(ctx context.Context, machines []*ufspb.Machine, pageSize int) (*ufsds.OpResults, error) {
	resMap := make(map[string]bool)
	for _, r := range machines {
		resMap[r.GetName()] = true
	}
	resp, err := registration.GetAllMachines(ctx)
	if err != nil {
		return nil, err
	}
	var toDelete []string
	for _, sr := range resp.Passed() {
		s := sr.Data.(*ufspb.Machine)
		if _, ok := resMap[s.GetName()]; !ok {
			toDelete = append(toDelete, s.GetName())
		}
	}
	logging.Debugf(ctx, "Deleting %d non-existing machines", len(toDelete))
	return deleteByPage(ctx, toDelete, pageSize, registration.DeleteMachines), nil
}

// ReplaceMachine replaces an old Machine with new Machine in datastore
//
// It does a delete of old machine and create of new Machine.
// All the steps are in done in a transaction to preserve consistency on failure.
// Before deleting the old Machine, it will get all the resources referencing
// the old Machine. It will update all the resources which were referencing
// the old Machine(got in the last step) with new Machine.
// Deletes the old Machine.
// Creates the new Machine.
// This will preserve data integrity in the system.
func ReplaceMachine(ctx context.Context, oldMachine *ufspb.Machine, newMachine *ufspb.Machine) (*ufspb.Machine, error) {
	f := func(ctx context.Context) error {
		hc := getMachineHistoryClient(newMachine)
		machinelses, err := inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", oldMachine.Name, false)
		if err != nil {
			return err
		}
		if machinelses != nil {
			for _, machinelse := range machinelses {
				machines := machinelse.GetMachines()
				for i := range machines {
					if machines[i] == oldMachine.Name {
						machines[i] = newMachine.Name
						break
					}
				}
				machinelse.Machines = machines
			}
			_, err := inventory.BatchUpdateMachineLSEs(ctx, machinelses)
			if err != nil {
				return err
			}
		}

		err = registration.DeleteMachine(ctx, oldMachine.Name)
		if err != nil {
			return err
		}

		err = validateMachineReferences(ctx, newMachine)
		if err != nil {
			return err
		}
		entity := &registration.MachineEntity{
			ID: newMachine.Name,
		}
		existsResults, err := datastore.Exists(ctx, entity)
		if err == nil {
			if existsResults.All() {
				return status.Errorf(codes.AlreadyExists, ufsds.AlreadyExists)
			}
		} else {
			logging.Errorf(ctx, "Failed to check existence: %s", err)
			return status.Errorf(codes.Internal, ufsds.InternalError)
		}

		_, err = registration.BatchUpdateMachines(ctx, []*ufspb.Machine{newMachine})
		if err != nil {
			return err
		}
		hc.LogMachineChanges(oldMachine, nil)
		hc.LogMachineChanges(nil, newMachine)
		hc.stUdt.replaceStateHelper(ctx, util.AddPrefix(util.MachineCollection, oldMachine.GetName()))
		return hc.SaveChangeEvents(ctx)
	}

	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to replace entity in datastore: %s", err)
		return nil, err
	}
	return newMachine, nil
}

// validateMachineReferences validates if the resources referenced by the machine
// are in the system.
//
// Checks if the resources referenced by the given Machine input already exists
// in the system. Returns an error if any resource referenced by the Machine input
// does not exist in the system.
func validateMachineReferences(ctx context.Context, machine *ufspb.Machine) error {
	var resources []*Resource
	var errorMsg strings.Builder
	errorMsg.WriteString(fmt.Sprintf("Cannot create Machine %s:\n", machine.Name))

	if kvmID := machine.GetChromeBrowserMachine().GetKvmInterface().GetKvm(); kvmID != "" {
		resources = append(resources, GetKVMResource(kvmID))
	}
	if rpmID := machine.GetChromeBrowserMachine().GetRpmInterface().GetRpm(); rpmID != "" {
		resources = append(resources, GetRPMResource(rpmID))
	}
	if chromePlatformID := machine.GetChromeBrowserMachine().GetChromePlatform(); chromePlatformID != "" {
		resources = append(resources, GetChromePlatformResource(chromePlatformID))
	}

	return ResourceExist(ctx, resources, &errorMsg)
}

// validateDeleteMachine validates if a Machine can be deleted
//
// Checks if this Machine(MachineID) is not referenced by other resources in the datastore.
// If there are any other references, delete will be rejected and an error will be returned.
func validateDeleteMachine(ctx context.Context, id string) error {
	machinelses, err := inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", id, true)
	if err != nil {
		return err
	}
	if len(machinelses) > 0 {
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("Machine %s cannot be deleted because there are other resources which are referring this Machine.", id))
		errorMsg.WriteString(fmt.Sprintf("\nHosts referring the Machine:\n"))
		for _, machinelse := range machinelses {
			errorMsg.WriteString(machinelse.Name + ", ")
		}
		errorMsg.WriteString(fmt.Sprintf("\nPlease delete the hosts and then delete the machine.\n"))
		logging.Errorf(ctx, errorMsg.String())
		return status.Errorf(codes.FailedPrecondition, errorMsg.String())
	}
	return nil
}

// getBrowserMachine gets the browser machine
func getBrowserMachine(ctx context.Context, machineName string) (*ufspb.Machine, error) {
	machine, err := registration.GetMachine(ctx, machineName)
	if err != nil {
		return nil, errors.Annotate(err, "getBrowserMachine - unable to get browser machine %s", machineName).Err()
	}
	if machine.GetChromeBrowserMachine() == nil {
		errorMsg := fmt.Sprintf("Machine %s is not a browser machine.", machineName)
		return nil, status.Errorf(codes.FailedPrecondition, errorMsg)
	}
	return machine, nil
}

// validateMachineRegistration validates if a machine/nics/drac can be created
func validateMachineRegistration(ctx context.Context, machine *ufspb.Machine) error {
	var resourcesNotFound []*Resource
	var resourcesAlreadyExists []*Resource
	var errorMsg strings.Builder
	errorMsg.WriteString(fmt.Sprintf("Cannot create machine %s:\n", machine.Name))
	var nics []*ufspb.Nic
	var drac *ufspb.Drac
	if machine.GetChromeBrowserMachine() != nil {
		nics = machine.GetChromeBrowserMachine().GetNicObjects()
		drac = machine.GetChromeBrowserMachine().GetDracObject()
	}
	// Aggregate resources to check if machine already exists
	resourcesAlreadyExists = append(resourcesAlreadyExists, GetMachineResource(machine.Name))
	for _, nic := range nics {
		// Aggregate resources to check if nic already exists
		resourcesAlreadyExists = append(resourcesAlreadyExists, GetNicResource(nic.Name))

		// Aggregate resources to check if resources referenced by the nic exists
		if switchID := nic.GetSwitchInterface().GetSwitch(); switchID != "" {
			resourcesNotFound = append(resourcesNotFound, GetSwitchResource(switchID))
		}
	}
	if drac != nil {
		// Aggregate resources to check if drac already exists
		resourcesAlreadyExists = append(resourcesAlreadyExists, GetDracResource(drac.Name))

		// Aggregate resources to check if resources referenced by the drac exists
		if switchID := drac.GetSwitchInterface().GetSwitch(); switchID != "" {
			resourcesNotFound = append(resourcesNotFound, GetSwitchResource(switchID))
		}
	}
	// Aggregate resources referenced by the machine to check if they do not exist
	if kvmID := machine.GetChromeBrowserMachine().GetKvmInterface().GetKvm(); kvmID != "" {
		resourcesNotFound = append(resourcesNotFound, GetKVMResource(kvmID))
	}
	if rpmID := machine.GetChromeBrowserMachine().GetRpmInterface().GetRpm(); rpmID != "" {
		resourcesNotFound = append(resourcesNotFound, GetRPMResource(rpmID))
	}
	if chromePlatformID := machine.GetChromeBrowserMachine().GetChromePlatform(); chromePlatformID != "" {
		resourcesNotFound = append(resourcesNotFound, GetChromePlatformResource(chromePlatformID))
	}
	// Check if machine/nics/drac already exists
	if err := resourceAlreadyExists(ctx, resourcesAlreadyExists, nil); err != nil {
		return err
	}
	// check if resources does not exist
	return ResourceExist(ctx, resourcesNotFound, &errorMsg)
}

// validateUpdateMachine validates if a machine can be updated
//
// checks if the machine and the resources referenced  by the machine
// does not exist in the system.
func validateUpdateMachine(ctx context.Context, machine *ufspb.Machine, mask *field_mask.FieldMask) error {
	var errorMsg strings.Builder
	errorMsg.WriteString(fmt.Sprintf("Cannot update machine %s:\n", machine.Name))
	// Aggregate resource to check if machine does not exist
	resourcesNotFound := []*Resource{GetMachineResource(machine.Name)}
	// Aggregate resources referenced by the machine to check if they do not exist
	if kvmID := machine.GetChromeBrowserMachine().GetKvmInterface().GetKvm(); kvmID != "" {
		resourcesNotFound = append(resourcesNotFound, GetKVMResource(kvmID))
	}
	if rpmID := machine.GetChromeBrowserMachine().GetRpmInterface().GetRpm(); rpmID != "" {
		resourcesNotFound = append(resourcesNotFound, GetRPMResource(rpmID))
	}
	if chromePlatformID := machine.GetChromeBrowserMachine().GetChromePlatform(); chromePlatformID != "" {
		resourcesNotFound = append(resourcesNotFound, GetChromePlatformResource(chromePlatformID))
	}

	// check if resources does not exist
	if err := ResourceExist(ctx, resourcesNotFound, &errorMsg); err != nil {
		return err
	}

	return validateMachineUpdateMask(machine, mask)
}

// validateMachineUpdateMask validates the update mask for machine update
func validateMachineUpdateMask(machine *ufspb.Machine, mask *field_mask.FieldMask) error {
	if mask != nil {
		// validate the give field mask
		for _, path := range mask.Paths {
			switch path {
			case "name":
				return status.Error(codes.InvalidArgument, "validateMachineUpdateMask - name cannot be updated, delete and create a new machine instead")
			case "update_time":
				return status.Error(codes.InvalidArgument, "validateMachineUpdateMask - update_time cannot be updated, it is a output only field")
			case "lab":
				if machine.GetLocation() == nil {
					return status.Error(codes.InvalidArgument, "validateMachineUpdateMask - location cannot be empty/nil.")
				}
			case "rack":
				if machine.GetLocation() == nil {
					return status.Error(codes.InvalidArgument, "validateMachineUpdateMask - location cannot be empty/nil.")
				}
			case "platform":
				if machine.GetChromeBrowserMachine() == nil {
					return status.Error(codes.InvalidArgument, "validateMachineUpdateMask - browser machine cannot be empty/nil.")
				}
			case "kvm":
				if machine.GetChromeBrowserMachine() == nil {
					return status.Error(codes.InvalidArgument, "validateMachineUpdateMask - browser machine cannot be empty/nil.")
				}
				if machine.GetChromeBrowserMachine().GetKvmInterface() == nil {
					return status.Error(codes.InvalidArgument, "validateMachineUpdateMask - kvm interface cannot be empty/nil.")
				}
			case "ticket":
				if machine.GetChromeBrowserMachine() == nil {
					return status.Error(codes.InvalidArgument, "validateMachineUpdateMask - browser machine cannot be empty/nil.")
				}
			case "tags":
				// valid fields, nothing to validate.
			default:
				return status.Errorf(codes.InvalidArgument, "validateMachineUpdateMask - unsupported update mask path %q", path)
			}
		}
	}
	return nil
}

func getMachineHistoryClient(m *ufspb.Machine) *HistoryClient {
	return &HistoryClient{
		stUdt: &stateUpdater{
			ResourceName: util.AddPrefix(util.MachineCollection, m.Name),
		},
		netUdt: &networkUpdater{
			Hostname: m.Name,
		},
	}
}

//setMachine populates the amchine object with nics and drac
func setMachine(ctx context.Context, machine *ufspb.Machine) {
	// get Nics for machine
	nics, err := registration.QueryNicByPropertyName(ctx, "machine", machine.GetName(), false)
	if err != nil {
		// Just log a warning message and dont fail operation
		logging.Warningf(ctx, "GetMachine - failed to query nics for machine %s: %s", machine.GetName(), err)
	}
	setNicsToMachine(machine, nics)

	// get Dracs for machine
	dracs, err := registration.QueryDracByPropertyName(ctx, "machine", machine.GetName(), false)
	if err != nil {
		// Just log a warning message and dont fail operation
		logging.Warningf(ctx, "GetMachine - failed to query dracs for machine %s: %s", machine.GetName(), err)
	}
	if dracs != nil && len(dracs) > 0 {
		setDracToMachine(machine, dracs[0])
	}
}

func setNicsToMachine(machine *ufspb.Machine, nics []*ufspb.Nic) {
	if len(nics) <= 0 {
		return
	}
	if machine.GetChromeBrowserMachine() == nil {
		machine.Device = &ufspb.Machine_ChromeBrowserMachine{
			ChromeBrowserMachine: &ufspb.ChromeBrowserMachine{
				NicObjects: nics,
			},
		}
	} else {
		machine.GetChromeBrowserMachine().NicObjects = nics
	}
}

func setDracToMachine(machine *ufspb.Machine, drac *ufspb.Drac) {
	if drac == nil {
		return
	}
	if machine.GetChromeBrowserMachine() == nil {
		machine.Device = &ufspb.Machine_ChromeBrowserMachine{
			ChromeBrowserMachine: &ufspb.ChromeBrowserMachine{
				DracObject: drac,
			},
		}
	} else {
		machine.GetChromeBrowserMachine().DracObject = drac
	}
}
