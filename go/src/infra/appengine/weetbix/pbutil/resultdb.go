// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package pbutil contains methods for manipulating Weetbix protos.
package pbutil

import (
	"go.chromium.org/luci/resultdb/pbutil"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"

	pb "infra/appengine/weetbix/proto/v1"
)

// TestResultIDFromResultDB returns a Weetbix TestResultId corresponding to the
// supplied ResultDB test result name.
// The format of name should be:
// "invocations/{INVOCATION_ID}/tests/{URL_ESCAPED_TEST_ID}/results/{RESULT_ID}".
func TestResultIDFromResultDB(name string) *pb.TestResultId {
	return &pb.TestResultId{System: "resultdb", Id: name}
}

// VariantFromResultDB returns a Weetbix Variant corresponding to the
// supplied ResultDB Variant.
func VariantFromResultDB(v *rdbpb.Variant) *pb.Variant {
	if v == nil {
		// Variant is optional in ResultDB.
		return &pb.Variant{Def: make(map[string]string)}
	}
	return &pb.Variant{Def: v.Def}
}

// VariantToResultDB returns a ResultDB Variant corresponding to the
// supplied Weetbix Variant.
func VariantToResultDB(v *pb.Variant) *rdbpb.Variant {
	if v == nil {
		return &rdbpb.Variant{Def: make(map[string]string)}
	}
	return &rdbpb.Variant{Def: v.Def}
}

// VariantHash returns a hash of the variant.
func VariantHash(v *pb.Variant) string {
	return pbutil.VariantHash(VariantToResultDB(v))
}

// StringPairFromResultDB returns a Weetbix StringPair corresponding to the
// supplied ResultDB StringPair.
func StringPairFromResultDB(v []*rdbpb.StringPair) []*pb.StringPair {
	pairs := []*pb.StringPair{}
	for _, pair := range v {
		pairs = append(pairs, &pb.StringPair{Key: pair.Key, Value: pair.Value})
	}
	return pairs
}

// FailureReasonFromResultDB returns a Weetbix FailureReason corresponding to the
// supplied ResultDB FailureReason.
func FailureReasonFromResultDB(fr *rdbpb.FailureReason) *pb.FailureReason {
	if fr == nil {
		return nil
	}
	return &pb.FailureReason{
		PrimaryErrorMessage: fr.PrimaryErrorMessage,
	}
}

// TestMetadataFromResultDB converts a ResultDB TestMetadata to a Weetbix
// TestMetadata.
func TestMetadataFromResultDB(rdbTmd *rdbpb.TestMetadata) *pb.TestMetadata {
	if rdbTmd == nil {
		return nil
	}

	tmd := &pb.TestMetadata{
		Name: rdbTmd.Name,
	}
	loc := tmd.GetLocation()
	if loc != nil {
		tmd.Location = &pb.TestLocation{
			Repo:     loc.Repo,
			FileName: loc.FileName,
			Line:     loc.Line,
		}
	}

	return tmd
}

// TestResultStatus returns the Weetbix test result status corresponding
// to the given ResultDB test result status.
func TestResultStatusFromResultDB(s rdbpb.TestStatus) pb.TestResultStatus {
	switch s {
	case rdbpb.TestStatus_ABORT:
		return pb.TestResultStatus_ABORT
	case rdbpb.TestStatus_CRASH:
		return pb.TestResultStatus_CRASH
	case rdbpb.TestStatus_FAIL:
		return pb.TestResultStatus_FAIL
	case rdbpb.TestStatus_PASS:
		return pb.TestResultStatus_PASS
	case rdbpb.TestStatus_SKIP:
		return pb.TestResultStatus_SKIP
	default:
		return pb.TestResultStatus_TEST_RESULT_STATUS_UNSPECIFIED
	}
}

// ExonerationReasonFromResultDB converts a ResultDB ExonerationReason to a
// Weetbix ExonerationReason.
func ExonerationReasonFromResultDB(s rdbpb.ExonerationReason) pb.ExonerationReason {
	switch s {
	case rdbpb.ExonerationReason_NOT_CRITICAL:
		return pb.ExonerationReason_NOT_CRITICAL
	case rdbpb.ExonerationReason_OCCURS_ON_MAINLINE:
		return pb.ExonerationReason_OCCURS_ON_MAINLINE
	case rdbpb.ExonerationReason_OCCURS_ON_OTHER_CLS:
		return pb.ExonerationReason_OCCURS_ON_OTHER_CLS
	case rdbpb.ExonerationReason_UNEXPECTED_PASS:
		return pb.ExonerationReason_UNEXPECTED_PASS
	default:
		return pb.ExonerationReason_EXONERATION_REASON_UNSPECIFIED
	}
}
