// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"infra/cros/dutstate"
	"infra/libs/fleet/boxster/swarming"
	skylabInv "infra/libs/skylab/inventory"
	skylabSwarming "infra/libs/skylab/inventory/swarming"
	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	"infra/unifiedfleet/app/controller"
	"infra/unifiedfleet/app/model/configuration"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"
	"infra/unifiedfleet/app/util"
)

// swarmingLabelsDiffHandler generated Swarming labels using new Boxster
// implementation and compares them with the labels generated by the old lib.
func swarmingLabelsDiffHandler(ctx context.Context) error {
	ctx = setupSwarmingDiffContext(ctx)

	// File writer setup
	filename := fmt.Sprintf("swarming_labels_diff/%s.log", time.Now().UTC().Format("2006-01-02T03:04:05"))
	writer, err := getCloudStorageWriter(ctx, filename)
	if err != nil {
		return err
	}
	defer func() {
		if writer != nil {
			if err := writer.Close(); err != nil {
				logging.Warningf(ctx, "failed to close cloud storage writer: %s", err)
			}
		}
	}()

	logs := []string{"############ Swarming Labels Diff ############"}

	// Get DutAttributes and MachineLSEs
	attrs, err := configuration.ListDutAttributes(ctx, false)
	if err != nil {
		return err
	}

	lseRes, err := inventory.GetAllMachineLSEs(ctx)
	if err != nil {
		return err
	}
	if lseRes == nil {
		return errors.New("machine lse entities are missing")
	}

	logs = append(logs, fmt.Sprintf("Found %d MachineLSEs", len(lseRes.Passed())))
	logs = append(logs, "##############################\n\n")

	if _, err := fmt.Fprint(writer, strings.Join(logs, "\n\n")); err != nil {
		return err
	}

	// Map to check if a board-model-sku combo has been checked
	programMap := map[string]bool{}

	for _, r := range lseRes.Passed() {
		lse := r.Data.(*ufspb.MachineLSE)
		logging.Infof(ctx, "Checking Dut %s Machine %s", lse.GetHostname(), lse.GetMachines()[0])

		state, err := controller.GetDutState(ctx, "", lse.GetHostname())
		if err != nil {
			logging.Warningf(ctx, "DutState not found for dut %s", lse.GetHostname())
		}

		// Get first machine attached to the LSE. Intentially dont use
		// GetMachineACL here since this is an automated process.
		m, err := registration.GetMachine(ctx, lse.GetMachines()[0])
		if err != nil {
			return err
		}
		if m == nil {
			return errors.New("machine entity corrupted")
		}

		if crosMachine := m.GetChromeosMachine(); crosMachine != nil {
			var ids string

			// Boxster implementation
			var ufsLabels swarming.Dimensions
			var newLabels []string
			fcId, err := configuration.GenerateFCIdFromCrosMachine(m)
			if err != nil {
				logging.Errorf(ctx, err.Error())
				ids = fmt.Sprintf("\tFlatConfig ID: %s\n", err.Error())
			} else {
				// Skip if checked this combo before
				if programMap[fcId] {
					continue
				}
				ids = fmt.Sprintf("\tFlatConfig ID: %s\n", fcId)
				ufsLabels, err = getUfsLabels(ctx, fcId, attrs, lse, state)
				if err != nil {
					logging.Warningf(ctx, err.Error())
					continue
				}
				for k, v := range ufsLabels {
					newLabels = append(newLabels, fmt.Sprintf("%s:%s", k, strings.Join(v, ",")))
				}
				programMap[fcId] = true
			}

			// Old label implementation
			ids = ids + fmt.Sprintf("\t\tBoard: %s\tModel: %s\tSku: %s", crosMachine.GetBuildTarget(), crosMachine.GetModel(), crosMachine.GetSku())
			oldBotInfo, err := botInfoForDUT(ctx, lse.GetHostname())
			if err != nil {
				continue
			}

			var oldLabels []string
			for k, v := range oldBotInfo.Dimensions {
				oldLabels = append(oldLabels, fmt.Sprintf("%s:%s", k, strings.Join(v, ",")))
			}

			sort.Strings(newLabels)
			sort.Strings(oldLabels)

			// Diff of UFS labels and old labels
			if err := logSwarmingDiff(ids, oldLabels, newLabels, writer); err != nil {
				return err
			}
		}
	}

	return nil
}

func setupSwarmingDiffContext(ctx context.Context) context.Context {
	ctx = logging.SetLevel(ctx, logging.Warning)
	ctx, _ = util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	return ctx
}

func getUfsLabels(ctx context.Context, fcId string, attrs []*api.DutAttribute, lse *ufspb.MachineLSE, state *chromeosLab.DutState) (swarming.Dimensions, error) {
	fc, err := configuration.GetFlatConfig(ctx, fcId)
	if err != nil {
		return nil, err
	}

	ufsLabels := make(swarming.Dimensions)
	for _, dutAttr := range attrs {
		labelsMap, err := controller.Convert(ctx, dutAttr, fc, lse, state)
		if err != nil {
			logging.Errorf(ctx, "Could not get label string for %s %s: %s", fcId, dutAttr.GetId().GetValue(), err)
			continue
		}
		for k, v := range labelsMap {
			ufsLabels[k] = v
		}
	}

	return ufsLabels, nil
}

