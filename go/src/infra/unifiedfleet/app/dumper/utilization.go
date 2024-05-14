// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"

	"infra/cros/dutstate"
	invV1 "infra/libs/skylab/inventory"
	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"
	"infra/unifiedfleet/app/util"
)

// inventoryCounter collects number of DUTs per bucket and status.
type inventoryCounter map[*bucket]int

// suMetric is the metric name for scheduling unit count.
var suMetric = metric.NewInt(
	"chromeos/skylab/inventory/scheduling_unit_count",
	"The number of scheduling units in a given bucket",
	nil,
	field.String("board"),
	field.String("model"),
	field.String("pool"),
	field.String("environment"),
	field.String("zone"),
	field.String("status"),
)

// reportUFSInventoryCronHandler push the ufs duts metrics to tsmon
func reportUFSInventoryCronHandler(ctx context.Context) (err error) {
	logging.Infof(ctx, "Reporting UFS inventory DUT metrics")
	env := config.Get(ctx).SelfStorageBucket
	// Set namespace to OS to get only MachineLSEs for chromeOS.
	ctx, err = util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	if err != nil {
		return err
	}
	// Get all the MachineLSEs
	lses, err := getAllMachineLSEs(ctx, false)
	if err != nil {
		return err
	}
	idTolseMap := make(map[string]*ufspb.MachineLSE, 0)
	for _, lse := range lses {
		idTolseMap[lse.GetName()] = lse
	}
	// Get all Machines
	machines, err := getAllMachines(ctx, false)
	if err != nil {
		return err
	}
	idTomachineMap := make(map[string]*ufspb.Machine, 0)
	for _, machine := range machines {
		idTomachineMap[machine.GetName()] = machine
	}
	sUnits, err := getAllSchedulingUnits(ctx, false)
	if err != nil {
		return err
	}
	c := make(inventoryCounter)
	// Map for MachineLSEs associated with SchedulingUnit for easy search.
	lseInSUnitMap := make(map[string]bool)
	for _, su := range sUnits {
		if len(su.GetMachineLSEs()) > 0 {
			suLses := make([]*ufspb.MachineLSE, len(su.GetMachineLSEs()))
			for i, lseID := range su.GetMachineLSEs() {
				suLses[i] = idTolseMap[lseID]
			}

			b, err := getBucketForSchedulingUnit(su, suLses, idTomachineMap, env)
			if err != nil {
				logging.Warningf(ctx, err.Error())
				continue
			}
			c[b]++
			for _, lseName := range su.GetMachineLSEs() {
				lseInSUnitMap[lseName] = true
			}
		}
	}
	for _, lse := range lses {
		name := lse.GetName()
		if lseInSUnitMap[name] {
			continue
		}
		machine, err := getMachineForLse(lse, idTomachineMap)
		if err != nil {
			logging.Warningf(ctx, err.Error())
			continue
		}
		b := getBucketForDevice(lse, machine, env)
		c[b]++
	}
	logging.Infof(ctx, "report UFS inventory metrics for %d devices", len(c))
	c.Report(ctx)
	return nil
}

func (c inventoryCounter) Report(ctx context.Context) {
	for b, count := range c {
		suMetric.Set(ctx, int64(count), b.board, b.model, b.pool, b.environment, b.zone, b.status)
	}
}

// getMachineForLse returns the Machine that's attached to the MachineLSE
// iff the MachineLSE references exactly one Machine
func getMachineForLse(lse *ufspb.MachineLSE, idTomachineMap map[string]*ufspb.Machine) (*ufspb.Machine, error) {
	machines := lse.GetMachines()
	if n := len(machines); n != 1 {
		return nil, errors.Reason("report ufs inventory cron handler: %d machines %v associated with %q", n, machines, lse.GetName()).Err()
	}
	machine, ok := idTomachineMap[machines[0]]
	if !ok {
		return nil, errors.Reason("report ufs inventory cron handler: machine %s not found for LSE %s", machines[0], lse.GetName()).Err()
	}
	return machine, nil
}

