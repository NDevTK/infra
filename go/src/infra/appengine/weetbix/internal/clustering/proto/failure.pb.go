// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: infra/appengine/weetbix/internal/clustering/proto/failure.proto

package clusteringpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	v1 "infra/appengine/weetbix/proto/v1"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Chunk is a set of unexpected test failures which are processed together
// for efficiency.
// Serialised and stored in GCS.
type Chunk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Failures []*Failure `protobuf:"bytes,1,rep,name=failures,proto3" json:"failures,omitempty"`
}

func (x *Chunk) Reset() {
	*x = Chunk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Chunk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Chunk) ProtoMessage() {}

func (x *Chunk) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Chunk.ProtoReflect.Descriptor instead.
func (*Chunk) Descriptor() ([]byte, []int) {
	return file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDescGZIP(), []int{0}
}

func (x *Chunk) GetFailures() []*Failure {
	if x != nil {
		return x.Failures
	}
	return nil
}

// Weetbix internal representation of an unexpected test failure.
type Failure struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The identity of the test result, as defined by the source system.
	TestResultId *v1.TestResultId `protobuf:"bytes,1,opt,name=test_result_id,json=testResultId,proto3" json:"test_result_id,omitempty"`
	// Timestamp representing the start of the data retention period. This acts
	// as the partitioning key in time/date-partitioned tables.
	PartitionTime *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=partition_time,json=partitionTime,proto3" json:"partition_time,omitempty"`
	// The zero-based index of this failure within the chunk. Assigned by
	// Weetbix ingestion.
	ChunkIndex int64 `protobuf:"varint,3,opt,name=chunk_index,json=chunkIndex,proto3" json:"chunk_index,omitempty"`
	// Security realm of the test result.
	// For test results from ResultDB, this must be set. The format is
	// "{LUCI_PROJECT}:{REALM_SUFFIX}", for example "chromium:ci".
	Realm string `protobuf:"bytes,4,opt,name=realm,proto3" json:"realm,omitempty"`
	// The unique identifier of the test.
	// For test results from ResultDB, see luci.resultdb.v1.TestResult.test_id.
	TestId string `protobuf:"bytes,5,opt,name=test_id,json=testId,proto3" json:"test_id,omitempty"`
	// key:value pairs to specify the way of running a particular test.
	// e.g. a specific bucket, builder and a test suite.
	Variant *v1.Variant `protobuf:"bytes,6,opt,name=variant,proto3" json:"variant,omitempty"`
	// Hash of the variant.
	// hex(sha256(''.join(sorted('%s:%s\n' for k, v in variant.items())))).
	VariantHash string `protobuf:"bytes,7,opt,name=variant_hash,json=variantHash,proto3" json:"variant_hash,omitempty"`
	// A failure reason describing why the test failed.
	FailureReason *v1.FailureReason `protobuf:"bytes,8,opt,name=failure_reason,json=failureReason,proto3" json:"failure_reason,omitempty"`
	// The bug tracking component corresponding to this test case, as identified
	// by the test results system. If no information is available, this is
	// unset.
	BugTrackingComponent *v1.BugTrackingComponent `protobuf:"bytes,9,opt,name=bug_tracking_component,json=bugTrackingComponent,proto3" json:"bug_tracking_component,omitempty"`
	// The point in time when the test case started to execute.
	StartTime *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	// The amount of time the test case took to execute.
	Duration *durationpb.Duration `protobuf:"bytes,11,opt,name=duration,proto3" json:"duration,omitempty"`
	// Was the test failure exonerated? Exonerated means the failure
	// was ignored and did not have further impact, in terms of causing
	// the build to fail or rejecting the CL being tested in a presubmit run.
	IsExonerated bool `protobuf:"varint,12,opt,name=is_exonerated,json=isExonerated,proto3" json:"is_exonerated,omitempty"`
	// Identity of the presubmit run that contains this test result.
	// This should be unique per "CQ+1"/"CQ+2" attempt on gerrit.
	//
	// One presumbit run MAY have many ingested invocation IDs (e.g. for its
	// various tryjobs), but every ingested invocation ID only ever has one
	// presubmit run ID (if any).
	//
	// All test results for the same presubmit run will have one
	// partition_time.
	//
	// If the test result was not collected as part of a presubmit run,
	// this is unset.
	PresubmitRunId *v1.PresubmitRunId `protobuf:"bytes,13,opt,name=presubmit_run_id,json=presubmitRunId,proto3" json:"presubmit_run_id,omitempty"`
	// The invocation from which this test result was ingested. This is
	// the top-level invocation that was ingested, an "invocation" being
	// a container of test results as identified by the source test result
	// system.
	//
	// For ResultDB, Weetbix ingests invocations corresponding to
	// buildbucket builds.
	//
	// All test results ingested from the same invocation (i.e. with the
	// same ingested_invocation_id) will have the same partition time.
	IngestedInvocationId string `protobuf:"bytes,14,opt,name=ingested_invocation_id,json=ingestedInvocationId,proto3" json:"ingested_invocation_id,omitempty"`
	// The zero-based index for this test result, in the sequence of the
	// ingested invocation's results for this test variant. Within the sequence,
	// test results are ordered by start_time and then by test result ID.
	// The first test result is 0, the last test result is
	// ingested_invocation_result_count - 1.
	IngestedInvocationResultIndex int64 `protobuf:"varint,15,opt,name=ingested_invocation_result_index,json=ingestedInvocationResultIndex,proto3" json:"ingested_invocation_result_index,omitempty"`
	// The number of test results having this test variant in the ingested
	// invocation.
	IngestedInvocationResultCount int64 `protobuf:"varint,16,opt,name=ingested_invocation_result_count,json=ingestedInvocationResultCount,proto3" json:"ingested_invocation_result_count,omitempty"`
	// Is the ingested invocation blocked by this test variant? This is
	// only true if all (non-skipped) test results for this test variant
	// (in the ingested invocation) are unexpected failures.
	//
	// Exoneration does not factor into this value; check is_exonerated
	// to see if the impact of this ingested invocation being blocked was
	// mitigated by exoneration.
	IsIngestedInvocationBlocked bool `protobuf:"varint,17,opt,name=is_ingested_invocation_blocked,json=isIngestedInvocationBlocked,proto3" json:"is_ingested_invocation_blocked,omitempty"`
	// The identifier of the test run the test ran in. Test results in different
	// test runs are generally considered independent as they should be unable
	// to leak state to one another.
	//
	// In Chrome and Chrome OS, a test run logically corresponds to a swarming
	// task that runs tests, but this ID is not necessarily the ID of that
	// task, but rather any other ID that is unique per such task.
	//
	// If test result system is ResultDB, this is the ID of the ResultDB
	// invocation the test result was immediately contained within, not including
	// any "invocations/" prefix.
	TestRunId string `protobuf:"bytes,18,opt,name=test_run_id,json=testRunId,proto3" json:"test_run_id,omitempty"`
	// The zero-based index for this test result, in the sequence of results
	// having this test variant and test run. Within the sequence, test
	// results are ordered by start_time and then by test result ID.
	// The first test result is 0, the last test result is
	// test_run_result_count - 1.
	TestRunResultIndex int64 `protobuf:"varint,19,opt,name=test_run_result_index,json=testRunResultIndex,proto3" json:"test_run_result_index,omitempty"`
	// The number of test results having this test variant and test run.
	TestRunResultCount int64 `protobuf:"varint,20,opt,name=test_run_result_count,json=testRunResultCount,proto3" json:"test_run_result_count,omitempty"`
	// Is the test run blocked by this test variant? This is only true if all
	// (non-skipped) test results for this test variant (in the test run)
	// are unexpected failures.
	//
	// Exoneration does not factor into this value; check is_exonerated
	// to see if the impact of this root invocation being blocked was
	// mitigated by exoneration.
	IsTestRunBlocked bool `protobuf:"varint,21,opt,name=is_test_run_blocked,json=isTestRunBlocked,proto3" json:"is_test_run_blocked,omitempty"`
}