func logSwarmingDiff(ids string, oldLabels, ufsLabels []string, writer *storage.Writer) error {
	var logs []string
	inOld, notInOld := sliceDiff(oldLabels, ufsLabels)
	logs = append(logs, fmt.Sprintf("Identifiers: %s", ids))
	logs = append(logs, fmt.Sprintf("Labels matched: %s", strings.Join(inOld, "\n\t")))
	logs = append(logs, fmt.Sprintf("Labels missing or mismatched: %s", strings.Join(notInOld, "\n\t")))
	logs = append(logs, fmt.Sprintf("Old labels: %s", strings.Join(oldLabels, "\n\t")))
	logs = append(logs, fmt.Sprintf("UFS labels: %s", strings.Join(ufsLabels, "\n\t")))
	logs = append(logs, "##############################\n\n")
	if _, err := fmt.Fprint(writer, strings.Join(logs, "\n\n")); err != nil {
		return err
	}
	return nil
}

func sliceDiff(a, b []string) ([]string, []string) {
	bMap := make(map[string]bool, len(b))
	for _, x := range b {
		bMap[strings.ToLower(x)] = true
	}

	var inA []string
	var notInA []string
	for _, x := range a {
		_, found := bMap[strings.ToLower(x)]
		if found {
			inA = append(inA, x)
		} else {
			notInA = append(notInA, x)
		}
	}
	return inA, notInA
}

// OLD IMPLEMENTATION
//
// This is the current implementation of label generation. Logic is copied from
// infra/go/src/infra/cmd/shivas/internal/ufs/cmds/bot/internal-print-bot-info.go

type botInfo struct {
	Dimensions skylabSwarming.Dimensions
	State      botState
}

type botState map[string][]string

func botInfoForDUT(ctx context.Context, hostname string) (*botInfo, error) {
	r := func(e error) { logging.Infof(ctx, "sanitize dimensions: %s\n", e) }

	data, err := controller.GetChromeOSDeviceData(ctx, "", hostname)
	if err != nil {
		return nil, err
	}

	return &botInfo{
		Dimensions: botDimensionsForDUT(data.GetDutV1(), readDutState(ctx, data.GetLabConfig().GetName()), r),
		State:      botStateForDUT(data),
	}, nil
}

func botStateForDUT(data *ufspb.ChromeOSDeviceData) botState {
	d := data.GetDutV1()
	s := make(botState)
	for _, kv := range d.GetCommon().GetAttributes() {
		k, v := kv.GetKey(), kv.GetValue()
		s[k] = append(s[k], v)
	}
	s["storage_state"] = []string{d.GetCommon().GetLabels().GetPeripherals().GetStorageState().String()[len("HARDWARE_"):]}
	s["servo_usb_state"] = []string{d.GetCommon().GetLabels().GetPeripherals().GetServoUsbState().String()[len("HARDWARE_"):]}
	s["battery_state"] = []string{d.GetCommon().GetLabels().GetPeripherals().GetBatteryState().String()[len("HARDWARE_"):]}
	s["wifi_state"] = []string{d.GetCommon().GetLabels().GetPeripherals().GetWifiState().String()[len("HARDWARE_"):]}
	s["bluetooth_state"] = []string{d.GetCommon().GetLabels().GetPeripherals().GetBluetoothState().String()[len("HARDWARE_"):]}
	s["rpm_state"] = []string{d.GetCommon().GetLabels().GetPeripherals().GetRpmState().String()}
	s["lab_config_version_index"] = []string{data.GetLabConfig().GetUpdateTime().AsTime().Format(util.TimestampBasedVersionKeyFormat)}
	s["dut_state_version_index"] = []string{data.GetDutState().GetUpdateTime().AsTime().Format(util.TimestampBasedVersionKeyFormat)}
	return s
}

func botDimensionsForDUT(d *skylabInv.DeviceUnderTest, ds dutstate.Info, r skylabSwarming.ReportFunc) skylabSwarming.Dimensions {
	c := d.GetCommon()
	dims := skylabSwarming.Convert(c.GetLabels())
	dims["dut_id"] = []string{c.GetId()}
	dims["dut_name"] = []string{c.GetHostname()}
	if v := c.GetHwid(); v != "" {
		dims["hwid"] = []string{v}
	}
	if v := c.GetSerialNumber(); v != "" {
		dims["serial_number"] = []string{v}
	}
	if v := c.GetLocation(); v != nil {
		dims["location"] = []string{formatLocation(v)}
	}
	dims["dut_state"] = []string{string(ds.State)}
	skylabSwarming.Sanitize(dims, r)
	return dims
}

func formatLocation(loc *skylabInv.Location) string {
	return fmt.Sprintf("%s-row%d-rack%d-host%d",
		loc.GetLab().GetName(),
		loc.GetRow(),
		loc.GetRack(),
		loc.GetHost(),
	)
}

// readDutState reads state from UFS.
//
// Logic copied from infra/go/src/infra/cros/dutstate/dutstate.go
func readDutState(ctx context.Context, hostname string) dutstate.Info {
	ctx, err := util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	if err != nil {
		logging.Errorf(ctx, err.Error())
	}

	logging.Infof(ctx, "dutstate: Try to read DUT/Labstation state for %s", hostname)
	res, err := controller.GetMachineLSE(ctx, util.AddPrefix(util.MachineLSECollection, hostname))
	if err != nil {
		if status.Code(err) == codes.NotFound {
			logging.Errorf(ctx, "dutstate: DUT/Labstation not found for %s; %s", hostname, err)
		} else {
			logging.Errorf(ctx, "dutstate: Fail to get DUT/Labstation for %s; %s", hostname, err)
		}
		// For default state time will not set and equal 0.
		return dutstate.Info{
			State: dutstate.Unknown,
		}
	}
	return dutstate.Info{
		State: dutstate.ConvertFromUFSState(res.GetResourceState()),
		Time:  res.GetUpdateTime().Seconds,
	}
}
