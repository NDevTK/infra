// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"

	ufspb "infra/unifiedfleet/api/v1/proto"
	"infra/unifiedfleet/app/model/history"
	"infra/unifiedfleet/app/util"
)

// HistoryClient is a client for managing change histories
type HistoryClient struct {
	netUdt  *networkUpdater
	stUdt   *stateUpdater
	changes []*ufspb.ChangeEvent
	msgs    []*history.SnapshotMsgEntity
	errs    []string
}

// SaveChangeEvents saves change events to database
func (hc *HistoryClient) SaveChangeEvents(ctx context.Context) error {
	if err := hc.saveMsgs(ctx); err != nil {
		return err
	}
	changes := hc.changes
	if hc.netUdt != nil {
		changes = append(changes, hc.netUdt.Changes...)
	}
	if hc.stUdt != nil {
		changes = append(changes, hc.stUdt.Changes...)
	}
	logging.Debugf(ctx, "Logging %d changes", len(changes))
	user := auth.CurrentUser(ctx)
	for _, c := range changes {
		c.UserEmail = user.Email
	}
	_, err := history.CreateBatchChangeEvents(ctx, changes)
	if err != nil {
		logging.Debugf(ctx, "fail to log changes: %s", err.Error())
		return err
	}
	logging.Debugf(ctx, "Finish logging change events successfully")
	return nil
}

func (hc *HistoryClient) saveMsgs(ctx context.Context) error {
	msgs := hc.msgs
	if hc.netUdt != nil {
		msgs = append(msgs, hc.netUdt.Msgs...)
	}
	if hc.stUdt != nil {
		msgs = append(msgs, hc.stUdt.Msgs...)
	}
	logging.Debugf(ctx, "Logging %d snapshot msgs", len(msgs))
	if err := history.BatchUpdateSnapshotMsg(ctx, msgs); err != nil {
		return err
	}
	logging.Debugf(ctx, "Finish logging snapshot msgs successfully")
	return nil
}

const (
	// LifeCycleRegistration indicates the registration of a resource.
	LifeCycleRegistration = "REGISTRATION"

	// LifeCycleDeployment indicates the deployment of a device.
	LifeCycleDeployment = "DEPLOYMENT"

	// LifeCycleDecomm indicates the decommission of a device.
	LifeCycleDecomm = "DECOMMISSION"

	// LifeCycleRetire indicates the retirement of a device.
	LifeCycleRetire = "RETIREMENT"
)

// logLifeCycle logs the life cycle event of a ChromeOSDevice.
func logLifeCycle(resourceName, resourceType, lifeCycleType string) (changes []*ufspb.ChangeEvent) {
	return []*ufspb.ChangeEvent{
		{
			Name:       resourceName,
			EventLabel: resourceType,
			OldValue:   lifeCycleType,
			NewValue:   lifeCycleType,
		},
	}
}

func (hc *HistoryClient) logMsgEntity(resourceName string, delete bool, pm proto.Message) {
	if e, err := history.NewSnapshotMsgEntity(resourceName, delete, pm); err != nil {
		hc.errs = append(hc.errs, err.Error())
	} else {
		hc.msgs = append(hc.msgs, e)
	}
}

// LogMachineChanges logs the change of the given machine.
func (hc *HistoryClient) LogMachineChanges(oldData *ufspb.Machine, newData *ufspb.Machine) {
	if oldData == nil && newData == nil {
		return
	}
	resourceName := util.AddPrefix(util.MachineCollection, newData.GetName())
	if newData == nil {
		resourceName = util.AddPrefix(util.MachineCollection, oldData.GetName())
	}
	if oldData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "machine", LifeCycleRegistration)...)
		hc.logMsgEntity(resourceName, false, newData)
		return
	}
	if newData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "machine", LifeCycleRetire)...)
		oldData.UpdateTime = ptypes.TimestampNow()
		hc.logMsgEntity(resourceName, true, oldData)
		return
	}
	hc.changes = append(hc.changes, logCommon(resourceName, "machine.serial_number", oldData.GetSerialNumber(), newData.GetSerialNumber())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "machine.location", oldData.GetLocation(), newData.GetLocation())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "machine.realm", oldData.GetRealm(), newData.GetRealm())...)
	if newData.GetChromeBrowserMachine() != nil {
		hc.changes = append(hc.changes, logChromeBrowserMachine(resourceName, oldData.GetChromeBrowserMachine(), newData.GetChromeBrowserMachine())...)
	}
	hc.logMsgEntity(resourceName, false, newData)
}