// getBucketForDevice instantiates a *bucket for a given MachineLSE and
// corresponding Machine
func getBucketForDevice(lse *ufspb.MachineLSE, machine *ufspb.Machine, env string) *bucket {
	b := &bucket{
		board:       machine.GetChromeosMachine().GetBuildTarget(),
		model:       machine.GetChromeosMachine().GetModel(),
		pool:        "[None]",
		environment: env,
		zone:        lse.GetZone(),
		status:      dutstate.ConvertFromUFSState(lse.GetResourceState()).String(),
	}
	if dut := lse.GetChromeosMachineLse().GetDeviceLse().GetDut(); dut != nil {
		b.pool = getReportPool(dut.GetPools())
	}
	if labstation := lse.GetChromeosMachineLse().GetDeviceLse().GetLabstation(); labstation != nil {
		b.pool = getReportPool(labstation.GetPools())
	}
	return b
}

// machineFieldToValueFunc is a helper type for extracting the DUT fields
// for a given scheduling unit
type machineFieldToValueFunc func(machine *ufspb.Machine) string

var (
	machineBoardValueFunc = func(machine *ufspb.Machine) string { return machine.GetChromeosMachine().GetBuildTarget() }
	machineModelValueFunc = func(machine *ufspb.Machine) string { return machine.GetChromeosMachine().GetModel() }
)

// getBucketForSchedulingUnit instantiates a *bucket for a given SchedulingUnit
// and corresponding DUTs.
// Depending on the ExposeType, the bucket dimensions are based on a combination
// of the primary DUT values and an aggregate on all DUTs
func getBucketForSchedulingUnit(su *ufspb.SchedulingUnit, lses []*ufspb.MachineLSE, idTomachineMap map[string]*ufspb.Machine, env string) (*bucket, error) {
	b := &bucket{
		board:       "[None]",
		model:       "[None]",
		pool:        getReportPool(su.GetPools()),
		environment: env,
		zone:        "[None]",
		status:      schedulingUnitStatusFromLses(lses),
	}
	// fields from all DUTs
	switch su.GetExposeType() {
	case ufspb.SchedulingUnit_DEFAULT:
		fallthrough
	case ufspb.SchedulingUnit_DEFAULT_PLUS_PRIMARY:
		board, err := schedulingUnitLabelForLses(lses, idTomachineMap, machineBoardValueFunc)
		if err != nil {
			return nil, err
		}
		model, err := schedulingUnitLabelForLses(lses, idTomachineMap, machineModelValueFunc)
		if err != nil {
			return nil, err
		}
		b.board = board
		b.model = model
	case ufspb.SchedulingUnit_STRICTLY_PRIMARY_ONLY:
		// nothing from all DUTs
	default:
		return nil, errors.Reason("Unknown SchedulingUnit Expose Type for %s", su.GetName()).Err()
	}
	// fields from primary
	var primaryLse *ufspb.MachineLSE
	for _, lse := range lses {
		if lse.GetName() == su.GetPrimaryDut() {
			primaryLse = lse
			break
		}
	}
	switch su.GetExposeType() {
	case ufspb.SchedulingUnit_DEFAULT:
		// nothing from the primary DUT
	case ufspb.SchedulingUnit_DEFAULT_PLUS_PRIMARY:
		if primaryLse == nil {
			return nil, errors.Reason("Could not find primary MachineLSE %s for scheduling unit %s", su.GetPrimaryDut(), su.GetName()).Err()
		}
		b.zone = primaryLse.GetZone()
	case ufspb.SchedulingUnit_STRICTLY_PRIMARY_ONLY:
		if primaryLse == nil {
			return nil, errors.Reason("Could not find primary MachineLSE %s for scheduling unit %s", su.GetPrimaryDut(), su.GetName()).Err()
		}
		machine, err := getMachineForLse(primaryLse, idTomachineMap)
		if err != nil {
			return nil, err
		}
		b.board = machine.GetChromeosMachine().GetBuildTarget()
		b.model = machine.GetChromeosMachine().GetModel()
		b.zone = primaryLse.GetZone()
	default:
		return nil, errors.Reason("Unknown SchedulingUnit Expose Type for %s", su.GetName()).Err()
	}
	return b, nil
}