func (x *Failure) Reset() {
	*x = Failure{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Failure) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Failure) ProtoMessage() {}

func (x *Failure) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Failure.ProtoReflect.Descriptor instead.
func (*Failure) Descriptor() ([]byte, []int) {
	return file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDescGZIP(), []int{1}
}

func (x *Failure) GetTestResultId() *v1.TestResultId {
	if x != nil {
		return x.TestResultId
	}
	return nil
}

func (x *Failure) GetPartitionTime() *timestamppb.Timestamp {
	if x != nil {
		return x.PartitionTime
	}
	return nil
}

func (x *Failure) GetChunkIndex() int64 {
	if x != nil {
		return x.ChunkIndex
	}
	return 0
}

func (x *Failure) GetRealm() string {
	if x != nil {
		return x.Realm
	}
	return ""
}

func (x *Failure) GetTestId() string {
	if x != nil {
		return x.TestId
	}
	return ""
}

func (x *Failure) GetVariant() *v1.Variant {
	if x != nil {
		return x.Variant
	}
	return nil
}

func (x *Failure) GetVariantHash() string {
	if x != nil {
		return x.VariantHash
	}
	return ""
}

func (x *Failure) GetFailureReason() *v1.FailureReason {
	if x != nil {
		return x.FailureReason
	}
	return nil
}