// LogMachineLSEChanges logs the change of the given machine lse.
func (hc *HistoryClient) LogMachineLSEChanges(oldData *ufspb.MachineLSE, newData *ufspb.MachineLSE) {
	if oldData == nil && newData == nil {
		return
	}
	resourceName := util.AddPrefix(util.HostCollection, newData.GetName())
	if newData == nil {
		resourceName = util.AddPrefix(util.HostCollection, oldData.GetName())
	}
	if oldData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "machine_lse", LifeCycleRegistration)...)
		hc.logMsgEntity(resourceName, false, newData)
		return
	}
	if newData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "machine_lse", LifeCycleRetire)...)
		oldData.UpdateTime = ptypes.TimestampNow()
		hc.logMsgEntity(resourceName, true, oldData)
		return
	}
	hc.changes = append(hc.changes, logCommon(resourceName, "machine_lse.machine_lse_prototype", oldData.GetMachineLsePrototype(), newData.GetMachineLsePrototype())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "machine_lse.hostname", oldData.GetHostname(), newData.GetHostname())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "machine_lse.machines", oldData.GetMachines(), newData.GetMachines())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "machine_lse.nic", oldData.GetNic(), newData.GetNic())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "machine_lse.lab", approxLab(oldData.GetLab()), approxLab(newData.GetLab()))...)
	hc.changes = append(hc.changes, logCommon(resourceName, "machine_lse.rack", oldData.GetRack(), newData.GetRack())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "machine_lse.manufacturer", oldData.GetManufacturer(), newData.GetManufacturer())...)
	if newData.GetChromeBrowserMachineLse() != nil {
		hc.changes = append(hc.changes, logChromeBrowserMachineLse(resourceName, oldData.GetChromeBrowserMachineLse(), newData.GetChromeBrowserMachineLse())...)
	}
	hc.logMsgEntity(resourceName, false, newData)
}

// LogVMChanges logs the change of the given vms.
func (hc *HistoryClient) LogVMChanges(oldData *ufspb.VM, newData *ufspb.VM) {
	if oldData == nil && newData == nil {
		return
	}
	resourceName := util.AddPrefix(util.VMCollection, newData.GetName())
	if newData == nil {
		resourceName = util.AddPrefix(util.VMCollection, oldData.GetName())
	}
	if oldData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "vm", LifeCycleRegistration)...)
		hc.logMsgEntity(resourceName, false, newData)
		return
	}
	if newData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "vm", LifeCycleRetire)...)
		oldData.UpdateTime = ptypes.TimestampNow()
		hc.logMsgEntity(resourceName, true, oldData)
		return
	}
	if oldData.GetName() != newData.GetName() {
		return
	}
	hc.changes = append(hc.changes, logCommon(resourceName, "vm.mac_address", oldData.GetMacAddress(), newData.GetMacAddress())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "vm.os_version", oldData.GetOsVersion(), newData.GetOsVersion())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "vm.vlan", oldData.GetVlan(), newData.GetVlan())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "vm.lab", approxLab(oldData.GetLab()), approxLab(newData.GetLab()))...)
	hc.changes = append(hc.changes, logCommon(resourceName, "vm.machine_lse_id", oldData.GetMachineLseId(), newData.GetMachineLseId())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "vm.state", approxState(oldData.GetState()), approxState(newData.GetState()))...)
	hc.logMsgEntity(resourceName, false, newData)
}

