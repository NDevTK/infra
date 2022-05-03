// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package monorail

import (
	"github.com/golang/protobuf/proto"

	"infra/appengine/weetbix/internal/bugs"
	configpb "infra/appengine/weetbix/proto/config"
	mpb "infra/monorailv2/api/v3/api_proto"
)

// ChromiumTestPriorityField is the resource name of the priority field
// that is consistent with ChromiumTestConfig.
const ChromiumTestPriorityField = "projects/chromium/fieldDefs/11"

// ChromiumTestTypeField is the resource name of the type field
// that is consistent with ChromiumTestConfig.
const ChromiumTestTypeField = "projects/chromium/fieldDefs/10"

// ChromiumTestConfig provides chromium-like configuration for tests
// to use.
func ChromiumTestConfig() *configpb.MonorailProject {
	projectCfg := &configpb.MonorailProject{
		Project: "chromium",
		DefaultFieldValues: []*configpb.MonorailFieldValue{
			{
				FieldId: 10,
				Value:   "Bug",
			},
		},
		PriorityFieldId: 11,
		Priorities: []*configpb.MonorailPriority{
			{
				Priority: "0",
				Threshold: &configpb.ImpactThreshold{
					TestResultsFailed: &configpb.MetricThreshold{
						OneDay: proto.Int64(1000),
					},
					TestRunsFailed: &configpb.MetricThreshold{
						OneDay: proto.Int64(100),
					},
				},
			},
			{
				Priority: "1",
				Threshold: &configpb.ImpactThreshold{
					TestResultsFailed: &configpb.MetricThreshold{
						OneDay: proto.Int64(500),
					},
					TestRunsFailed: &configpb.MetricThreshold{
						OneDay: proto.Int64(50),
					},
				},
			},
			{
				Priority: "2",
				Threshold: &configpb.ImpactThreshold{
					TestResultsFailed: &configpb.MetricThreshold{
						OneDay: proto.Int64(100),
					},
					TestRunsFailed: &configpb.MetricThreshold{
						OneDay: proto.Int64(10),
					},
				},
			},
			{
				Priority: "3",
				// Should be less onerous than the bug-filing thresholds
				// used in BugUpdater tests, to avoid bugs that were filed
				// from being immediately closed.
				Threshold: &configpb.ImpactThreshold{
					TestResultsFailed: &configpb.MetricThreshold{
						OneDay:   proto.Int64(50),
						ThreeDay: proto.Int64(300),
						SevenDay: proto.Int64(1), // Set to 1 so that we check hysteresis never rounds down to 0 and prevents bugs from closing.
					},
				},
			},
		},
		PriorityHysteresisPercent: 10,
	}
	return projectCfg
}

// ChromiumP0Impact returns cluster impact that is consistent with a P0 bug.
func ChromiumP0Impact() *bugs.ClusterImpact {
	return &bugs.ClusterImpact{
		TestResultsFailed: bugs.MetricImpact{
			OneDay: 1500,
		},
	}
}

// ChromiumP1Impact returns cluster impact that is consistent with a P1 bug.
func ChromiumP1Impact() *bugs.ClusterImpact {
	return &bugs.ClusterImpact{
		TestResultsFailed: bugs.MetricImpact{
			OneDay: 750,
		},
	}
}

// ChromiumLowP1Impact returns cluster impact that is consistent with a P1
// bug, but if hysteresis is applied, could also be compatible with P2.
func ChromiumLowP1Impact() *bugs.ClusterImpact {
	return &bugs.ClusterImpact{
		// (500 * (1.0 + PriorityHysteresisPercent / 100.0)) - 1
		TestResultsFailed: bugs.MetricImpact{
			OneDay: 549,
		},
	}
}

// ChromiumP2Impact returns cluster impact that is consistent with a P2 bug.
func ChromiumP2Impact() *bugs.ClusterImpact {
	return &bugs.ClusterImpact{
		TestResultsFailed: bugs.MetricImpact{
			OneDay: 300,
		},
	}
}

// ChromiumHighP3Impact returns cluster impact that is consistent with a P3
// bug, but if hysteresis is applied, could also be compatible with P2.
func ChromiumHighP3Impact() *bugs.ClusterImpact {
	return &bugs.ClusterImpact{
		// (100 / (1.0 + PriorityHysteresisPercent / 100.0)) + 1
		TestResultsFailed: bugs.MetricImpact{
			OneDay: 91,
		},
	}
}

// ChromiumP3Impact returns cluster impact that is consistent with a P3 bug.
func ChromiumP3Impact() *bugs.ClusterImpact {
	return &bugs.ClusterImpact{
		TestResultsFailed: bugs.MetricImpact{
			OneDay: 75,
		},
	}
}

// ChromiumP3LowImpact returns cluster impact that is consistent with a P3
// bug, but if hysteresis is applied, could also be compatible with a closed
// (verified) bug.
func ChromiumP3LowImpact() *bugs.ClusterImpact {
	return &bugs.ClusterImpact{
		// (50 * (1.0 + PriorityHysteresisPercent / 100.0)) - 1
		TestResultsFailed: bugs.MetricImpact{
			OneDay: 54,
		},
	}
}

// ChromiumClosureHighImpact returns cluster impact that is consistent with a
// closed (verified) bug, but if hysteresis is applied, could also be
// compatible with a P3 bug.
func ChromiumClosureHighImpact() *bugs.ClusterImpact {
	return &bugs.ClusterImpact{
		// (50 / (1.0 + PriorityHysteresisPercent / 100.0)) + 1
		TestResultsFailed: bugs.MetricImpact{
			OneDay: 46,
		},
	}
}

// ChromiumClosureImpact returns cluster impact that is consistent with a
// closed (verified) bug.
func ChromiumClosureImpact() *bugs.ClusterImpact {
	return &bugs.ClusterImpact{}
}

// ChromiumTestIssuePriority returns the priority of an issue, assuming
// it has been created consistent with ChromiumTestConfig.
func ChromiumTestIssuePriority(issue *mpb.Issue) string {
	for _, fv := range issue.FieldValues {
		if fv.Field == ChromiumTestPriorityField {
			return fv.Value
		}
	}
	return ""
}