func (x *Failure) GetBugTrackingComponent() *v1.BugTrackingComponent {
	if x != nil {
		return x.BugTrackingComponent
	}
	return nil
}

func (x *Failure) GetStartTime() *timestamppb.Timestamp {
	if x != nil {
		return x.StartTime
	}
	return nil
}

func (x *Failure) GetDuration() *durationpb.Duration {
	if x != nil {
		return x.Duration
	}
	return nil
}

func (x *Failure) GetIsExonerated() bool {
	if x != nil {
		return x.IsExonerated
	}
	return false
}

func (x *Failure) GetPresubmitRunId() *v1.PresubmitRunId {
	if x != nil {
		return x.PresubmitRunId
	}
	return nil
}

func (x *Failure) GetIngestedInvocationId() string {
	if x != nil {
		return x.IngestedInvocationId
	}
	return ""
}

func (x *Failure) GetIngestedInvocationResultIndex() int64 {
	if x != nil {
		return x.IngestedInvocationResultIndex
	}
	return 0
}

func (x *Failure) GetIngestedInvocationResultCount() int64 {
	if x != nil {
		return x.IngestedInvocationResultCount
	}
	return 0
}

func (x *Failure) GetIsIngestedInvocationBlocked() bool {
	if x != nil {
		return x.IsIngestedInvocationBlocked
	}
	return false
}

func (x *Failure) GetTestRunId() string {
	if x != nil {
		return x.TestRunId
	}
	return ""
}

func (x *Failure) GetTestRunResultIndex() int64 {
	if x != nil {
		return x.TestRunResultIndex
	}
	return 0
}

func (x *Failure) GetTestRunResultCount() int64 {
	if x != nil {
		return x.TestRunResultCount
	}
	return 0
}

func (x *Failure) GetIsTestRunBlocked() bool {
	if x != nil {
		return x.IsTestRunBlocked
	}
	return false
}

var File_infra_appengine_weetbix_internal_clustering_proto_failure_proto protoreflect.FileDescriptor