// LogRackChanges logs the change of the given rack.
func (hc *HistoryClient) LogRackChanges(oldData *ufspb.Rack, newData *ufspb.Rack) {
	if oldData == nil && newData == nil {
		return
	}
	resourceName := util.AddPrefix(util.RackCollection, newData.GetName())
	if newData == nil {
		resourceName = util.AddPrefix(util.RackCollection, oldData.GetName())
	}
	if oldData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "rack", LifeCycleRegistration)...)
		hc.logMsgEntity(resourceName, false, newData)
		return
	}
	if newData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "rack", LifeCycleRetire)...)
		oldData.UpdateTime = ptypes.TimestampNow()
		hc.logMsgEntity(resourceName, true, oldData)
		return
	}
	hc.changes = append(hc.changes, logCommon(resourceName, "rack.location", oldData.GetLocation(), newData.GetLocation())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "rack.capacity_ru", oldData.GetCapacityRu(), newData.GetCapacityRu())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "rack.realm", oldData.GetRealm(), newData.GetRealm())...)
	if newData.GetChromeBrowserRack() != nil {
		hc.changes = append(hc.changes, logChromeBrowserRack(resourceName, oldData.GetChromeBrowserRack(), newData.GetChromeBrowserRack())...)
	}
	hc.logMsgEntity(resourceName, false, newData)
}

// LogNicChanges logs the change of the given nic.
func (hc *HistoryClient) LogNicChanges(oldData, newData *ufspb.Nic) {
	if oldData == nil && newData == nil {
		return
	}
	resourceName := util.AddPrefix(util.NicCollection, newData.GetName())
	if newData == nil {
		resourceName = util.AddPrefix(util.NicCollection, oldData.GetName())
	}
	if oldData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "nic", LifeCycleRegistration)...)
		hc.logMsgEntity(resourceName, false, newData)
		return
	}
	if newData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "nic", LifeCycleRetire)...)
		oldData.UpdateTime = ptypes.TimestampNow()
		hc.logMsgEntity(resourceName, true, oldData)
		return
	}
	hc.changes = append(hc.changes, logCommon(resourceName, "nic.mac_address", oldData.GetMacAddress(), newData.GetMacAddress())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "nic.machine", oldData.GetMachine(), newData.GetMachine())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "nic.lab", approxLab(oldData.GetLab()), approxLab(newData.GetLab()))...)
	hc.changes = append(hc.changes, logCommon(resourceName, "nic.rack", oldData.GetRack(), newData.GetRack())...)
	hc.changes = append(hc.changes, logSwitchInterface(resourceName, oldData.GetSwitchInterface(), newData.GetSwitchInterface())...)
	hc.logMsgEntity(resourceName, false, newData)
}

// LogDracChanges logs the change of the given drac.
func (hc *HistoryClient) LogDracChanges(oldData, newData *ufspb.Drac) {
	if oldData == nil && newData == nil {
		return
	}
	resourceName := util.AddPrefix(util.DracCollection, newData.GetName())
	if newData == nil {
		resourceName = util.AddPrefix(util.DracCollection, oldData.GetName())
	}
	if oldData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "drac", LifeCycleRegistration)...)
		hc.logMsgEntity(resourceName, false, newData)
		return
	}
	if newData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "drac", LifeCycleRetire)...)
		oldData.UpdateTime = ptypes.TimestampNow()
		hc.logMsgEntity(resourceName, true, oldData)
		return
	}
	hc.changes = append(hc.changes, logCommon(resourceName, "drac.display_name", oldData.GetDisplayName(), newData.GetDisplayName())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "drac.mac_address", oldData.GetMacAddress(), newData.GetMacAddress())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "drac.machine", oldData.GetMachine(), newData.GetMachine())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "drac.lab", approxLab(oldData.GetLab()), approxLab(newData.GetLab()))...)
	hc.changes = append(hc.changes, logCommon(resourceName, "drac.rack", oldData.GetRack(), newData.GetRack())...)
	hc.changes = append(hc.changes, logSwitchInterface(resourceName, oldData.GetSwitchInterface(), newData.GetSwitchInterface())...)
	hc.logMsgEntity(resourceName, false, newData)
}

