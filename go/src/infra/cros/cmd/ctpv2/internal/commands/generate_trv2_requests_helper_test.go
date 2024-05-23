// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	. "infra/cros/cmd/ctpv2/internal/commands"
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
)

func TestSchedulingMatch(t *testing.T) {
	metadataMap := fakeSchedulingUnitsMetadataMap()

	// Primary and companions with models.
	primary := FakeHwTarget("dedede", "model", "one")
	companions := []*HwTarget{
		FakeHwTarget("pixel6", "pixel6pro", ""),
		FakeHwTarget("pixel6", "pixel6pro", ""),
	}
	schedUnit := FindSchedulingUnit(primary, companions, metadataMap)
	if schedUnit == nil {
		t.Fatalf("Expected scheduling unit match, found none.")
	}
	if schedUnit.GetCompanionTargets()[0].GetSwarmingDef().GetDutInfo().GetChromeos().GetDutModel().GetModelName() != "pixel6pro" {
		t.Fatalf("Expected pixel6pro, got %s", schedUnit.GetCompanionTargets()[0].GetSwarmingDef().GetDutInfo().GetChromeos().GetDutModel().GetModelName())
	}
	if schedUnit.GetCompanionTargets()[1].GetSwarmingDef().GetDutInfo().GetChromeos().GetDutModel().GetModelName() != "" {
		t.Fatalf("Expected pixel6pro, got %s", schedUnit.GetCompanionTargets()[0].GetSwarmingDef().GetDutInfo().GetChromeos().GetDutModel().GetModelName())
	}

	// Primary and companions without models.
	primary = FakeHwTarget("dedede", "", "one")
	companions = []*HwTarget{
		FakeHwTarget("pixel6", "", ""),
	}
	schedUnit = FindSchedulingUnit(primary, companions, metadataMap)
	if schedUnit == nil {
		t.Fatalf("Expected scheduling unit match, found none.")
	}

	// Primary without variant.
	primary = FakeHwTarget("dedede", "", "")
	companions = []*HwTarget{
		FakeHwTarget("pixel6", "", ""),
	}
	schedUnit = FindSchedulingUnit(primary, companions, metadataMap)
	if schedUnit == nil {
		t.Fatalf("Expected scheduling unit match, found none.")
	}

	// Primary match, but companions do not.
	primary = FakeHwTarget("dedede", "", "one")
	companions = []*HwTarget{
		FakeHwTarget("pixel6", "pixel6pro", ""),
		FakeHwTarget("pixel6", "pixel6pro", ""),
		FakeHwTarget("pixel6", "", ""),
	}
	schedUnit = FindSchedulingUnit(primary, companions, metadataMap)
	if schedUnit != nil {
		t.Fatalf("Expected no scheduling unit match, found one.")
	}

	// Primary has no match.
	primary = FakeHwTarget("dedede", "", "two")
	companions = []*HwTarget{
		FakeHwTarget("pixel6", "pixel6pro", ""),
	}
	schedUnit = FindSchedulingUnit(primary, companions, metadataMap)
	if schedUnit != nil {
		t.Fatalf("Expected no scheduling unit match, found one.")
	}
}

func fakeSchedulingUnitsMetadataMap() map[string][]*api.SchedulingUnit {
	return map[string][]*api.SchedulingUnit{
		"dedede-one": {
			{
				PrimaryTarget: fakeApiTarget("dedede", "model", "one"),
				CompanionTargets: []*api.Target{
					fakeApiTarget("pixel6", "pixel6pro", ""),
					fakeApiTarget("pixel6", "", ""),
				},
			},
			{
				PrimaryTarget: fakeApiTarget("dedede", "model", "one"),
				CompanionTargets: []*api.Target{
					fakeApiTarget("pixel7", "pixel7pro", ""),
				},
			},
		},
		"dedede-": {
			{
				PrimaryTarget: fakeApiTarget("dedede", "model", ""),
				CompanionTargets: []*api.Target{
					fakeApiTarget("pixel6", "pixel6pro", ""),
				},
			},
			{
				PrimaryTarget: fakeApiTarget("dedede", "model", ""),
				CompanionTargets: []*api.Target{
					fakeApiTarget("pixel7", "pixel7pro", ""),
				},
			},
		},
	}
}

func fakeApiTarget(board, model, variant string) *api.Target {
	return &api.Target{
		SwarmingDef: &api.SwarmingDefinition{
			DutInfo: &labapi.Dut{
				DutType: &labapi.Dut_Chromeos{
					Chromeos: &labapi.Dut_ChromeOS{
						DutModel: &labapi.DutModel{
							ModelName:   model,
							BuildTarget: board,
						},
					},
				},
			},
			Variant: variant,
		},
	}
}
