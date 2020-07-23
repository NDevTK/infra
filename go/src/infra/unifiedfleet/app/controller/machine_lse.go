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
	crimson "go.chromium.org/luci/machine-db/api/crimson/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	invV2Api "infra/appengine/cros/lab_inventory/api/v1"
	ufspb "infra/unifiedfleet/api/v1/proto"
	chromeosLab "infra/unifiedfleet/api/v1/proto/chromeos/lab"
	"infra/unifiedfleet/app/model/configuration"
	ufsds "infra/unifiedfleet/app/model/datastore"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/util"
)

// CreateMachineLSE creates a new machinelse in datastore.
func CreateMachineLSE(ctx context.Context, machinelse *ufspb.MachineLSE, machineNames []string) (*ufspb.MachineLSE, error) {
	// 1. MachineLSE name and hostname must always be the same
	// Overwrite the name with hostname
	machinelse.Name = machinelse.GetHostname()

	// 2. If its a labstation, make the Hostname of the Labstation same as the machinelse name
	// Labstation hostname must be same as the machinelse hostname
	if machinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation() != nil {
		machinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().Hostname = machinelse.GetHostname()
	}

	// 3. Overwrite the OUTPUT_ONLY fields
	// This is output only field. User is not allowed to set its value.
	machinelse.Machines = machineNames

	// If its a DUT
	if machinelse.GetChromeosMachineLse().GetDeviceLse().GetDut() != nil {
		// ChromeOSMachineLSE for a DUT
		machinelse.GetChromeosMachineLse().GetDeviceLse().GetDut().Hostname = machinelse.GetHostname()
		return createChromeOSMachineLSEDUT(ctx, machinelse, machineNames)
	}

	// If its a Chrome browser host, ChromeOS server or a ChormeOS labstation
	// ChromeBrowserMachineLSE, ChromeOSMachineLSE for a Server and Labstation
	f := func(ctx context.Context) error {
		// 1. Validate input
		err := validateCreateMachineLSE(ctx, machinelse, machineNames)
		if err != nil {
			return errors.Annotate(err, "Validation error - Failed to create MachineLSE").Err()
		}

		// 2. Create the machinelse
		// we use this func as it is a non-atomic operation and can be used to
		// run within a transaction to make it atomic. Datastore doesnt allow
		// nested transactions.
		if _, err := inventory.BatchUpdateMachineLSEs(ctx, []*ufspb.MachineLSE{machinelse}); err != nil {
			return errors.Annotate(err, "Failed to BatchUpdate MachineLSEs %s", machinelse.Name).Err()
		}
		return nil
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to create machinelse in datastore: %s", err)
		return nil, err
	}
	SaveChangeEvents(ctx, LogMachineLSEChanges(nil, machinelse))
	return machinelse, nil
}

