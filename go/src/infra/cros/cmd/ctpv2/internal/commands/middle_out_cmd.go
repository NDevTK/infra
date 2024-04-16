// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	hashstructure "github.com/mitchellh/hashstructure/v2"
	"google.golang.org/protobuf/proto"

	"go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/analytics"
	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/ctpv2/data"
)

// FilterExecutionCmd represents test execution cmd.
type MiddleOutRequestCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	InternalTestPlan *api.InternalTestplan

	// Updates
	MiddledOutResp *data.MiddleOutResponse

	// BQ client for logging
	BQClient *bigquery.Client
	// BuildState
	BuildState *build.State
}

// ExtractDependencies (Boiler plate)
func (cmd *MiddleOutRequestCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.FilterStateKeeper:
		err = cmd.extractDepsFromFilterStateKeeper(ctx, sk)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// UpdateStateKeeper (Boiler plate)
func (cmd *MiddleOutRequestCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.FilterStateKeeper:
		err = cmd.updateFilterStateKeeper(ctx, sk)
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

func (cmd *MiddleOutRequestCmd) extractDepsFromFilterStateKeeper(
	ctx context.Context,
	sk *data.FilterStateKeeper) error {

	// If no plan...
	if sk.TestPlanStates == nil || len(sk.TestPlanStates) == 0 {
		if sk.InitialInternalTestPlan != nil {
			// Set the first state from initial test plan
			sk.TestPlanStates = append(sk.TestPlanStates, sk.InitialInternalTestPlan)
			// Set the cmd input test plan
			cmd.InternalTestPlan = proto.Clone(sk.InitialInternalTestPlan).(*api.InternalTestplan)
		} else {
			return fmt.Errorf("Cmd %q missing dependency: InputTestPlan", cmd.GetCommandType())
		}
	} else {
		// Get the last test plan state and set it as input test plan for current filter
		cmd.InternalTestPlan = proto.Clone(sk.TestPlanStates[len(sk.TestPlanStates)-1]).(*api.InternalTestplan)
	}

	if sk.BQClient != nil {
		cmd.BQClient = sk.BQClient
	}
	cmd.BuildState = sk.BuildState
	return nil
}

func (cmd *MiddleOutRequestCmd) updateFilterStateKeeper(
	ctx context.Context,
	sk *data.FilterStateKeeper) error {

	if cmd.MiddledOutResp != nil {
		sk.MiddledOutResp = cmd.MiddledOutResp
	}

	return nil
}

// Execute executes the command.
func (cmd *MiddleOutRequestCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Middle Out")
	defer func() { step.End(err) }()

	key := "middleout-execute"

	if cmd.BQClient != nil {
		analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, &analytics.BqData{Step: key, Status: analytics.Start}, cmd.InternalTestPlan, cmd.BuildState)
	}
	start := time.Now()
	status := analytics.Success

	// TODO (dbeckett/Aziz) figure out how to properly make this
	pool := cmd.InternalTestPlan.GetSuiteInfo().GetSuiteMetadata().GetPool()
	if pool == "" {
		pool = "DUT_POOL_QUOTA"
	}
	cfg := distroCfg{maxInShard: 150, pool: pool}

	trReqs, err := middleOut(ctx, cmd.InternalTestPlan, cfg)
	if err != nil {
		status = analytics.Fail
		if cmd.BQClient != nil {
			analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, &analytics.BqData{Step: key, Status: status, Duration: float32(time.Now().Sub(start).Seconds())}, cmd.InternalTestPlan, cmd.BuildState)
		}

		return errors.Annotate(err, "Failed to execute MiddleOPut: ").Err()
	}
	if cmd.BQClient != nil {
		analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, &analytics.BqData{Step: key, Status: status, Duration: float32(time.Now().Sub(start).Seconds())}, cmd.InternalTestPlan, cmd.BuildState)
	}
	logging.Infof(
		ctx,
		"len of data: %d",
		len(trReqs))

	cmd.MiddledOutResp = &data.MiddleOutResponse{TrReqs: trReqs, SuiteInfo: cmd.InternalTestPlan.GetSuiteInfo()}

	middleOutData, err := json.MarshalIndent(cmd.MiddledOutResp, "", "  ")
	if err != nil {
		logging.Infof(
			ctx,
			"error during writing MO response to log: %s",
			err.Error())
	}
	_, _ = step.Log("Middle out output").Write(middleOutData)

	return nil
}