// schedulingUnitLabelForLses calculates an overall label for a scheduling unit
// given a list of MachineLSEs
func schedulingUnitLabelForLses(lses []*ufspb.MachineLSE, idTomachineMap map[string]*ufspb.Machine, f machineFieldToValueFunc) (string, error) {
	machines := make([]*ufspb.Machine, len(lses))
	for i, lse := range lses {
		machine, err := getMachineForLse(lse, idTomachineMap)
		if err != nil {
			return "", err
		}
		machines[i] = machine
	}
	labelSet := make(map[string]struct{}) // Set of all label values
	for _, machine := range machines {
		machineLabel := f(machine)
		if len(machineLabel) > 0 {
			labelSet[machineLabel] = struct{}{}
		}
	}
	labels := make([]string, 0, len(labelSet))
	for k := range labelSet {
		labels = append(labels, k)
	}
	return summarizeValues(labels), nil
}

// schedulingUnitStatusFromLses calculates a weighted status based on all DUTs
// to represent the scheduling unit
func schedulingUnitStatusFromLses(lses []*ufspb.MachineLSE) string {
	states := make([]string, len(lses))
	for i, lse := range lses {
		s := dutstate.ConvertFromUFSState(lse.GetResourceState()).String()
		states[i] = s
	}
	return util.SchedulingUnitDutState(states)
}

// bucket contains static DUT dimensions.
//
// These dimensions do not change often. If all DUTs with a given set of
// dimensions are removed, the related metric is not automatically reset. The
// metric will get reset eventually.
type bucket struct {
	board       string
	model       string
	pool        string
	environment string
	zone        string
	status      string
}

func (b *bucket) String() string {
	return fmt.Sprintf("board: %s, model: %s, pool: %s, env: %s, zone: %q", b.board, b.model, b.pool, b.environment, b.zone)
}

func summarizeValues(vs []string) string {
	switch len(vs) {
	case 0:
		return "[None]"
	case 1:
		return vs[0]
	default:
		return "[Multiple]"
	}
}

func isManagedPool(p string) bool {
	_, ok := invV1.SchedulableLabels_DUTPool_value[p]
	return ok
}

func getReportPool(pools []string) string {
	p := summarizeValues(pools)
	if isManagedPool(p) {
		return fmt.Sprintf("managed:%s", p)
	}
	return p
}

func getAllMachineLSEs(ctx context.Context, keysOnly bool) ([]*ufspb.MachineLSE, error) {
	var lses []*ufspb.MachineLSE
	for startToken := ""; ; {
		res, nextToken, err := inventory.ListMachineLSEs(ctx, pageSize, startToken, nil, keysOnly)
		if err != nil {
			return nil, errors.Annotate(err, "get all MachineLSEs").Err()
		}
		lses = append(lses, res...)
		if nextToken == "" {
			break
		}
		startToken = nextToken
	}
	return lses, nil
}

func getAllMachines(ctx context.Context, keysOnly bool) ([]*ufspb.Machine, error) {
	var machines []*ufspb.Machine
	for startToken := ""; ; {
		res, nextToken, err := registration.ListMachines(ctx, pageSize, startToken, nil, keysOnly)
		if err != nil {
			return nil, errors.Annotate(err, "get all Machines").Err()
		}
		machines = append(machines, res...)
		if nextToken == "" {
			break
		}
		startToken = nextToken
	}
	return machines, nil
}