// UpdateMachineLSE updates machinelse in datastore.
func UpdateMachineLSE(ctx context.Context, machinelse *ufspb.MachineLSE, machineNames []string) (*ufspb.MachineLSE, error) {
	// 1. MachineLSEs name and hostname must always be the same
	// Overwrite the name with hostname
	machinelse.Name = machinelse.GetHostname()

	// 2. If its a labstation, make the Hostname of the Labstation same as the machinelse name
	// Labstation hostname must be same as the machinelse hostname
	if machinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation() != nil {
		machinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().Hostname = machinelse.GetHostname()
	}

	// 3. Overwrite the OUTPUT_ONLY fields
	// This is output only field. User is not allowed to set its value.
	if machineNames != nil {
		machinelse.Machines = machineNames
	}

	// If its a DUT
	if machinelse.GetChromeosMachineLse().GetDeviceLse().GetDut() != nil {
		// ChromeOSMachineLSE for a DUT
		machinelse.GetChromeosMachineLse().GetDeviceLse().GetDut().Hostname = machinelse.GetHostname()
		return updateChromeOSMachineLSEDUT(ctx, machinelse, machineNames)
	}

	// If its a Chrome browser host, ChromeOS server or a ChormeOS labstation
	// ChromeBrowserMachineLSE, ChromeOSMachineLSE for a Server and Labstation
	var oldMachinelse *ufspb.MachineLSE
	f := func(ctx context.Context) error {
		// 1. Validate the input
		err := validateUpdateMachineLSE(ctx, machinelse, machineNames)
		if err != nil {
			return errors.Annotate(err, "Validation error - Failed to update MachineLSE").Err()
		}

		// 2. Get the old machinelse
		// getting oldmachinelse for change history logging
		oldMachinelse, err = inventory.GetMachineLSE(ctx, machinelse.GetName())
		if err != nil {
			return errors.Annotate(err, "Failed to get old MachineLSE").Err()
		}

		if machineNames == nil {
			// 3. Overwrite the OUTPUT_ONLY fields
			// This is output only field. Assign already existing values.
			machinelse.Machines = oldMachinelse.GetMachines()
		}

		// 4. Update machinelse entry
		// we use this func as it is a non-atomic operation and can be used to
		// run within a transaction. Datastore doesnt allow nested transactions.
		if _, err = inventory.BatchUpdateMachineLSEs(ctx, []*ufspb.MachineLSE{machinelse}); err != nil {
			return errors.Annotate(err, "Unable to create MachineLSE %s", machinelse.Name).Err()
		}
		return nil
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to create entity in datastore: %s", err)
		return nil, err
	}
	SaveChangeEvents(ctx, LogMachineLSEChanges(oldMachinelse, machinelse))
	return machinelse, nil
}

// GetMachineLSE returns machinelse for the given id from datastore.
func GetMachineLSE(ctx context.Context, id string) (*ufspb.MachineLSE, error) {
	return inventory.GetMachineLSE(ctx, id)
}

// ListMachineLSEs lists the machinelses
func ListMachineLSEs(ctx context.Context, pageSize int32, pageToken string) ([]*ufspb.MachineLSE, string, error) {
	return inventory.ListMachineLSEs(ctx, pageSize, pageToken)
}

// DeleteMachineLSE deletes the machinelse in datastore
//
// For referential data intergrity,
// Delete if this MachineLSE is not referenced by other resources in the datastore.
// If there are any references, delete will be rejected and an error will be returned.
func DeleteMachineLSE(ctx context.Context, id string) error {
	err := validateDeleteMachineLSE(ctx, id)
	if err != nil {
		return err
	}
	existingMachinelse := &ufspb.MachineLSE{}
	f := func(ctx context.Context) error {
		existingMachinelse, err = inventory.GetMachineLSE(ctx, id)
		if err != nil {
			return err
		}
		// Check if it is a DUT MachineLSE and has servo info.
		// Update corresponding Labstation MachineLSE.
		if existingMachinelse.GetChromeosMachineLse().GetDeviceLse().GetDut() != nil {
			existingServo := existingMachinelse.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetServo()
			if existingServo != nil {
				// 1. remove the existingServo entry of DUT form existingLabstationMachinelse
				existingLabstationMachinelse, err := inventory.GetMachineLSE(ctx, existingServo.GetServoHostname())
				if err != nil {
					return err
				}
				removeServoEntryFromLabstation(existingServo, existingLabstationMachinelse)
				// BatchUpdate Labstation - Using Batch update and not UpdateMachineLSE,
				// because we cant have nested transaction in datastore
				_, err = inventory.BatchUpdateMachineLSEs(ctx, []*ufspb.MachineLSE{existingLabstationMachinelse})
				if err != nil {
					logging.Errorf(ctx, "Failed to BatchUpdate Labstation MachineLSE %s", err)
					return err
				}
			}
		}
		return inventory.DeleteMachineLSE(ctx, id)
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to delete MachineLSE in datastore: %s", err)
		return err
	}
	SaveChangeEvents(ctx, LogMachineLSEChanges(existingMachinelse, nil))
	return nil
}

