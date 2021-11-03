// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clustering

import (
	cpb "infra/appengine/weetbix/internal/clustering/proto"
)

// Failure captures the minimal information required to cluster a failure.
// This is a subset of the information captured by Weetbix, additional context
// is stored for analysis.
type Failure struct {
	// The name of the test that failed.
	TestID string
	// The primary error message explaining the reason why the test failed.
	Reason string
}

func FailureFromProto(proto *cpb.Failure) *Failure {
	result := &Failure{
		TestID: proto.TestId,
	}
	if proto.FailureReason != nil {
		result.Reason = proto.FailureReason.PrimaryErrorMessage
	}
	return result
}

func FailuresFromProto(protos []*cpb.Failure) []*Failure {
	result := make([]*Failure, len(protos))
	for i, p := range protos {
		result[i] = FailureFromProto(p)
	}
	return result
}