// LogKVMChanges logs the change of the given kvm.
func (hc *HistoryClient) LogKVMChanges(oldData, newData *ufspb.KVM) {
	if oldData == nil && newData == nil {
		return
	}
	resourceName := util.AddPrefix(util.KVMCollection, newData.GetName())
	if newData == nil {
		resourceName = util.AddPrefix(util.KVMCollection, oldData.GetName())
	}
	if oldData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "kvm", LifeCycleRegistration)...)
		hc.logMsgEntity(resourceName, false, newData)
		return
	}
	if newData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "kvm", LifeCycleRetire)...)
		oldData.UpdateTime = ptypes.TimestampNow()
		hc.logMsgEntity(resourceName, true, oldData)
		return
	}
	hc.changes = append(hc.changes, logCommon(resourceName, "kvm.mac_address", oldData.GetMacAddress(), newData.GetMacAddress())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "kvm.chrome_platform", oldData.GetChromePlatform(), newData.GetChromePlatform())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "kvm.capacity_port", oldData.GetCapacityPort(), newData.GetCapacityPort())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "kvm.lab", approxLab(oldData.GetLab()), approxLab(newData.GetLab()))...)
	hc.changes = append(hc.changes, logCommon(resourceName, "kvm.rack", oldData.GetRack(), newData.GetRack())...)
	hc.logMsgEntity(resourceName, false, newData)
}

// LogSwitchChanges logs the change of the given switch.
func (hc *HistoryClient) LogSwitchChanges(oldData, newData *ufspb.Switch) {
	if oldData == nil && newData == nil {
		return
	}
	resourceName := util.AddPrefix(util.SwitchCollection, newData.GetName())
	if newData == nil {
		resourceName = util.AddPrefix(util.SwitchCollection, oldData.GetName())
	}
	if oldData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "switch", LifeCycleRegistration)...)
		hc.logMsgEntity(resourceName, false, newData)
		return
	}
	if newData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "switch", LifeCycleRetire)...)
		oldData.UpdateTime = ptypes.TimestampNow()
		hc.logMsgEntity(resourceName, true, oldData)
		return
	}
	hc.changes = append(hc.changes, logCommon(resourceName, "switch.description", oldData.GetDescription(), newData.GetDescription())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "switch.capacity_port", oldData.GetCapacityPort(), newData.GetCapacityPort())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "switch.lab", approxLab(oldData.GetLab()), approxLab(newData.GetLab()))...)
	hc.changes = append(hc.changes, logCommon(resourceName, "switch.rack", oldData.GetRack(), newData.GetRack())...)
	hc.logMsgEntity(resourceName, false, newData)
}

// LogRPMChanges logs the change of the given rpms.
func (hc *HistoryClient) LogRPMChanges(oldData, newData *ufspb.RPM) {
	if oldData == nil && newData == nil {
		return
	}
	resourceName := util.AddPrefix(util.RPMCollection, newData.GetName())
	if newData == nil {
		resourceName = util.AddPrefix(util.RPMCollection, oldData.GetName())
	}
	if oldData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "rpm", LifeCycleRegistration)...)
		hc.logMsgEntity(resourceName, false, newData)
		return
	}
	if newData == nil {
		hc.changes = append(hc.changes, logLifeCycle(resourceName, "rpm", LifeCycleRetire)...)
		oldData.UpdateTime = ptypes.TimestampNow()
		hc.logMsgEntity(resourceName, true, oldData)
		return
	}
	hc.changes = append(hc.changes, logCommon(resourceName, "rpm.mac_address", oldData.GetMacAddress(), newData.GetMacAddress())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "rpm.capacity_port", oldData.GetCapacityPort(), newData.GetCapacityPort())...)
	hc.changes = append(hc.changes, logCommon(resourceName, "rpm.lab", approxLab(oldData.GetLab()), approxLab(newData.GetLab()))...)
	hc.changes = append(hc.changes, logCommon(resourceName, "rpm.rack", oldData.GetRack(), newData.GetRack())...)
	hc.logMsgEntity(resourceName, false, newData)
}