// ImportOSMachineLSEs implements the logic of importing ChromeOS machine lses and related info to backend storage.
//
// The function will return:
//      * all of the results of the operations that already run
//      * the first error that it meets
//
// The function will stop at the very first error.
func ImportOSMachineLSEs(ctx context.Context, labConfigs []*invV2Api.ListCrosDevicesLabConfigResponse_LabConfig, pageSize int) (*ufsds.OpResults, error) {
	allRes := make(ufsds.OpResults, 0)
	logging.Debugf(ctx, "Importing the machine lse prototypes for OS lab")
	res, err := configuration.ImportMachineLSEPrototypes(ctx, util.GetOSMachineLSEPrototypes())
	if err != nil {
		return res, err
	}
	allRes = append(allRes, *res...)

	lses := util.ToOSMachineLSEs(labConfigs)
	deleteNonExistingMachineLSEs(ctx, lses, pageSize, "os-lab")
	logging.Debugf(ctx, "Importing %d lses", len(lses))
	for i := 0; ; i += pageSize {
		end := util.Min(i+pageSize, len(lses))
		res, err := inventory.ImportMachineLSEs(ctx, lses[i:end])
		allRes = append(allRes, *res...)
		if err != nil {
			return &allRes, err
		}
		if i+pageSize >= len(lses) {
			break
		}
	}
	return &allRes, nil
}

// ImportMachineLSEs implements the logic of importing machine lses and related info to backend storage.
//
// The function will return:
//      * all of the results of the operations that already run
//      * the first error that it meets
//
// The function will stop at the very first error.
func ImportMachineLSEs(ctx context.Context, hosts []*crimson.PhysicalHost, vms []*crimson.VM, pageSize int) (*ufsds.OpResults, error) {
	allRes := make(ufsds.OpResults, 0)
	logging.Debugf(ctx, "Importing the basic lse prototypes for browser lab")
	lps := []*ufspb.MachineLSEPrototype{
		{
			Name: "browser-lab:no-vm",
			VirtualRequirements: []*ufspb.VirtualRequirement{
				{
					VirtualType: ufspb.VirtualType_VIRTUAL_TYPE_VM,
					Min:         0,
					Max:         0,
				},
			},
		},
		{
			Name: "browser-lab:vm",
			VirtualRequirements: []*ufspb.VirtualRequirement{
				{
					VirtualType: ufspb.VirtualType_VIRTUAL_TYPE_VM,
					Min:         1,
					// A random number, not true.
					Max: 100,
				},
			},
		},
	}
	res, err := configuration.ImportMachineLSEPrototypes(ctx, lps)
	if err != nil {
		return res, err
	}
	allRes = append(allRes, *res...)

	lses, ips, dhcps := util.ToMachineLSEs(hosts, vms)
	deleteNonExistingMachineLSEs(ctx, lses, pageSize, "browser-lab")
	logging.Debugf(ctx, "Importing %d lses", len(lses))
	for i := 0; ; i += pageSize {
		end := util.Min(i+pageSize, len(lses))
		res, err := inventory.ImportMachineLSEs(ctx, lses[i:end])
		allRes = append(allRes, *res...)
		if err != nil {
			return &allRes, err
		}
		if i+pageSize >= len(lses) {
			break
		}
	}

	logging.Debugf(ctx, "Importing %d ips", len(ips))
	for i := 0; ; i += pageSize {
		end := util.Min(i+pageSize, len(ips))
		res, err := configuration.ImportIPs(ctx, ips[i:end])
		allRes = append(allRes, *res...)
		if err != nil {
			return &allRes, err
		}
		if i+pageSize >= len(ips) {
			break
		}
	}

	logging.Debugf(ctx, "Importing %d dhcps", len(dhcps))
	for i := 0; ; i += pageSize {
		end := util.Min(i+pageSize, len(dhcps))
		res, err := configuration.ImportDHCPConfigs(ctx, dhcps[i:end])
		allRes = append(allRes, *res...)
		if err != nil {
			return &allRes, err
		}
		if i+pageSize >= len(dhcps) {
			break
		}
	}
	return &allRes, nil
}

