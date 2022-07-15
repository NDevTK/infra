// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package updater

import (
	"infra/appengine/weetbix/internal/analysis"
	"infra/appengine/weetbix/internal/bugs"
)

// ExtractResidualImpact extracts the residual impact from a
// cluster. For suggested clusters, residual impact
// is the impact of the cluster after failures that are already
// part of a bug cluster are removed.
func ExtractResidualImpact(c *analysis.Cluster) *bugs.ClusterImpact {
	return &bugs.ClusterImpact{
		CriticalFailuresExonerated: bugs.MetricImpact{
			OneDay:   c.CriticalFailuresExonerated1d.Residual,
			ThreeDay: c.CriticalFailuresExonerated3d.Residual,
			SevenDay: c.CriticalFailuresExonerated7d.Residual,
		},
		TestResultsFailed: bugs.MetricImpact{
			OneDay:   c.Failures1d.Residual,
			ThreeDay: c.Failures3d.Residual,
			SevenDay: c.Failures7d.Residual,
		},
		TestRunsFailed: bugs.MetricImpact{
			OneDay:   c.TestRunFails1d.Residual,
			ThreeDay: c.TestRunFails3d.Residual,
			SevenDay: c.TestRunFails7d.Residual,
		},
		PresubmitRunsFailed: bugs.MetricImpact{
			OneDay:   c.PresubmitRejects1d.Residual,
			ThreeDay: c.PresubmitRejects3d.Residual,
			SevenDay: c.PresubmitRejects7d.Residual,
		},
	}
}

// SetResidualImpact sets the residual impact on a cluster summary.
func SetResidualImpact(cs *analysis.Cluster, impact *bugs.ClusterImpact) {
	cs.CriticalFailuresExonerated1d.Residual = impact.CriticalFailuresExonerated.OneDay
	cs.CriticalFailuresExonerated3d.Residual = impact.CriticalFailuresExonerated.ThreeDay
	cs.CriticalFailuresExonerated7d.Residual = impact.CriticalFailuresExonerated.SevenDay

	cs.Failures1d.Residual = impact.TestResultsFailed.OneDay
	cs.Failures3d.Residual = impact.TestResultsFailed.ThreeDay
	cs.Failures7d.Residual = impact.TestResultsFailed.SevenDay

	cs.TestRunFails1d.Residual = impact.TestRunsFailed.OneDay
	cs.TestRunFails3d.Residual = impact.TestRunsFailed.ThreeDay
	cs.TestRunFails7d.Residual = impact.TestRunsFailed.SevenDay

	cs.PresubmitRejects1d.Residual = impact.PresubmitRunsFailed.OneDay
	cs.PresubmitRejects3d.Residual = impact.PresubmitRunsFailed.ThreeDay
	cs.PresubmitRejects7d.Residual = impact.PresubmitRunsFailed.SevenDay
}