func NewMiddleOutRequestCmd() *MiddleOutRequestCmd {
	abstractCmd := interfaces.NewAbstractCmd(MiddleoutExecutionType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &MiddleOutRequestCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}

// loading is used in lab avalability such that devices with the same HW;
// but different SW requirements, still share the same pool of physical devices.
type loading struct {
	value int
}

type hwInfo struct {
	req *api.SchedulingUnitOptions

	// TODO; when HwRequirements is fully deprecated, remove this.
	oldReq *api.HWRequirements

	labLoading         *loading
	numInCurrentShard  int
	hwValue            uint64
	matchingValue      uint64
	provValue          uint64
	labDevices         int64
	shardHarness       string
	dimsExcludingReady []string
}

// Kv structs are useful for gobased sorting.
type kv struct {
	key   uint64
	value int
}

type distroCfg struct {
	pool            string
	isUnitTest      bool
	unitTestDevices int
	maxInShard      int
}

type middleOutData struct {
	// Originally defined HW to TC
	hwToTCMap map[uint64][]string

	// Map between original HW objects, and the flat HW equivs
	// Example: [sha123]: [sha1, sha2, sha3]
	hwEquivalenceMap map[uint64][]uint64

	// the Sha of the HWRequirements object from a TC, pointing back to the HWRequirements object
	hwUUIDMap map[uint64]*api.SchedulingUnitOptions

	// TODO; when HwRequirements is fully deprecated, remove this.
	oldhwUUIDMap map[uint64]*api.HWRequirements

	cfg distroCfg

	// Flat HWUUID to flat HW
	flatHWUUIDMap map[uint64]*hwInfo

	// The TestCase name pointing to its respective object
	tcUUIDMap map[string]*api.CTPTestCase

	finalAssignments map[uint64][][]string
}

// newMiddleOutData returns a struct of the middleOutData with the data init'd but empty.
func newMiddleOutData() *middleOutData {
	mo := &middleOutData{
		hwToTCMap:        make(map[uint64][]string),
		hwEquivalenceMap: make(map[uint64][]uint64),
		hwUUIDMap:        make(map[uint64]*api.SchedulingUnitOptions),

		// TODO; when HwRequirements is fully deprecated, remove this.
		oldhwUUIDMap: make(map[uint64]*api.HWRequirements),

		cfg:           distroCfg{},
		flatHWUUIDMap: make(map[uint64]*hwInfo),
		tcUUIDMap:     make(map[string]*api.CTPTestCase),

		finalAssignments: make(map[uint64][][]string),
	}
	return mo

}

func getName(tc *api.CTPTestCase) string {
	return tc.GetMetadata().GetTestCase().GetId().GetValue()
}

func oldProto(targs []*api.HWRequirements) bool {
	return len(targs) > 0
}

// middleOut creates TRRequest(S) from a ctpv2 internal test plan.
func middleOut(ctx context.Context, resp *api.InternalTestplan, cfg distroCfg) ([]*data.TrRequest, error) {
	solverData := newMiddleOutData()
	solverData.cfg = cfg
	for _, tc := range resp.GetTestCases() {
		tcUUID := getName(tc)
		// Drop all of the HW fluff in the TC for memory sakes.
		tcForMap := &api.CTPTestCase{
			Name:     tc.GetName(),
			Metadata: tc.GetMetadata(),
		}
		solverData.tcUUIDMap[tcUUID] = tcForMap

		// TODO; when HwRequirements is fully deprecated, remove this.
		if oldProto(tc.HwRequirements) {
			for _, hw := range tc.HwRequirements {
				// Note: Each `hw` is still a repeated list of HW *options* for the test.
				hash := oldAddHWtohwUUIDMap(solverData.oldhwUUIDMap, hw)
				for k, v := range oldFlattenList(ctx, []*api.HWRequirements{hw}) {
					err := addHWtoFlatHWUUIDMap(ctx, solverData.flatHWUUIDMap, k, v)
					if err != nil {
						logging.Infof(ctx, fmt.Sprintf("error found in addHWtoFlatHWUUIDMap: %s", err))
						return nil, err
					}
				}
				solverData.hwToTCMap[hash] = append(solverData.hwToTCMap[hash], tcUUID)

			}
		} else
		// tc.HwRequirements example:
		// [[hw1] && [hw2] && [hw3]]
		// OR [[hw1 || hw2] && [hw3 || hw4]]
		{
			for _, hw := range tc.SchedulingUnitOptions {
				// Note: Each `hw` is still a repeated list of HW *options* for the test.
				hash := addHWtohwUUIDMap(solverData.hwUUIDMap, hw)
				for k, v := range flattenList(ctx, []*api.SchedulingUnitOptions{hw}) {
					err := addHWtoFlatHWUUIDMap(ctx, solverData.flatHWUUIDMap, k, v)
					if err != nil {
						logging.Infof(ctx, fmt.Sprintf("error found in addHWtoFlatHWUUIDMap: %s", err))
						return nil, err
					}
				}
				solverData.hwToTCMap[hash] = append(solverData.hwToTCMap[hash], tcUUID)

			}
		}
	}

	// TODO; when HwRequirements is fully deprecated, remove this.
	if len(solverData.oldhwUUIDMap) > 0 {
		for key, hwOptions := range solverData.oldhwUUIDMap {
			solverData.hwEquivalenceMap[key] = oldFindMatches(hwOptions, solverData.flatHWUUIDMap)
		}
	} else {
		// Goal here is to make the hwEquivalenceMap; such that when we go to distribute the hwToTCMap
		// We can lookup what HW options are available.
		for key, hwOptions := range solverData.hwUUIDMap {
			solverData.hwEquivalenceMap[key] = findMatches(ctx, hwOptions, solverData.flatHWUUIDMap)
		}
	}

	return createTrRequests(greedyDistro(ctx, solverData), solverData)
}

// createTrRequests translates a final {hw:[[shard], [shard]]} map into a flat list of TrRequests.
func createTrRequests(distro map[uint64][][]string, solverData *middleOutData) ([]*data.TrRequest, error) {
	TrRequests := []*data.TrRequest{}
	for k, shards := range distro {
		for _, tcs := range shards {
			shardedtcs := []*api.CTPTestCase{}
			for _, tc := range tcs {
				lltc, ok := solverData.tcUUIDMap[tc]
				if !ok {
					return TrRequests, fmt.Errorf("tc assigned but was not given, something critically wrong happened to end up here")
				}
				shardedtcs = append(shardedtcs, lltc)
			}

			// TODO; when HwRequirements is fully deprecated, remove `Req`.

			trReq := &data.TrRequest{
				Req:    solverData.flatHWUUIDMap[k].oldReq,
				NewReq: solverData.flatHWUUIDMap[k].req,
				Tcs:    shardedtcs,
				DevicesInfo: &data.DevicesInfo{LabDevicesCount: solverData.flatHWUUIDMap[k].labDevices,
					Dims: solverData.flatHWUUIDMap[k].dimsExcludingReady,
				},
			}
			TrRequests = append(TrRequests, trReq)
		}
	}
	return TrRequests, nil
}

// Given the class map, and the tc map, gather the lab loading for the devices, and make our best guess at assigning tests to HW.
func greedyDistro(ctx context.Context, solverData *middleOutData) map[uint64][][]string {

	populateLabAvalability(ctx, solverData)
	orderedHw := hwSearchOrdering(solverData.hwEquivalenceMap)

	// Starting with the HW classes with the least amount of  equivalent classes
	for _, hwS := range orderedHw {
		// Find the TCS that require this class, and assign them devices.
		hwHash := hwS.key
		tcs := solverData.hwToTCMap[hwHash]

		// Currently we will not try anymore than basic sharding.
		// As in, we won't attempt to "fill" a pod, then spill over.
		// Its either "you can take all these tests" or we get a new pod.
		shards := shard(tcs, solverData.cfg.maxInShard)
		for _, shardedtc := range shards {

			harness := ""
			if len(shardedtc) > 0 {
				harness = getHarness(shardedtc[0])
			}
			selectedDevice, expandCurrentShard := getDevices(solverData, len(shardedtc), hwHash, harness)
			assignHardware(solverData, selectedDevice, expandCurrentShard, shardedtc)

		}
	}

	return solverData.finalAssignments
}

// assignHardware will add the tests to the selectedDevice, being aware if it should go into a non-filled hard, or a new one.
// assignHardware will also decrement the number of devices remaining every time device is assigned tests.
func assignHardware(solverData *middleOutData, selectedDevice uint64, expandCurrentShard bool, shardedtc []string) {
	if expandCurrentShard {
		lastElement := len(solverData.finalAssignments[selectedDevice])
		solverData.finalAssignments[selectedDevice][lastElement-1] = append(solverData.finalAssignments[selectedDevice][lastElement-1], shardedtc...)
		solverData.flatHWUUIDMap[selectedDevice].numInCurrentShard += len(shardedtc) // Show the status of the current shard. Might be wrong.
		if solverData.flatHWUUIDMap[selectedDevice].numInCurrentShard == solverData.cfg.maxInShard {
			solverData.flatHWUUIDMap[selectedDevice].numInCurrentShard = 0
		}
	} else {
		_, ok := solverData.finalAssignments[selectedDevice]
		if !ok {
			solverData.finalAssignments[selectedDevice] = [][]string{}
		}

		solverData.finalAssignments[selectedDevice] = append(solverData.finalAssignments[selectedDevice], shardedtc)

		solverData.flatHWUUIDMap[selectedDevice].labLoading.value-- // Reduce the # of open devices by 1.
		// If the shard is not full, mark it as such.
		if len(shardedtc) != solverData.cfg.maxInShard {
			solverData.flatHWUUIDMap[selectedDevice].numInCurrentShard += len(shardedtc)
			solverData.flatHWUUIDMap[selectedDevice].shardHarness = getHarness(shardedtc[0])
		}
	}
}

type helper struct {
	hwOption *api.SwarmingDefinition
	hashV    uint64
	// TODO remove `hashProvV` once hwRequirements has been fully removed.
	hashProvV      uint64
	swarmingLabels []string
}

// TODO remove this method once hwRequirements has been fully removed.
// oldFindMatches: find options for the given hwOption from the given flatHWUUIDMap. LEGACY PROTO
func oldFindMatches(hwOption *api.HWRequirements, flatHWUUIDMap map[uint64]*hwInfo) []uint64 {
	matches := []uint64{}
	cnt := 0

	helperList := []*helper{}
	for _, child := range hwOption.GetHwDefinition() {
		childHash, _ := hashstructure.Hash(child.DutInfo, hashstructure.FormatV2, nil)
		childHashProv, _ := hashstructure.Hash(child.Variant, hashstructure.FormatV2, nil)
		h := &helper{
			hashV:          childHash,
			hashProvV:      childHashProv,
			swarmingLabels: child.GetSwarmingLabels(),
		}
		helperList = append(helperList, h)
	}

	for sha, foundHW := range flatHWUUIDMap {
		if len(foundHW.oldReq.GetHwDefinition()) > 1 {
			panic("foundHW.req.HwDefinition should not have more than 1 object inside it")
		}

		if oldIsParent(foundHW, helperList) {
			matches = append(matches, sha)
		}
		cnt++

	}
	return matches
}

// TODO remove this method once hwRequirements has been fully removed.
// oldIsParent determines if the first is fully encompassing of the second object. LEGACY PROTO
func oldIsParent(parentSrc *hwInfo, childSrc []*helper) bool {
	/*
		In other words, if the parent has at least all of the fields of the child, it can run the child.
	*/
	// [1]

	// [1... 5000]
	if len(parentSrc.oldReq.GetHwDefinition()) != 1 && len(childSrc) > 0 {
		return false
	}

	// isParent is true by default if there is no child
	correct := true
	for _, child := range childSrc {
		childHash := child.hashV
		childHashProv := child.hashProvV
		// TODO (dbeckett/azrahman): determine if the AllItemsIn check is going to be problematic and needs to be cached.
		if childHash == parentSrc.hwValue && childHashProv == parentSrc.provValue && allItemsIn(child.swarmingLabels, parentSrc.oldReq.GetHwDefinition()[0].GetSwarmingLabels()) {
			return true
		}

		correct = false
	}

	return correct
}

// findMatches: find options for the given hwOption from the given allFlatHWUUIDMap
func findMatches(ctx context.Context, oneOfHws *api.SchedulingUnitOptions, allFlatHWUUIDMap map[uint64]*hwInfo) []uint64 {
	matches := []uint64{}
	cnt := 0

	/* Using this example:
	 {tc.SchedulingUnits = [
		SchedulingUnitOptions = {
			Schedulingunits = [
				SchedulingUnit = {
					Primary = SwarmingDef1
				},
				SchedulingUnit = {
					Primary = SwarmingDef12
				},
			],
			state: ONEOF
		},
		SchedulingUnitOptions = {
			Schedulingunits = [
				SchedulingUnit = {
					Primary = SwarmingDef3
				},
				SchedulingUnit = {
					Primary = SwarmingDef2
				},
			],
			state: ONEOF
		},

	]}


	`hwOption` is going to only _one_ of the `SchedulingUnitOptions` (not the entire []SchedulingUnits)

		SchedulingUnitOptions = {
			Schedulingunits = [
				SchedulingUnit = {
					Primary = SwarmingDef1
				},
				SchedulingUnit = {
					Primary = SwarmingDef12
				},
			],
			state: ONEOF
		}

	`flatHWUUIDMap` is going to be a full list of all flat HW. Eg:

	[
				SchedulingUnit = {
					Primary = SwarmingDef1
				},
				SchedulingUnit = {
					Primary = SwarmingDef2
				},
				SchedulingUnit = {
					Primary = SwarmingDef3
				},
	]


	The goal of the method is to see per option in `SchedulingUnitOptions`; which DUTS in the flat map match it.
	This is useful for bundling tests which have semi-overlapping deps.
	*/
	EqClassList := []*helper{}
	for _, child := range oneOfHws.SchedulingUnits {

		childHash := hashForSchedulingUnit(child)

		h := &helper{
			hashV:          childHash,
			swarmingLabels: child.GetPrimaryTarget().GetSwarmingDef().GetSwarmingLabels(),
		}

		EqClassList = append(EqClassList, h)
	}

	// From the existing HWUIID map, see what DUTs can run what classes.
	// This effectively means that selecting:
	// `HW.1234` is valid for ['HW.1234' || 'HW.1222']
	// `HW.1222` is valid for ['HW.1234' || 'HW.1222']
	for sha, foundHW := range allFlatHWUUIDMap {
		if len(foundHW.req.SchedulingUnits) > 1 {
			panic("foundHW.req.SchedulingUnits should not have more than 1 object inside it")
		}
		ip := isParentofAtleastOne(foundHW, EqClassList)
		if ip {
			matches = append(matches, sha)
		}
		cnt++

	}
	return matches
}

// isParentofOne determines if the first is fully encompassing of the second object.
func isParentofAtleastOne(parentSrc *hwInfo, childSrc []*helper) bool {
	/*
		In other words, if the parent has at least all of the fields of the child, it can run the child.
	*/
	// [1]

	// [1... 5000]

	// isParent is true by default if there is no child
	correct := true
	for _, child := range childSrc {
		childHash := child.hashV

		// TODO (dbeckett/azrahman): determine if the AllItemsIn check is going to be problematic and needs to be cached.
		if len(parentSrc.req.GetSchedulingUnits()) < 1 {
			return false
		}
		if childHash == parentSrc.hwValue && allItemsIn(child.swarmingLabels, parentSrc.req.GetSchedulingUnits()[0].GetPrimaryTarget().GetSwarmingDef().GetSwarmingLabels()) {
			return true
		}
		correct = false
	}

	return correct
}

// Check all items from 1 are in 2
func allItemsIn(item1 []string, item2 []string) bool {
	// TODO (dbeckett/azrahman): squish this such that we are doing set comparisons, rather than list.
	for _, string1 := range item1 {
		found := false
		for _, string2 := range item2 {
			if string1 == string2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true

}

// shard will device the list into a list of lists where each item in the list length of maxInShard
// eg: [1,2,3,4], maxInShard=2 --> [[1,2], [3,4]]
func shard(alltests []string, maxInShard int) (shards [][]string) {
	harnessBuckets := make(map[string][]string)
	for _, test := range alltests {
		h := getHarness(test)
		_, ok := harnessBuckets[h]

		if !ok {
			harnessBuckets[h] = []string{test}
		} else {
			harnessBuckets[h] = append(harnessBuckets[h], test)
		}
	}

	for _, tests := range harnessBuckets {

		for maxInShard < len(tests) {
			tests, shards = tests[maxInShard:], append(shards, tests[0:maxInShard:maxInShard])
		}
		shards = append(shards, tests)
	}

	return shards
}

// extracts the harness out of the test name.
func getHarness(t string) string {
	v := strings.Split(t, ".")
	// if there is less than 2 parts to the name, then its not a known test format.
	// so just return "unknown"
	if len(v) < 2 {
		return "unknown"
	}
	return v[0]
}

// TODO (azrahman@; integrate with swarming API)
func labAvalability(*api.HWRequirements) int {
	// returns the number of devices which can meet the requirement.
	// will need to include pool
	return -1
}

// Will add the amount of devices in the lab to each of the HW items.
func populateLabAvalability(ctx context.Context, solverData *middleOutData) {
	// toComplete will track what hwInfo will need labAvailibility info to be populated.
	toComplete := []*hwInfo{}
	hwFound := make(map[uint64]*loading)
	for _, value := range solverData.flatHWUUIDMap {
		hash := uint64(0)

		// TODO remove this if block (keep the else) once hwRequirements has been fully removed.
		if len(value.oldReq.GetHwDefinition()) > 0 {
			hash, _ = hashstructure.Hash(value.oldReq.GetHwDefinition()[0].GetDutInfo(), hashstructure.FormatV2, nil)
		} else {

			// 0 index is safe as this is coming from the flatHWUUIDMap; which garuntees the list to be flat.
			// We have to use the `HwOnly` hashfor; as we should consider a variant as the same underlying hardware.
			// Eg; Brya and brya-kernelnext, are the same underlying DUT.
			hash = hashForSchedulingUnitHwOnly(value.req.GetSchedulingUnits()[0])

		}

		_, exists := hwFound[hash]
		if exists {
			value.labLoading = hwFound[hash]
			continue
		}

		if solverData.cfg.unitTestDevices != 0 {
			hwFound[hash] = &loading{value: solverData.cfg.unitTestDevices}
		} else {
			value.labLoading = &loading{value: 5}
		}
		toComplete = append(toComplete, value)
		value.labLoading = hwFound[hash]
	}

	if solverData.cfg.unitTestDevices == 0 {
		logging.Infof(ctx, "Looking for lab devices")
		swarmingServ, err := common.CreateNewSwarmingService(context.Background())
		if err != nil {
			logging.Infof(ctx, fmt.Sprintf("error found while creating new swarming service: %s", err))
			return
		}
		// Query swarming asynchronously for device availability.
		wg := sync.WaitGroup{}
		for _, hwInfoObj := range toComplete {
			wg.Add(1)
			go func(hwInfoInput *hwInfo) {
				defer wg.Done()
				dims := CreateDims(ctx, hwInfoInput, solverData.cfg.pool, true)

				dimsExcludingReady := CreateDims(ctx, hwInfoInput, solverData.cfg.pool, false)

				botCount, err := common.GetBotCount(ctx, dims, swarmingServ)
				if err != nil {
					logging.Infof(ctx, fmt.Sprintf("error found in GetBOTcount: %s", err))
				}
				totalBotCount, err := common.GetBotCount(ctx, dimsExcludingReady, swarmingServ)
				if err != nil {
					logging.Infof(ctx, fmt.Sprintf("error found in GetBOTcount: %s", err))
				}

				hwInfoInput.labLoading = &loading{value: int(botCount)}
				hwInfoInput.labDevices = totalBotCount
				hwInfoInput.dimsExcludingReady = dimsExcludingReady
				logging.Infof(ctx, "Found for lab devices: %v", botCount)

			}(hwInfoObj)
		}
		wg.Wait()
	}
}

func dutModelFromSwarmingDef(def *api.SwarmingDefinition) *labapi.DutModel {
	switch hw := def.GetDutInfo().GetDutType().(type) {
	case *labapi.Dut_Chromeos:
		return hw.Chromeos.GetDutModel()
	case *labapi.Dut_Android_:
		return hw.Android.GetDutModel()
	case *labapi.Dut_Devboard_:
		return hw.Devboard.GetDutModel()
	}
	return nil
}

func addBoardModelToDims(unit *api.SchedulingUnit, dims []string) []string {
	primaryBoard := dutModelFromSwarmingDef(unit.GetPrimaryTarget().GetSwarmingDef()).GetBuildTarget()
	dims = append(dims, fmt.Sprintf("label-board:%s", primaryBoard))

	primaryModel := dutModelFromSwarmingDef(unit.GetPrimaryTarget().GetSwarmingDef()).GetModelName()
	if primaryModel != "" {
		dims = append(dims, fmt.Sprintf("label-model:%s", primaryModel))

	}
	for _, secondary := range unit.GetCompanionTargets() {
		board := dutModelFromSwarmingDef(secondary.GetSwarmingDef()).GetBuildTarget()
		// When equal, the secondary needs the _2.
		if board == primaryBoard {
			board = fmt.Sprintf("%s_2", board)
		}
		dims = append(dims, fmt.Sprintf("label-board:%s", board))

		model := dutModelFromSwarmingDef(secondary.GetSwarmingDef()).GetModelName()
		if model == primaryModel && model != "" {
			model = fmt.Sprintf("%s_2", model)
		}

		dims = append(dims, fmt.Sprintf("label-model:%s", model))

	}

	return dims
}

// CreateDims creates dims list from hwInfo object.
func CreateDims(ctx context.Context, hwInfo *hwInfo, pool string, readycheck bool) []string {
	// TODO replace `units` with just `hwInfo.req.GetSchedulingUnits()` once hwRequirements has been fully removed.
	units := 0
	if hwInfo.oldReq != nil {
		units = len(hwInfo.oldReq.GetHwDefinition())
	} else if hwInfo.req == nil {
		units = len(hwInfo.req.GetSchedulingUnits())
	}

	if hwInfo == nil || (hwInfo.req == nil && hwInfo.oldReq == nil) || units < 1 {
		return []string{}
	}

	dims := []string{}
	if readycheck {
		dims = append(dims, "dut_state:ready")
	}

	if len(hwInfo.req.GetSchedulingUnits()) > 0 {
		dims = ConvertSwarmingLabelsToDims(dims, hwInfo.req.GetSchedulingUnits()[0].GetPrimaryTarget().GetSwarmingDef().GetSwarmingLabels())
		dims = addBoardModelToDims(hwInfo.req.GetSchedulingUnits()[0], dims)
	} else if len(hwInfo.oldReq.GetHwDefinition()) == 1 {
		// TODO remove this entire `else` statement when HWRequirements is done.
		dims = ConvertSwarmingLabelsToDims(dims, hwInfo.oldReq.GetHwDefinition()[0].GetSwarmingLabels())
		if hwInfo.oldReq.GetHwDefinition()[0] != nil {
			dutInfo := hwInfo.oldReq.GetHwDefinition()[0].GetDutInfo()
			if dutInfo.GetChromeos().GetDutModel() != nil {
				dutModel := dutInfo.GetChromeos().GetDutModel()
				if dutModel.GetBuildTarget() != "" {
					dims = append(dims, fmt.Sprintf("label-board:%s", strings.ToLower(dutModel.GetBuildTarget())))
				}
				if dutModel.GetModelName() != "" {
					dims = append(dims, fmt.Sprintf("label-model:%s", strings.ToLower(dutModel.GetModelName())))
				}
			}
		}
	} else {
		return []string{}
	}
	if hwidFromHwInfo(hwInfo) != "" {
		dims = append(dims, fmt.Sprintf("hwid:%s", hwidFromHwInfo(hwInfo)))
	}
	if pool != "" {
		dims = append(dims, fmt.Sprintf("label-pool:%s", pool))
	}

	return dims
}

func hwidFromHwInfo(hwInfo *hwInfo) string {
	if hwInfo.oldReq != nil {
		return hwInfo.oldReq.GetHwDefinition()[0].GetDutInfo().GetChromeos().GetHwid()
	} else if hwInfo.req != nil {
		return hwInfo.req.GetSchedulingUnits()[0].GetPrimaryTarget().GetSwarmingDef().GetDutInfo().GetChromeos().GetHwid()
	}
	return ""
}

// ConvertSwarmingLabelsToDims converts provided swarming labels to swarming dims.
func ConvertSwarmingLabelsToDims(defaultDims []string, swarmingLabels []string) []string {
	dims := defaultDims
	for _, label := range swarmingLabels {
		if strings.Contains(label, "label-") {
			dims = append(dims, label)
		} else if strings.HasPrefix(label, "dut_name") {
			dims = append(dims, label)
		} else if strings.Contains(label, ":") {
			dims = append(dims, fmt.Sprintf("label-%s", label))
		} else {
			dims = append(dims, fmt.Sprintf("label-%s:True", label))
		}
	}
	return dims
}

// translate the given hwEquivalenceMap into a 2d map:
// EG [id1=[1 || 2] || id2=[3 || 4]] needs to be [[1, 2, 3, 4]]
// hwEquivalenceMap is like id1 = [id1, id2]
// now needs to be like id1 = [newid1, newid2, newid3, newid4]
func flattenEqMap(hwEquivalenceMap map[uint64][]uint64, hwUUIDMap map[uint64]*api.SchedulingUnitOptions) (map[uint64][]uint64, map[uint64]*hwInfo) {
	newHWUUIDMap := make(map[uint64]*hwInfo)
	newHwEquivalenceMap := make(map[uint64][]uint64)

	for hw, results := range hwEquivalenceMap {
		var allHw []*api.SchedulingUnitOptions
		newHwEquivalenceMap[hw] = []uint64{}

		allHw = append(allHw, hwUUIDMap[hw])
		for _, hash := range results {
			allHw = append(allHw, hwUUIDMap[hash])
		}

		// flattened is now the uuid for [newid1: obj, etc]
		flattened := flattenList(context.Background(), allHw)

		for k, v := range flattened {
			newHWUUIDMap[k] = v
			newHwEquivalenceMap[hw] = append(newHwEquivalenceMap[hw], k)
		}
	}
	return newHwEquivalenceMap, newHWUUIDMap
}

// TODO; when HwRequirements is fully deprecated, remove this.
func oldFlattenEqMap(hwEquivalenceMap map[uint64][]uint64, hwUUIDMap map[uint64]*api.HWRequirements) (map[uint64][]uint64, map[uint64]*hwInfo) {
	newHWUUIDMap := make(map[uint64]*hwInfo)
	newHwEquivalenceMap := make(map[uint64][]uint64)

	for hw, results := range hwEquivalenceMap {
		allHw := []*api.HWRequirements{hwUUIDMap[hw]}

		for _, hash := range results {
			allHw = append(allHw, hwUUIDMap[hash])
		}

		// flattened is now the uuid for [newid1: obj, etc]
		flattened := oldFlattenList(context.Background(), allHw)

		for k, v := range flattened {
			newHWUUIDMap[k] = v
			newHwEquivalenceMap[hw] = append(newHwEquivalenceMap[hw], k)
		}
	}
	return newHwEquivalenceMap, newHWUUIDMap
}

// TODO; when HwRequirements is fully deprecated, remove this.
// oldFlattenList translates [hw.requirements=[1||2], hw.requirements=[3||4]] into [[1],[2],[3],[4]]
func oldFlattenList(ctx context.Context, allHw []*api.HWRequirements) map[uint64]*hwInfo {
	flatHW := make(map[uint64]*hwInfo)

	for _, hw := range allHw {

		for _, innerHW := range hw.GetHwDefinition() {

			flattened := &api.HWRequirements{
				HwDefinition: []*api.SwarmingDefinition{innerHW},
			}
			dutInfoHash, err := hashstructure.Hash(innerHW.DutInfo, hashstructure.FormatV2, nil)
			if err != nil {
				logging.Infof(ctx, fmt.Sprintf("error while creating hash for dutInfo: %s", err))
			}
			provInfoHash, err := hashstructure.Hash(innerHW.Variant, hashstructure.FormatV2, nil)
			if err != nil {
				logging.Infof(ctx, fmt.Sprintf("error while creating hash for provisionInfo: %s", err))
			}

			flattenedHash, err := hashstructure.Hash(flattened, hashstructure.FormatV2, nil)
			if err != nil {
				logging.Infof(ctx, fmt.Sprintf("error while creating hash for flattened: %s", err))
			}

			newHwInfo := &hwInfo{
				oldReq:    flattened,
				hwValue:   dutInfoHash,
				provValue: provInfoHash,
			}
			flatHW[flattenedHash] = newHwInfo
		}
	}

	return flatHW
}

// flattenList translates [hw.requirements=[1||2], hw.requirements=[3||4]] into [[1],[2],[3],[4]]
func flattenList(ctx context.Context, allHw []*api.SchedulingUnitOptions) map[uint64]*hwInfo {

	flatHW := make(map[uint64]*hwInfo)

	for _, hw := range allHw {

		for _, innerHW := range hw.SchedulingUnits {

			flattened := &api.SchedulingUnitOptions{
				SchedulingUnits: []*api.SchedulingUnit{innerHW},
			}

			flattenedHash, err := hashstructure.Hash(flattened, hashstructure.FormatV2, nil)
			if err != nil {
				logging.Infof(ctx, fmt.Sprintf("error while creating hash for flattened: %s", err))
			}

			newHwInfo := &hwInfo{
				req:           flattened,
				hwValue:       hashForSchedulingUnit(innerHW),
				matchingValue: hashForSchedulingUnit(innerHW),
			}

			flatHW[flattenedHash] = newHwInfo

		}

	}

	return flatHW
}

// getDevices finds a device from the devicepool + hwEquivalenceMap to satsify the need for the test
// It will first look for a matching device with a non-full shard that fits,
// otherwise it will look for a device with the most availability in the lab.
func getDevices(solverData *middleOutData, numTests int, hwHash uint64, harness string) (selectedDevice uint64, append bool) {
	// This is a pretty expensive approach to sharding:
	// We will always check all devices to see if they have room in a non-empty shard.
	// So even when we fully fill a device, or it hasn't been touched, we still check it.
	// First, try to fill a non-empty shard
	devices := solverData.hwEquivalenceMap[hwHash]

	for _, device := range devices {
		// if the shard is empty, we need to use the labloading process block
		// not the shard filler.
		if solverData.flatHWUUIDMap[device].numInCurrentShard < 1 {
			continue
		}

		if solverData.flatHWUUIDMap[device].shardHarness != harness {
			continue
		}

		// Only assign it into a shard if there is actually devices.
		// There are cases where a test requires a device which doesn't exist (to later be rejected)
		// But in these examples, its viewed as an "open shard", so we toss other tests with overlapping eq classes
		// into the shard; resulting in those tests being skipped.
		// Instead, when we put the `0` check, we will not put the test in the shard; and grab a different (existing) device.
		if (solverData.flatHWUUIDMap[device].numInCurrentShard+numTests <= solverData.cfg.maxInShard) && solverData.flatHWUUIDMap[device].labLoading.value > 0 {
			selectedDevice = device
			return selectedDevice, true
		}
	}

	// If that cannot be done, then just pick the device with the most available.
	maxAvalibleFound := math.MinInt32
	for _, device := range devices {
		if solverData.flatHWUUIDMap[device].labLoading.value > maxAvalibleFound {
			maxAvalibleFound = solverData.flatHWUUIDMap[device].labLoading.value
			selectedDevice = device
		}
	}

	return selectedDevice, false
}

// hwSearchOrdering: given the flatUUIDLoadingMap, return the order of least common to most common boards/eqs.
func hwSearchOrdering(flatEqMap map[uint64][]uint64) []kv {
	/*
		 For example: hwEqMap {hw1: [hw1,hw2,hw3], hw3: [hw3]}

		we want to allow hw3 to find matches first. This is to give a chance for all tests which specifically require hw3
		to take prioty on that hw.

		Example problem we are trying to solve (given the hwEq map above)
			- labLoading provides there are 2 devices for each type
			- There is a maximum of 2 tests per shard allowed
			- The HW to TC Loading:
				hw1: [tc1, tc2, ... 8]
				hw3: [tc1 .. 4]

			If we allowed hw1 to be solved first, it could potentially be assigned HW3
			(as it has the same avalbility as HW1/Hw2). Then when the Hw3 TC come to be assigned, all the HW3 devices have been
			loaded.

	*/

	var sortedStruct []kv
	for k, v := range flatEqMap {
		sortedStruct = append(sortedStruct, kv{key: k, value: len(v)})
	}

	sort.Slice(sortedStruct, func(i, j int) bool {
		return sortedStruct[i].value < sortedStruct[j].value
	})
	return sortedStruct
}

// addHWtohwUUIDMap is a helper method to inject into the map without overwriting the keys existing values.
func addHWtohwUUIDMap(hwUUIDMap map[uint64]*api.SchedulingUnitOptions, hw *api.SchedulingUnitOptions) uint64 {
	hash, _ := hashstructure.Hash(hw, hashstructure.FormatV2, nil)
	if _, hwExists := hwUUIDMap[hash]; !hwExists {
		// Add the HW to the lookup map if not seen before.
		hwUUIDMap[hash] = hw
	}
	return hash
}

// TODO; when HwRequirements is fully deprecated, remove this.
// oldAddHWtohwUUIDMap is a helper method to inject into the map without overwriting the keys existing values.
func oldAddHWtohwUUIDMap(hwUUIDMap map[uint64]*api.HWRequirements, hw *api.HWRequirements) uint64 {
	hash, _ := hashstructure.Hash(hw, hashstructure.FormatV2, nil)
	_, hwExists := hwUUIDMap[hash]
	// Add the HW to the lookup map if not seen before.
	if !hwExists {
		hwUUIDMap[hash] = hw
	}
	return hash
}

// addHWtoFlatHWUUIDMap is a helper method to inject into the map without overwriting the keys existing values.
func addHWtoFlatHWUUIDMap(ctx context.Context, flatHWUUIDMap map[uint64]*hwInfo, k uint64, v *hwInfo) error {
	_, exists := flatHWUUIDMap[k]
	// Add the HW to the lookup map if not seen before.
	if !exists {
		flatHWUUIDMap[k] = v
	} else if !reflect.DeepEqual(flatHWUUIDMap[k].req, v.req) {
		flatHWUUIDMap[k] = v
		return fmt.Errorf("mismatch")
	}
	return nil
}

type hashHelper struct {
	S *labapi.Dut
	V string
}

func hashForSchedulingUnit(unit *api.SchedulingUnit) uint64 {
	return hashSchedulingUnit(unit, false)
}

// This only looks at the actual DUTinfo, thus it ignores variant.
func hashForSchedulingUnitHwOnly(unit *api.SchedulingUnit) uint64 {
	return hashSchedulingUnit(unit, true)
}

func hashSchedulingUnit(unit *api.SchedulingUnit, hwOnly bool) uint64 {
	h := hashHelper{S: unit.GetPrimaryTarget().GetSwarmingDef().GetDutInfo()}
	if !hwOnly {
		h.V = unit.GetPrimaryTarget().GetSwarmingDef().GetVariant()
	}

	hasher := []hashHelper{h}
	for _, secondary := range unit.GetCompanionTargets() {

		if secondary.GetSwarmingDef().GetDutInfo() != nil {
			sh := hashHelper{S: secondary.GetSwarmingDef().GetDutInfo()}
			if !hwOnly {
				h.V = secondary.GetSwarmingDef().GetVariant()
			}
			hasher = append(hasher, sh)
		}
	}
	hash, _ := hashstructure.Hash(hasher, hashstructure.FormatV2, nil)
	return hash
}