func deleteNonExistingMachineLSEs(ctx context.Context, machineLSEs []*ufspb.MachineLSE, pageSize int, lseType string) (*ufsds.OpResults, error) {
	resMap := make(map[string]bool)
	for _, r := range machineLSEs {
		resMap[r.GetName()] = true
	}
	resp, err := inventory.GetAllMachineLSEs(ctx)
	if err != nil {
		return nil, err
	}
	var toDelete []string
	var toDeleteDHCPHost []string
	for _, sr := range resp.Passed() {
		s := sr.Data.(*ufspb.MachineLSE)
		if lseType == "browser-lab" && s.GetChromeosMachineLse() != nil {
			continue
		}
		if lseType == "os-lab" && s.GetChromeBrowserMachineLse() != nil {
			continue
		}
		if _, ok := resMap[s.GetName()]; !ok {
			toDelete = append(toDelete, s.GetName())
			toDeleteDHCPHost = append(toDeleteDHCPHost, s.GetName())
		}
		if s.GetChromeBrowserMachineLse() != nil {
			for _, vm := range s.GetChromeBrowserMachineLse().GetVms() {
				toDeleteDHCPHost = append(toDeleteDHCPHost, vm.GetHostname())
			}
		}
	}
	logging.Debugf(ctx, "Deleting %d non-existing machine lses", len(toDelete))
	allRes := *deleteByPage(ctx, toDelete, pageSize, inventory.DeleteMachineLSEs)
	logging.Debugf(ctx, "Deleting %d non-existing host and vm-related dhcps", len(toDelete))
	allRes = append(allRes, *deleteByPage(ctx, toDelete, pageSize, configuration.DeleteDHCPs)...)
	return &allRes, nil
}

// createChromeOSMachineLSEDUT creates ChromeOSMachineLSE entities.
//
// creates one MachineLSE for DUT and updates another MachineLSE for the
// Labstation(with new Servo info from DUT)
func createChromeOSMachineLSEDUT(ctx context.Context, machinelse *ufspb.MachineLSE, machineNames []string) (*ufspb.MachineLSE, error) {
	f := func(ctx context.Context) error {
		machinelses := make([]*ufspb.MachineLSE, 0, 0)
		// A. Validate input
		err := validateCreateMachineLSE(ctx, machinelse, machineNames)
		if err != nil {
			return errors.Annotate(err, "Validation error - Failed to Create ChromeOSMachineLSEDUT").Err()
		}

		// B. Check if the DUT has Servo information.
		// Update Labstation MachineLSE with new Servo info.
		newServo := machinelse.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetServo()
		if newServo != nil {
			// 1. Check if the Labstation MachineLSE exists in the system.
			labstationMachinelse, err := getLabstationMachineLSE(ctx, newServo.GetServoHostname())
			if err != nil {
				return err
			}
			// 2. Check if the ServoHostName and ServoPort are already in use
			_, err = validateServoInfoForDUT(ctx, newServo, machinelse.GetName())
			if err != nil {
				return err
			}
			// 3. Update the Labstation MachineLSE with new Servo information.
			// Append the new Servo entry to the Labstation
			appendServoEntryToLabstation(newServo, labstationMachinelse)
			machinelses = append(machinelses, labstationMachinelse)
		}

		// C. BatchUpdate both DUT and Labstation
		_, err = inventory.BatchUpdateMachineLSEs(ctx, machinelses)
		if err != nil {
			return errors.Annotate(err, "Failed to BatchUpdate MachineLSEs").Err()
		}
		return nil
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to create MachineLSE in datastore: %s", err)
		return nil, err
	}
	SaveChangeEvents(ctx, LogMachineLSEChanges(nil, machinelse))
	return machinelse, nil
}

