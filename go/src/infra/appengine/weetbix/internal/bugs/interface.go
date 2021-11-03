// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bugs

import (
	"errors"
	"infra/appengine/weetbix/internal/analysis"
)

type BugToUpdate struct {
	BugName string
	// Cluster details for the given bug.
	Impact *ClusterImpact
}

var ErrCreateSimulated = errors.New("CreateNew did not create a bug as the bug manager is in simulation mode")

// CreateRequest captures key details of a cluster and its impact,
// as needed for filing new bugs.
type CreateRequest struct {
	// Title is a human-readable display name for the cluster,
	// for use in bug titles.
	Title string
	// Description is a detailed description of the cluster, to appear
	// in the bug body.
	Description string
	// Impact describes the impact of cluster.
	Impact *ClusterImpact
}

// ClusterImpact captures details of a cluster's impact, as needed
// to control the priority and verified status of bugs.
type ClusterImpact struct {
	Failures1d int64
	Failures3d int64
	Failures7d int64
}

func ExtractResidualImpact(cs *analysis.ClusterSummary) *ClusterImpact {
	return &ClusterImpact{
		Failures1d: cs.Failures1d.Residual,
		Failures3d: cs.Failures3d.Residual,
		Failures7d: cs.Failures7d.Residual,
	}
}

func SetResidualImpact(cs *analysis.ClusterSummary, impact *ClusterImpact) {
	cs.Failures1d.Residual = impact.Failures1d
	cs.Failures3d.Residual = impact.Failures3d
	cs.Failures7d.Residual = impact.Failures7d
}