// LogDHCPChanges logs the change of the given dhcp.
func LogDHCPChanges(oldData, newData *ufspb.DHCPConfig) ([]*ufspb.ChangeEvent, *history.SnapshotMsgEntity) {
	changes := make([]*ufspb.ChangeEvent, 0)
	if oldData == nil && newData == nil {
		return changes, nil
	}
	resourceName := util.AddPrefix(util.DHCPCollection, newData.GetHostname())
	if newData == nil {
		resourceName = util.AddPrefix(util.DHCPCollection, oldData.GetHostname())
	}
	if oldData == nil {
		e, _ := history.NewSnapshotMsgEntity(resourceName, false, newData)
		return logCommon(resourceName, "dhcp_config.ip", "", newData.GetIp()), e
	}
	if newData == nil {
		oldData.UpdateTime = ptypes.TimestampNow()
		e, _ := history.NewSnapshotMsgEntity(resourceName, true, oldData)
		return logCommon(resourceName, "dhcp_config.ip", oldData.GetIp(), ""), e
	}
	changes = append(changes, logCommon(resourceName, "dhcp_config.mac_address", oldData.GetMacAddress(), newData.GetMacAddress())...)
	changes = append(changes, logCommon(resourceName, "dhcp_config.ip", oldData.GetIp(), newData.GetIp())...)
	changes = append(changes, logCommon(resourceName, "dhcp_config.vlan", oldData.GetVlan(), newData.GetVlan())...)
	e, _ := history.NewSnapshotMsgEntity(resourceName, false, newData)
	return changes, e
}

// LogIPChanges logs the change of the given ip.
func LogIPChanges(oldData, newData *ufspb.IP) []*ufspb.ChangeEvent {
	changes := make([]*ufspb.ChangeEvent, 0)
	if oldData == nil && newData == nil {
		return changes
	}
	var resourceName string
	if oldData != nil {
		resourceName = util.AddPrefix(util.IPCollection, oldData.GetId())
	} else {
		resourceName = util.AddPrefix(util.IPCollection, newData.GetId())
	}
	return logCommon(resourceName, "ip.occupied", oldData.GetOccupied(), newData.GetOccupied())
}

// LogStateChanges logs the change of the given state record.
func LogStateChanges(oldData, newData *ufspb.StateRecord) ([]*ufspb.ChangeEvent, *history.SnapshotMsgEntity) {
	changes := make([]*ufspb.ChangeEvent, 0)
	if oldData == nil && newData == nil {
		return changes, nil
	}
	var resourceName string
	if oldData != nil {
		resourceName = util.AddPrefix(util.StateCollection, oldData.GetResourceName())
	} else {
		resourceName = util.AddPrefix(util.StateCollection, newData.GetResourceName())
	}
	var e *history.SnapshotMsgEntity
	if newData != nil {
		e, _ = history.NewSnapshotMsgEntity(resourceName, false, newData)
	} else {
		oldData.UpdateTime = ptypes.TimestampNow()
		e, _ = history.NewSnapshotMsgEntity(resourceName, true, oldData)
	}
	return logCommon(resourceName, "state_record.state", oldData.GetState().String(), newData.GetState().String()), e
}

func logChromeBrowserRack(resourceName string, oldData, newData *ufspb.ChromeBrowserRack) []*ufspb.ChangeEvent {
	changes := make([]*ufspb.ChangeEvent, 0)
	changes = append(changes, logCommon(resourceName, "rack.chrome_browser_rack.rpms", oldData.GetRpms(), newData.GetRpms())...)
	changes = append(changes, logCommon(resourceName, "rack.chrome_browser_rack.kvms", oldData.GetKvms(), newData.GetKvms())...)
	changes = append(changes, logCommon(resourceName, "rack.chrome_browser_rack.switches", oldData.GetSwitches(), newData.GetSwitches())...)
	return changes
}