// updateChromeOSMachineLSEDUT updates ChromeOSMachineLSE entities.
//
// updates one MachineLSE for DUT and updates Labstation MachineLSE
// (with new Servo info from DUT). If DUT is connected to the same
// Labstation but different port, The servo entry in Labstation is updated.
// If DUT is connected to a different labstation, then old servo info of DUT
// is removed from old Labstation and new servo info from the DUT is added
// to the new labstation.
func updateChromeOSMachineLSEDUT(ctx context.Context, machinelse *ufspb.MachineLSE, machineNames []string) (*ufspb.MachineLSE, error) {
	f := func(ctx context.Context) error {
		// A. Validate the input
		err := validateUpdateMachineLSE(ctx, machinelse, machineNames)
		if err != nil {
			return errors.Annotate(err, "Validation error - Failed to update ChromeOSMachineLSEDUT").Err()
		}

		// B. Get the existing MachineLSE(DUT)
		oldMachinelse, err := inventory.GetMachineLSE(ctx, machinelse.GetName())
		if err != nil {
			return errors.Annotate(err, "Failed to get existing MachineLSE").Err()
		}

		if machineNames == nil {
			// C. Overwrite the OUTPUT_ONLY fields
			// This is output only field. Assign already existing values.
			machinelse.Machines = oldMachinelse.GetMachines()
		}

		machinelses := []*ufspb.MachineLSE{machinelse}

		// D. Check if the DUT has Servo information.
		// Update Labstation MachineLSE with new Servo info.
		newServo := machinelse.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetServo()
		if newServo != nil {
			// 1. Check if the Labstation MachineLSE exists in the system.
			newLabstationMachinelse, err := getLabstationMachineLSE(ctx, newServo.GetServoHostname())
			if err != nil {
				return err
			}
			// 2. Check if the ServoHostName and ServoPort are already in use
			_, err = validateServoInfoForDUT(ctx, newServo, machinelse.GetName())
			if err != nil {
				return err
			}
			// 3. Update the Labstation MachineLSE with new Servo information.
			oldServo := oldMachinelse.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetServo()
			// check if the DUT is connected to the same Labstation or different Labstation
			if newServo.GetServoHostname() == oldServo.GetServoHostname() {
				// DUT is connected to the same Labstation,
				// 1. replace the oldServo entry from the Labstation with the newServo entry
				replaceServoEntryInLabstation(oldServo, newServo, newLabstationMachinelse)
				machinelses = append(machinelses, newLabstationMachinelse)
			} else {
				// DUT is connected to a different Labstation,
				// 1. remove the oldServo entry of DUT form oldLabstationMachinelse
				oldLabstationMachinelse, err := inventory.GetMachineLSE(ctx, oldServo.GetServoHostname())
				if err != nil {
					return err
				}
				removeServoEntryFromLabstation(oldServo, oldLabstationMachinelse)
				machinelses = append(machinelses, oldLabstationMachinelse)
				// 2. Append the newServo entry of DUT to the newLabstationMachinelse
				appendServoEntryToLabstation(newServo, newLabstationMachinelse)
				machinelses = append(machinelses, newLabstationMachinelse)
			}
		}

		// E. BatchUpdate both DUT and Labstation/s
		_, err = inventory.BatchUpdateMachineLSEs(ctx, machinelses)
		if err != nil {
			logging.Errorf(ctx, "Failed to BatchUpdate ChromeOSMachineLSEDUTs %s", err)
			return err
		}
		return nil
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		logging.Errorf(ctx, "Failed to update MachineLSE DUT in datastore: %s", err)
		return nil, err
	}
	return machinelse, nil
}

