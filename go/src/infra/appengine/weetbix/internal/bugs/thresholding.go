// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bugs

import (
	configpb "infra/appengine/weetbix/proto/config"
)

// InflateThreshold inflates or deflates impact thresholds by the given factor.
// This method is provided to help implement hysteresis. inflationPercent can
// be positive or negative (or zero), and is interpreted as follows:
// - If inflationPercent is positive, the new threshold is (threshold * (1 + (inflationPercent/100)))
// - If inflationPercent is negative, the new threshold used is (threshold / (1 + (-inflationPercent/100))
// i.e. inflationPercent of +100 would result in a threshold that is 200% the
// original threshold being used, inflationPercent of -100 would result in a
// threshold that is 50% of the original.
func InflateThreshold(t *configpb.ImpactThreshold, inflationPercent int64) *configpb.ImpactThreshold {
	return &configpb.ImpactThreshold{
		PresubmitRunsFailed: inflateMetricThreshold(t.PresubmitRunsFailed, inflationPercent),
		TestResultsFailed:   inflateMetricThreshold(t.TestResultsFailed, inflationPercent),
		TestRunsFailed:      inflateMetricThreshold(t.TestRunsFailed, inflationPercent),
	}
}

func inflateMetricThreshold(t *configpb.MetricThreshold, inflationPercent int64) *configpb.MetricThreshold {
	if t == nil {
		// No thresholds specified for metric.
		return nil
	}
	return &configpb.MetricThreshold{
		OneDay:   inflateSingleThreshold(t.OneDay, inflationPercent),
		ThreeDay: inflateSingleThreshold(t.ThreeDay, inflationPercent),
		SevenDay: inflateSingleThreshold(t.SevenDay, inflationPercent),
	}
}

func inflateSingleThreshold(threshold *int64, inflationPercent int64) *int64 {
	if threshold == nil {
		// No threshold was specified.
		return nil
	}
	thresholdValue := *threshold
	if inflationPercent >= 0 {
		// I.E. +100% doubles the threshold.
		thresholdValue = (thresholdValue * (100 + inflationPercent)) / 100
	} else {
		// I.E. -100% halves the threshold.
		thresholdValue = (thresholdValue * 100) / (100 + -inflationPercent)
	}
	return &thresholdValue
}

// MeetsThreshold returns whether the nominal impact of the cluster meets
// or exceeds the specified threshold.
func (c *ClusterImpact) MeetsThreshold(t *configpb.ImpactThreshold) bool {
	if c.TestResultsFailed.meetsThreshold(t.TestResultsFailed) {
		return true
	}
	if c.TestRunsFailed.meetsThreshold(t.TestRunsFailed) {
		return true
	}
	if c.PresubmitRunsFailed.meetsThreshold(t.PresubmitRunsFailed) {
		return true
	}
	return false
}

func (m MetricImpact) meetsThreshold(t *configpb.MetricThreshold) bool {
	if t == nil {
		t = &configpb.MetricThreshold{}
	}
	if meetsThreshold(m.OneDay, t.OneDay) {
		return true
	}
	if meetsThreshold(m.ThreeDay, t.ThreeDay) {
		return true
	}
	if meetsThreshold(m.SevenDay, t.SevenDay) {
		return true
	}
	return false
}

// meetsThreshold tests whether value exceeds the given threshold.
// If threshold is nil, the threshold is considered "not set"
// and the method always returns false.
func meetsThreshold(value int64, threshold *int64) bool {
	if threshold == nil {
		return false
	}
	thresholdValue := *threshold
	return value >= thresholdValue
}

// ThresholdExplanation describes a threshold which was evaluated on
// a cluster's impact.
type ThresholdExplanation struct {
	// A human-readable explanation of the metric.
	Metric string
	// The number of days the metric value was measured over.
	TimescaleDays int
	// The threshold value of the metric.
	Threshold int64
}