var file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDesc = []byte{
	0x0a, 0x3f, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2f, 0x77, 0x65, 0x65, 0x74, 0x62, 0x69, 0x78, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x1b, 0x77, 0x65, 0x65, 0x74, 0x62, 0x69, 0x78, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x1a, 0x1f,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x2d, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65,
	0x2f, 0x77, 0x65, 0x65, 0x74, 0x62, 0x69, 0x78, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x76,
	0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x35,
	0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2f,
	0x77, 0x65, 0x65, 0x74, 0x62, 0x69, 0x78, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x76, 0x31,
	0x2f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x49, 0x0a, 0x05, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x12, 0x40,
	0x0a, 0x08, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x24, 0x2e, 0x77, 0x65, 0x65, 0x74, 0x62, 0x69, 0x78, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x46,
	0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x52, 0x08, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x73,
	0x22, 0xe7, 0x08, 0x0a, 0x07, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x12, 0x3e, 0x0a, 0x0e,
	0x74, 0x65, 0x73, 0x74, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x77, 0x65, 0x65, 0x74, 0x62, 0x69, 0x78, 0x2e, 0x76,
	0x31, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x49, 0x64, 0x52, 0x0c,
	0x74, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x49, 0x64, 0x12, 0x41, 0x0a, 0x0e,
	0x70, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x0d, 0x70, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x1f, 0x0a, 0x0b, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x49, 0x6e, 0x64, 0x65, 0x78,
	0x12, 0x14, 0x0a, 0x05, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x69,
	0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12,
	0x2d, 0x0a, 0x07, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x13, 0x2e, 0x77, 0x65, 0x65, 0x74, 0x62, 0x69, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61,
	0x72, 0x69, 0x61, 0x6e, 0x74, 0x52, 0x07, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x12, 0x21,
	0x0a, 0x0c, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x48, 0x61, 0x73,
	0x68, 0x12, 0x40, 0x0a, 0x0e, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x72, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x77, 0x65, 0x65, 0x74,
	0x62, 0x69, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x52, 0x65,
	0x61, 0x73, 0x6f, 0x6e, 0x52, 0x0d, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x52, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x12, 0x56, 0x0a, 0x16, 0x62, 0x75, 0x67, 0x5f, 0x74, 0x72, 0x61, 0x63, 0x6b,
	0x69, 0x6e, 0x67, 0x5f, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x18, 0x09, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x77, 0x65, 0x65, 0x74, 0x62, 0x69, 0x78, 0x2e, 0x76, 0x31,
	0x2e, 0x42, 0x75, 0x67, 0x54, 0x72, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x67, 0x43, 0x6f, 0x6d, 0x70,
	0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x52, 0x14, 0x62, 0x75, 0x67, 0x54, 0x72, 0x61, 0x63, 0x6b, 0x69,
	0x6e, 0x67, 0x43, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x12, 0x39, 0x0a, 0x0a, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x35, 0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x23, 0x0a,
	0x0d, 0x69, 0x73, 0x5f, 0x65, 0x78, 0x6f, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x18, 0x0c,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x69, 0x73, 0x45, 0x78, 0x6f, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x65, 0x64, 0x12, 0x44, 0x0a, 0x10, 0x70, 0x72, 0x65, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x5f,
	0x72, 0x75, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x77,
	0x65, 0x65, 0x74, 0x62, 0x69, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x65, 0x73, 0x75, 0x62,
	0x6d, 0x69, 0x74, 0x52, 0x75, 0x6e, 0x49, 0x64, 0x52, 0x0e, 0x70, 0x72, 0x65, 0x73, 0x75, 0x62,
	0x6d, 0x69, 0x74, 0x52, 0x75, 0x6e, 0x49, 0x64, 0x12, 0x34, 0x0a, 0x16, 0x69, 0x6e, 0x67, 0x65,
	0x73, 0x74, 0x65, 0x64, 0x5f, 0x69, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x69, 0x64, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x14, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74,
	0x65, 0x64, 0x49, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x47,
	0x0a, 0x20, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x64, 0x5f, 0x69, 0x6e, 0x76, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x5f, 0x69, 0x6e, 0x64,
	0x65, 0x78, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x03, 0x52, 0x1d, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74,
	0x65, 0x64, 0x49, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x47, 0x0a, 0x20, 0x69, 0x6e, 0x67, 0x65, 0x73,
	0x74, 0x65, 0x64, 0x5f, 0x69, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x72,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x10, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x1d, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x64, 0x49, 0x6e, 0x76, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x43, 0x0a, 0x1e, 0x69, 0x73, 0x5f, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x64, 0x5f,
	0x69, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x65, 0x64, 0x18, 0x11, 0x20, 0x01, 0x28, 0x08, 0x52, 0x1b, 0x69, 0x73, 0x49, 0x6e, 0x67, 0x65,
	0x73, 0x74, 0x65, 0x64, 0x49, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x65, 0x64, 0x12, 0x1e, 0x0a, 0x0b, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x72, 0x75,
	0x6e, 0x5f, 0x69, 0x64, 0x18, 0x12, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x65, 0x73, 0x74,
	0x52, 0x75, 0x6e, 0x49, 0x64, 0x12, 0x31, 0x0a, 0x15, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x72, 0x75,
	0x6e, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x13,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x12, 0x74, 0x65, 0x73, 0x74, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x31, 0x0a, 0x15, 0x74, 0x65, 0x73, 0x74,
	0x5f, 0x72, 0x75, 0x6e, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x5f, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x18, 0x14, 0x20, 0x01, 0x28, 0x03, 0x52, 0x12, 0x74, 0x65, 0x73, 0x74, 0x52, 0x75, 0x6e,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x2d, 0x0a, 0x13, 0x69,
	0x73, 0x5f, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x72, 0x75, 0x6e, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x65, 0x64, 0x18, 0x15, 0x20, 0x01, 0x28, 0x08, 0x52, 0x10, 0x69, 0x73, 0x54, 0x65, 0x73, 0x74,
	0x52, 0x75, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x42, 0x40, 0x5a, 0x3e, 0x69, 0x6e,
	0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2f, 0x77, 0x65,
	0x65, 0x74, 0x62, 0x69, 0x78, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x63,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3b,
	0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDescOnce sync.Once
	file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDescData = file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDesc
)

