// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bugs

import (
	"errors"

	"infra/appengine/weetbix/internal/clustering"
)

type BugUpdateRequest struct {
	// The bug to update.
	Bug BugID
	// Cluster details for the given bug.
	Impact *ClusterImpact
	// Whether the bug should be updated. If this if false, only
	// the BugUpdateRequest is only made to determine if the bug is
	// the duplicate of another bug.
	ShouldUpdate bool
}

type BugUpdateResponse struct {
	// IsDuplicate is set if the bug is a duplicate of another.
	IsDuplicate bool
}

var ErrCreateSimulated = errors.New("CreateNew did not create a bug as the bug manager is in simulation mode")

// CreateRequest captures key details of a cluster and its impact,
// as needed for filing new bugs.
type CreateRequest struct {
	// Description is the human-readable description of the cluster.
	Description *clustering.ClusterDescription
	// Impact describes the impact of cluster.
	Impact *ClusterImpact
	// The monorail components (if any) to use.
	MonorailComponents []string
}

// ClusterImpact captures details of a cluster's impact, as needed
// to control the priority and verified status of bugs.
type ClusterImpact struct {
	CriticalFailuresExonerated MetricImpact
	TestResultsFailed          MetricImpact
	TestRunsFailed             MetricImpact
	PresubmitRunsFailed        MetricImpact
}

// MetricImpact captures impact measurements for one metric, over
// different timescales.
type MetricImpact struct {
	OneDay   int64
	ThreeDay int64
	SevenDay int64
}