// ExplainThresholdMet provides an explanation of why cluster impact would
// not have met the given priority threshold. As the overall threshold is an
// 'OR' combination of its underlying thresholds, this returns a list of all
// thresholds which would not have been met by the cluster's impact.
func ExplainThresholdNotMet(threshold *configpb.ImpactThreshold) []ThresholdExplanation {
	var results []ThresholdExplanation
	results = append(results, explainMetricCriteriaNotMet("Presubmit Runs Failed", threshold.PresubmitRunsFailed)...)
	results = append(results, explainMetricCriteriaNotMet("Test Runs Failed", threshold.TestRunsFailed)...)
	results = append(results, explainMetricCriteriaNotMet("Test Results Failed", threshold.TestResultsFailed)...)
	return results
}

func explainMetricCriteriaNotMet(metric string, threshold *configpb.MetricThreshold) []ThresholdExplanation {
	if threshold == nil {
		return nil
	}
	var results []ThresholdExplanation
	if threshold.OneDay != nil {
		results = append(results, ThresholdExplanation{
			Metric:        metric,
			TimescaleDays: 1,
			Threshold:     *threshold.OneDay,
		})
	}
	if threshold.ThreeDay != nil {
		results = append(results, ThresholdExplanation{
			Metric:        metric,
			TimescaleDays: 3,
			Threshold:     *threshold.ThreeDay,
		})
	}
	if threshold.SevenDay != nil {
		results = append(results, ThresholdExplanation{
			Metric:        metric,
			TimescaleDays: 7,
			Threshold:     *threshold.SevenDay,
		})
	}
	return results
}

// ExplainThresholdMet provides an explanation of why the given cluster impact
// met the given priority threshold. As the overall threshold is an 'OR' combination of
// its underlying thresholds, this returns an example of a threshold which a metric
// value exceeded.
func (c *ClusterImpact) ExplainThresholdMet(threshold *configpb.ImpactThreshold) ThresholdExplanation {
	explanation := explainMetricThresholdMet("Presubmit Runs Failed", c.PresubmitRunsFailed, threshold.PresubmitRunsFailed)
	if explanation != nil {
		return *explanation
	}
	explanation = explainMetricThresholdMet("Test Runs Failed", c.TestRunsFailed, threshold.TestRunsFailed)
	if explanation != nil {
		return *explanation
	}
	explanation = explainMetricThresholdMet("Test Results Failed", c.TestResultsFailed, threshold.TestResultsFailed)
	if explanation != nil {
		return *explanation
	}
	// This should not occur, unless the threshold was not met.
	return ThresholdExplanation{}
}

func explainMetricThresholdMet(metric string, impact MetricImpact, threshold *configpb.MetricThreshold) *ThresholdExplanation {
	if threshold == nil {
		return nil
	}
	if threshold.OneDay != nil && impact.OneDay >= *threshold.OneDay {
		return &ThresholdExplanation{
			Metric:        metric,
			TimescaleDays: 1,
			Threshold:     *threshold.OneDay,
		}
	}
	if threshold.ThreeDay != nil && impact.ThreeDay >= *threshold.ThreeDay {
		return &ThresholdExplanation{
			Metric:        metric,
			TimescaleDays: 3,
			Threshold:     *threshold.ThreeDay,
		}
	}
	if threshold.SevenDay != nil && impact.SevenDay >= *threshold.SevenDay {
		return &ThresholdExplanation{
			Metric:        metric,
			TimescaleDays: 7,
			Threshold:     *threshold.SevenDay,
		}
	}
	return nil
}

// MergeThresholdMetExplanations merges multiple explanations for why thresholds
// were met into a minimal list, that removes redundant explanations.
func MergeThresholdMetExplanations(explanations []ThresholdExplanation) []ThresholdExplanation {
	var results []ThresholdExplanation
	for _, exp := range explanations {
		var merged bool
		for i, otherExp := range results {
			if otherExp.Metric == exp.Metric && otherExp.TimescaleDays == exp.TimescaleDays {
				threshold := otherExp.Threshold
				if exp.Threshold > threshold {
					threshold = exp.Threshold
				}
				results[i].Threshold = threshold
				merged = true
				break
			}
		}
		if !merged {
			results = append(results, exp)
		}
	}
	return results
}