func file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDescGZIP() []byte {
	file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDescOnce.Do(func() {
		file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDescData)
	})
	return file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDescData
}

var file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_goTypes = []interface{}{
	(*Chunk)(nil),                   // 0: weetbix.internal.clustering.Chunk
	(*Failure)(nil),                 // 1: weetbix.internal.clustering.Failure
	(*v1.TestResultId)(nil),         // 2: weetbix.v1.TestResultId
	(*timestamppb.Timestamp)(nil),   // 3: google.protobuf.Timestamp
	(*v1.Variant)(nil),              // 4: weetbix.v1.Variant
	(*v1.FailureReason)(nil),        // 5: weetbix.v1.FailureReason
	(*v1.BugTrackingComponent)(nil), // 6: weetbix.v1.BugTrackingComponent
	(*durationpb.Duration)(nil),     // 7: google.protobuf.Duration
	(*v1.PresubmitRunId)(nil),       // 8: weetbix.v1.PresubmitRunId
}
var file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_depIdxs = []int32{
	1, // 0: weetbix.internal.clustering.Chunk.failures:type_name -> weetbix.internal.clustering.Failure
	2, // 1: weetbix.internal.clustering.Failure.test_result_id:type_name -> weetbix.v1.TestResultId
	3, // 2: weetbix.internal.clustering.Failure.partition_time:type_name -> google.protobuf.Timestamp
	4, // 3: weetbix.internal.clustering.Failure.variant:type_name -> weetbix.v1.Variant
	5, // 4: weetbix.internal.clustering.Failure.failure_reason:type_name -> weetbix.v1.FailureReason
	6, // 5: weetbix.internal.clustering.Failure.bug_tracking_component:type_name -> weetbix.v1.BugTrackingComponent
	3, // 6: weetbix.internal.clustering.Failure.start_time:type_name -> google.protobuf.Timestamp
	7, // 7: weetbix.internal.clustering.Failure.duration:type_name -> google.protobuf.Duration
	8, // 8: weetbix.internal.clustering.Failure.presubmit_run_id:type_name -> weetbix.v1.PresubmitRunId
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_init() }
func file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_init() {
	if File_infra_appengine_weetbix_internal_clustering_proto_failure_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Chunk); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Failure); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_goTypes,
		DependencyIndexes: file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_depIdxs,
		MessageInfos:      file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_msgTypes,
	}.Build()
	File_infra_appengine_weetbix_internal_clustering_proto_failure_proto = out.File
	file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_rawDesc = nil
	file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_goTypes = nil
	file_infra_appengine_weetbix_internal_clustering_proto_failure_proto_depIdxs = nil
}
