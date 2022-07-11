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
	// Impact for the given bug. This is only set if valid impact is available,
	// if re-clustering is currently ongoing for the failure association rule
	// and impact is unreliable, this will be unset to avoid erroneous
	// priority updates.
	Impact *ClusterImpact
	// Whether the user enabled priority updates and auto-closure for the bug.
	// If this if false, the BugUpdateRequest is only made to determine if the
	// bug is the duplicate of another bug and if the rule should be archived.
	IsManagingBug bool
	// The identity of the rule associated with the bug.
	RuleID string
}

type BugUpdateResponse struct {
	// IsDuplicate is set if the bug is a duplicate of another.
	IsDuplicate bool

	// ShouldArchive indicates if the rule for this bug should be archived.
	// This should be set if:
	// - The bug is managed by Weetbix (IsManagingBug = true) and it has
	//   been marked as Closed (verified) by Weetbix for the last 30 days.
	// - The bug is managed by the user (IsManagingBug = false), and the
	//   bug has been closed for the last 30 days.
	ShouldArchive bool
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