// validateServoInfoForDUT Checks if the DUT Machinelse has ServoHostname and ServoPort
// already used by a different deployed DUT
func validateServoInfoForDUT(ctx context.Context, servo *chromeosLab.Servo, DUTHostname string) (string, error) {
	// Validating the DUT's servo is already occupied
	servoID := ufsds.GetServoID(servo.GetServoHostname(), servo.GetServoPort())
	dutMachinelses, err := inventory.QueryMachineLSEByPropertyName(ctx, "servo_id", servoID, true)
	if err != nil {
		return "", err
	}
	if dutMachinelses != nil && dutMachinelses[0].GetName() != DUTHostname {
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("Port: %d in Labstation: %s is already "+
			"in use by DUT: %s. Please provide a different ServoPort.\n",
			servo.GetServoPort(), servo.GetServoHostname(), dutMachinelses[0].GetName()))
		logging.Errorf(ctx, errorMsg.String())
		return dutMachinelses[0].GetName(), status.Errorf(codes.FailedPrecondition, errorMsg.String())
	}
	if dutMachinelses != nil {
		return dutMachinelses[0].GetName(), nil
	}
	return "", nil
}

// getLabstationMachineLSE get the Labstation MachineLSE
func getLabstationMachineLSE(ctx context.Context, labstationMachinelseName string) (*ufspb.MachineLSE, error) {
	labstationMachinelse, err := inventory.GetMachineLSE(ctx, labstationMachinelseName)
	if status.Code(err) == codes.Internal {
		return nil, err
	}
	if labstationMachinelse == nil {
		// There is no Labstation MachineLSE existing in the system
		errorMsg := fmt.Sprintf("Labstation %s not found in the system. "+
			"Please deploy the Labstation %s before deploying the DUT.",
			labstationMachinelseName, labstationMachinelseName)
		logging.Errorf(ctx, errorMsg)
		return nil, status.Errorf(codes.FailedPrecondition, errorMsg)
	}
	return labstationMachinelse, nil
}

// appendServoEntryToLabstation append servo entry to the Labstation
func appendServoEntryToLabstation(newServo *chromeosLab.Servo, labstationMachinelse *ufspb.MachineLSE) {
	existingServos := labstationMachinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().GetServos()
	existingServos = append(existingServos, newServo)
	labstationMachinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().Servos = existingServos
}

// replaceServoEntryInLabstation replaces oldServo entry with newServo entry in the Labstation
func replaceServoEntryInLabstation(oldServo, newServo *chromeosLab.Servo, labstationMachinelse *ufspb.MachineLSE) {
	servos := labstationMachinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().GetServos()
	for i, s := range servos {
		if s.GetServoHostname() == oldServo.GetServoHostname() && s.GetServoPort() == oldServo.GetServoPort() {
			servos[i] = newServo
			break
		}
	}
	labstationMachinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().Servos = servos
}

// removeServoEntryFromLabstation removes servo entry from the Labstation
func removeServoEntryFromLabstation(servo *chromeosLab.Servo, labstationMachinelse *ufspb.MachineLSE) {
	servos := labstationMachinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().GetServos()
	for i, s := range servos {
		if s.GetServoHostname() == servo.GetServoHostname() && s.GetServoPort() == servo.GetServoPort() {
			servos[i] = servos[len(servos)-1]
			servos = servos[:len(servos)-1]
			break
		}
	}
	labstationMachinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().Servos = servos
}