func logChromeBrowserMachineLse(resourceName string, oldData, newData *ufspb.ChromeBrowserMachineLSE) []*ufspb.ChangeEvent {
	changes := make([]*ufspb.ChangeEvent, 0)
	changes = append(changes, logCommon(resourceName, "machine_lse.chrome_browser_machine_lse.vm_capacity", oldData.GetVmCapacity(), newData.GetVmCapacity())...)
	changes = append(changes, logCommon(resourceName, "machine_lse.chrome_browser_machine_lse.os_version", oldData.GetOsVersion(), newData.GetOsVersion())...)
	return changes
}

func logChromeBrowserMachine(resourceName string, oldData, newData *ufspb.ChromeBrowserMachine) []*ufspb.ChangeEvent {
	changes := make([]*ufspb.ChangeEvent, 0)
	changes = append(changes, logCommon(resourceName, "machine.chrome_browser_machine.display_name", oldData.GetDisplayName(), newData.GetDisplayName())...)
	changes = append(changes, logCommon(resourceName, "machine.chrome_browser_machine.chrome_platform", oldData.GetChromePlatform(), newData.GetChromePlatform())...)
	changes = append(changes, logCommon(resourceName, "machine.chrome_browser_machine.deployment_ticket", oldData.GetDeploymentTicket(), newData.GetDeploymentTicket())...)
	changes = append(changes, logKVMInterface(resourceName, oldData.GetKvmInterface(), newData.GetKvmInterface())...)
	changes = append(changes, logRPMInterface(resourceName, oldData.GetRpmInterface(), newData.GetRpmInterface())...)
	return changes
}

func logKVMInterface(resourceName string, oldData, newData *ufspb.KVMInterface) []*ufspb.ChangeEvent {
	changes := make([]*ufspb.ChangeEvent, 0)
	changes = append(changes, logCommon(resourceName, "machine.chrome_browser_machine.kvm_interface.kvm", oldData.GetKvm(), newData.GetKvm())...)
	return append(changes, logCommon(resourceName, "machine.chrome_browser_machine.kvm_interface.port", oldData.GetPort(), newData.GetPort())...)
}

func logRPMInterface(resourceName string, oldData, newData *ufspb.RPMInterface) []*ufspb.ChangeEvent {
	changes := make([]*ufspb.ChangeEvent, 0)
	changes = append(changes, logCommon(resourceName, "machine.chrome_browser_machine.rpm_interface.rpm", oldData.GetRpm(), newData.GetRpm())...)
	return append(changes, logCommon(resourceName, "machine.chrome_browser_machine.rpm_interface.port", oldData.GetPort(), newData.GetPort())...)
}

func logSwitchInterface(resourceName string, oldData, newData *ufspb.SwitchInterface) []*ufspb.ChangeEvent {
	changes := make([]*ufspb.ChangeEvent, 0)
	changes = append(changes, logCommon(resourceName, "switch_interface.switch", oldData.GetSwitch(), newData.GetSwitch())...)
	return append(changes, logCommon(resourceName, "switch_interface.port", oldData.GetPort(), newData.GetPort())...)
}

func logCommon(resourceName, label string, oldValue interface{}, newValue interface{}) []*ufspb.ChangeEvent {
	oldValueStr := fmt.Sprintf("%v", oldValue)
	newValueStr := fmt.Sprintf("%v", newValue)
	if oldValueStr == newValueStr {
		return nil
	}
	return []*ufspb.ChangeEvent{
		{
			Name:       resourceName,
			EventLabel: label,
			OldValue:   oldValueStr,
			NewValue:   newValueStr,
		},
	}
}

func approxLab(lab string) string {
	if lab == "" {
		return ufspb.Lab_LAB_UNSPECIFIED.String()
	}
	return lab
}

func approxState(s string) string {
	if s == "" {
		return ufspb.State_STATE_UNSPECIFIED.String()
	}
	return s
}
