// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package migrator

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
)

func TestComputeBoardModelToState(t *testing.T) {
	t.Parallel()
	m := &migrator{}
	t.Run("Happy path", func(t *testing.T) {
		t.Parallel()
		mcs := []*ufspb.Machine{
			{
				Name: "machines/machine-1",
				Device: &ufspb.Machine_ChromeosMachine{
					ChromeosMachine: &ufspb.ChromeOSMachine{
						BuildTarget: "board-1",
						Model:       "model-1",
					},
				},
			},
			{
				Name: "machines/machine-2",
				Device: &ufspb.Machine_ChromeosMachine{
					ChromeosMachine: &ufspb.ChromeOSMachine{
						BuildTarget: "board-1",
						Model:       "model-1",
					},
				},
			},
			{
				Name: "machines/machine-3",
				Device: &ufspb.Machine_ChromeosMachine{
					ChromeosMachine: &ufspb.ChromeOSMachine{
						BuildTarget: "board-1",
						Model:       "model-1",
					},
				},
			},
			{
				Name: "machines/machine-4",
				Device: &ufspb.Machine_ChromeosMachine{
					ChromeosMachine: &ufspb.ChromeOSMachine{
						BuildTarget: "board-2",
						Model:       "model-1",
					},
				},
			},
		}
		lses := []*ufspb.MachineLSE{
			{
				Name: "machineLSEs/dut-1",
				Machines: []string{
					"machine-1",
				},
				Lse: &ufspb.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufspb.ChromeOSMachineLSE{
						ChromeosLse: &ufspb.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufspb.ChromeOSDeviceLSE{
								Device: &ufspb.ChromeOSDeviceLSE_Dut{
									Dut: &chromeosLab.DeviceUnderTest{
										Hive: "cloudbots",
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "machineLSEs/dut-2",
				Machines: []string{
					"machine-2",
				},
				Lse: &ufspb.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufspb.ChromeOSMachineLSE{
						ChromeosLse: &ufspb.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufspb.ChromeOSDeviceLSE{
								Device: &ufspb.ChromeOSDeviceLSE_Dut{
									Dut: &chromeosLab.DeviceUnderTest{
										Hive: "cloudbots",
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "machineLSEs/dut-3",
				Machines: []string{
					"machine-3",
				},
				Lse: &ufspb.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufspb.ChromeOSMachineLSE{
						ChromeosLse: &ufspb.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufspb.ChromeOSDeviceLSE{
								Device: &ufspb.ChromeOSDeviceLSE_Dut{
									Dut: &chromeosLab.DeviceUnderTest{
										Hive: "e",
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "machineLSEs/dut-4",
				Machines: []string{
					"machine-4",
				},
				Lse: &ufspb.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufspb.ChromeOSMachineLSE{
						ChromeosLse: &ufspb.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufspb.ChromeOSDeviceLSE{
								Device: &ufspb.ChromeOSDeviceLSE_Dut{
									Dut: &chromeosLab.DeviceUnderTest{
										Hive: "",
									},
								},
							},
						},
					},
				},
			},
		}
		cs := &configSearchable{
			overrideDUTs: map[string]struct{}{
				"dut-1": {},
			},
		}
		got, err := m.ComputeBoardModelToState(context.Background(), mcs, lses, cs)
		if err != nil {
			t.Fatalf("should not error: %v", err)
		}
		want := map[string]*migrationState{
			"board-1/model-1": {
				Cloudbots: []string{
					"dut-2",
				},
				Drone: []string{
					"dut-3",
				},
			},
			"board-2/model-1": {
				Drone: []string{
					"dut-4",
				},
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
}