// validateCreateMachineLSE validates if a machinelse can be created in the datastore.
func validateCreateMachineLSE(ctx context.Context, machinelse *ufspb.MachineLSE, machineNames []string) error {
	//1. Check for servos for Labstation deployment
	if machinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation() != nil {
		newServos := machinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().GetServos()
		if len(newServos) != 0 {
			return status.Errorf(codes.FailedPrecondition, "Servos are not allowed "+
				"to be added in deploying labstations")
		}
	}

	// 2. Check if machinelse already exists
	if err := resourceAlreadyExists(ctx, []*Resource{GetMachineLSEResource(machinelse.Name)}, nil); err != nil {
		return err
	}

	// 3. Check if resources does not exist
	var resourcesNotfound []*Resource
	// Aggregate resource to check if machines does not exist
	for _, machineName := range machineNames {
		resourcesNotfound = append(resourcesNotfound, GetMachineResource(machineName))
	}
	// Aggregate resources referenced by the machinelse to check if they do not exist
	if machineLSEPrototypeID := machinelse.GetMachineLsePrototype(); machineLSEPrototypeID != "" {
		resourcesNotfound = append(resourcesNotfound, GetMachineLSEProtoTypeResource(machineLSEPrototypeID))
	}
	if vlanID := machinelse.GetChromeosMachineLse().GetServerLse().GetSupportedRestrictedVlan(); vlanID != "" {
		resourcesNotfound = append(resourcesNotfound, GetVlanResource(vlanID))
	}
	if rpmID := machinelse.GetChromeosMachineLse().GetDeviceLse().GetRpmInterface().GetRpm(); rpmID != "" {
		resourcesNotfound = append(resourcesNotfound, GetRPMResource(rpmID))
	}
	if err := ResourceExist(ctx, resourcesNotfound, nil); err != nil {
		return err
	}

	// 4. Check if any machine is already associated with another MachineLSE
	// A machine cannot be associated with multiple hosts/machinelses
	for _, machineName := range machineNames {
		machinelses, err := inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", machineName, true)
		if err != nil {
			return errors.Annotate(err, "Failed to query machinelses for machine %s", machineName).Err()
		}
		if len(machinelses) > 1 {
			logging.Warningf(ctx, "More than one machinelse associated with the "+
				"machine %s. Data discrepancy error.\n", machineName)
		}
		if len(machinelses) > 0 {
			var errorMsg strings.Builder
			errorMsg.WriteString(fmt.Sprintf("Host %s cannot be created because "+
				"there are other hosts which are referring this machine %s. "+
				"A machine cannot be associated with multiple hosts.",
				machinelse.Name, machineName))
			errorMsg.WriteString(fmt.Sprintf("\nHosts referring the machine %s:\n", machineName))
			for _, mlse := range machinelses {
				errorMsg.WriteString(mlse.Name + ", ")
			}
			errorMsg.WriteString(fmt.Sprintf("\nPlease delete the hosts and then "+
				"add this host %s.\n", machinelse.Name))
			logging.Errorf(ctx, errorMsg.String())
			return status.Errorf(codes.FailedPrecondition, errorMsg.String())
		}
	}
	return nil
}

// validateUpdateMachineLSE validates if a machinelse can be updated in the datastore.
func validateUpdateMachineLSE(ctx context.Context, machinelse *ufspb.MachineLSE, machineNames []string) error {
	// 1. This check is only for a Labstation
	// Check if labstation MachineLSE is updating any servo information
	// It is also not allowed to update the servo Hostname and servo Port of any servo.
	// Servo info is added/updated into Labstation only when a DUT(MachineLSE) is added/updated
	if machinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation() != nil {
		newServos := machinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().GetServos()
		existingLabstationMachinelse, err := inventory.GetMachineLSE(ctx, machinelse.GetName())
		if status.Code(err) == codes.Internal {
			return err
		}
		if existingLabstationMachinelse != nil {
			existingServos := existingLabstationMachinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().GetServos()
			if !testServoEq(newServos, existingServos) {
				return status.Errorf(codes.FailedPrecondition, "Servos are not allowed to be updated in redeploying labstations")
			}
		}
	}

	// 2. Check if resources does not exist
	// Aggregate resource to check if machinelse does not exist
	resourcesNotfound := []*Resource{GetMachineLSEResource(machinelse.Name)}
	// Aggregate resource to check if machines does not exist
	for _, machineName := range machineNames {
		resourcesNotfound = append(resourcesNotfound, GetMachineResource(machineName))
	}
	// Aggregate resources referenced by the machinelse to check if they do not exist
	if machineLSEPrototypeID := machinelse.GetMachineLsePrototype(); machineLSEPrototypeID != "" {
		resourcesNotfound = append(resourcesNotfound, GetMachineLSEProtoTypeResource(machineLSEPrototypeID))
	}
	if vlanID := machinelse.GetChromeosMachineLse().GetServerLse().GetSupportedRestrictedVlan(); vlanID != "" {
		resourcesNotfound = append(resourcesNotfound, GetVlanResource(vlanID))
	}
	if rpmID := machinelse.GetChromeosMachineLse().GetDeviceLse().GetRpmInterface().GetRpm(); rpmID != "" {
		resourcesNotfound = append(resourcesNotfound, GetRPMResource(rpmID))
	}
	if err := ResourceExist(ctx, resourcesNotfound, nil); err != nil {
		return err
	}

	// 3. Check if any machine is already associated with another MachineLSE
	// A machine cannot be associated with multiple hosts/machinelses
	for _, machineName := range machineNames {
		machinelses, err := inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", machineName, true)
		if err != nil {
			return errors.Annotate(err, "Failed to query machinelses for machine %s", machineName).Err()
		}
		if len(machinelses) == 1 && machinelses[0].GetName() != machinelse.Name {
			errorMsg := fmt.Sprintf("Host %s cannot be updated because "+
				"there is another host %s which is referring this machine %s. "+
				"A machine cannot be associated with multiple hosts. "+
				"Please delete the other host and then update this host.\n",
				machinelse.Name, machinelses[0].Name, machineName)
			return status.Errorf(codes.FailedPrecondition, errorMsg)
		}
		if len(machinelses) > 1 {
			var errorMsg strings.Builder
			errorMsg.WriteString(fmt.Sprintf("More than one host associated "+
				"with the machine %s. Data discrepancy error. Host %s cannot be "+
				"updated because there are other hosts which are referring this "+
				"machine %s. A machine cannot be associated with multiple hosts. ",
				machineName, machinelse.Name, machineName))
			errorMsg.WriteString(fmt.Sprintf("Hosts referring the machine %s:\n", machineName))
			for _, mlse := range machinelses {
				errorMsg.WriteString(mlse.Name + ", ")
			}
			errorMsg.WriteString(fmt.Sprintf("\nPlease delete the hosts and then "+
				"add this host %s.\n", machinelse.Name))
			logging.Errorf(ctx, errorMsg.String())
			return status.Errorf(codes.FailedPrecondition, errorMsg.String())
		}
	}
	return nil
}

// validateDeleteMachineLSE validates if a MachineLSE can be deleted
func validateDeleteMachineLSE(ctx context.Context, id string) error {
	existingMachinelse, err := inventory.GetMachineLSE(ctx, id)
	if err != nil {
		return err
	}
	if existingMachinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation() != nil {
		existingServos := existingMachinelse.GetChromeosMachineLse().GetDeviceLse().GetLabstation().GetServos()
		nonDeletedHosts := make([]string, 0, len(existingServos))
		for _, servo := range existingServos {
			dutHostName, err := validateServoInfoForDUT(ctx, servo, "")
			if err != nil {
				nonDeletedHosts = append(nonDeletedHosts, dutHostName)
			}
		}
		if len(nonDeletedHosts) != 0 {
			errorMsg := fmt.Sprintf("Labstation %s cannot be deleted because "+
				"there are servos in the labstation referenced by other DUTs: %s.",
				id, strings.Join(nonDeletedHosts, ", "))
			logging.Errorf(ctx, errorMsg)
			return status.Errorf(codes.FailedPrecondition, errorMsg)
		}
	}
	return nil
}
