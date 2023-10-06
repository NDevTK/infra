// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.3
// source: infra/experimental/golangbuild/golangbuildpb/params.proto

package golangbuildpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// golangbuild runs in one of these modes.
type Mode int32

const (
	// MODE_ALL builds and tests the project all within the same build.
	Mode_MODE_ALL Mode = 0
	// MODE_COORDINATOR launches and coordinates tasks that build Go and
	// test the provided project.
	Mode_MODE_COORDINATOR Mode = 1
	// MODE_BUILD indicates golangbuild should just run make.bash.
	Mode_MODE_BUILD Mode = 2
	// MODE_TEST indicates golangbuild should only run tests.
	//
	// A prebuilt toolchain must be available for the provided source.
	Mode_MODE_TEST Mode = 3
)

// Enum value maps for Mode.
var (
	Mode_name = map[int32]string{
		0: "MODE_ALL",
		1: "MODE_COORDINATOR",
		2: "MODE_BUILD",
		3: "MODE_TEST",
	}
	Mode_value = map[string]int32{
		"MODE_ALL":         0,
		"MODE_COORDINATOR": 1,
		"MODE_BUILD":       2,
		"MODE_TEST":        3,
	}
)

func (x Mode) Enum() *Mode {
	p := new(Mode)
	*p = x
	return p
}

func (x Mode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Mode) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_enumTypes[0].Descriptor()
}

func (Mode) Type() protoreflect.EnumType {
	return &file_infra_experimental_golangbuild_golangbuildpb_params_proto_enumTypes[0]
}

func (x Mode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Mode.Descriptor instead.
func (Mode) EnumDescriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{0}
}

// Input properties.
type Inputs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the Gerrit project to be tested.
	Project string `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	// GoBranch specifies what Go toolchain branch to use.
	// Its value is a branch name like "master" or "release-branch.go1.20"
	// (the "refs/heads/" prefix is omitted).
	GoBranch string `protobuf:"bytes,2,opt,name=go_branch,json=goBranch,proto3" json:"go_branch,omitempty"`
	// GoCommit optionally specifies a commit on GoBranch branch to use.
	// If it isn't provided, it means to use the tip of the GoBranch branch (a moving target).
	// It can be set only when invoked in a project other than "go".
	// Its value is a commit ID like "4368e1cdfd37cbcdbc7a4fbcc78ad61139f7ba90".
	//
	// Note: this property is mutable in the builder configuration.
	GoCommit string `protobuf:"bytes,19,opt,name=go_commit,json=goCommit,proto3" json:"go_commit,omitempty"`
	// BootstrapVersion specifies the version of Go to use as the bootstrap
	// toolchain when needed.
	BootstrapVersion string `protobuf:"bytes,16,opt,name=bootstrap_version,json=bootstrapVersion,proto3" json:"bootstrap_version,omitempty"`
	// LongTest controls whether the build runs in long test mode.
	LongTest bool `protobuf:"varint,3,opt,name=long_test,json=longTest,proto3" json:"long_test,omitempty"`
	// RaceMode controls whether the build runs with the race detector enabled.
	RaceMode bool `protobuf:"varint,4,opt,name=race_mode,json=raceMode,proto3" json:"race_mode,omitempty"`
	// NoNetwork controls whether the build disables network access during test execution.
	//
	// It's meant to catch tests that accidentally need internet without realizing it,
	// or otherwise forget to skip themselves when testing.Short() is true.
	//
	// This mode is only supported on Linux systems with unshare and ip available.
	// The build fails if the check is on but its system requirements are unmet.
	NoNetwork bool `protobuf:"varint,17,opt,name=no_network,json=noNetwork,proto3" json:"no_network,omitempty"`
	// CompileOnly controls whether the build does as much as possible to check
	// for problems but stops short of process execution for the target OS/arch.
	// It will compile packages and tests (but not execute tests), run static
	// analysis (such as vet), and so on.
	//
	// This mode makes it possible for any highly available machine type to run
	// a quick smoke test for any port.
	CompileOnly bool `protobuf:"varint,5,opt,name=compile_only,json=compileOnly,proto3" json:"compile_only,omitempty"`
	// MiscPorts controls whether the build is run across miscellaneous ports
	// instead of the default single current port.
	//
	// It exists because misc-compile builders are inherently faster at testing
	// a single port in CompileOnly-mode, so they compensate for that by testing
	// multiple ports at once.
	//
	// MiscPorts is only supported when CompileOnly is also set.
	MiscPorts bool `protobuf:"varint,18,opt,name=misc_ports,json=miscPorts,proto3" json:"misc_ports,omitempty"`
	// Extra environment variables to set for building and testing.
	Env map[string]string `protobuf:"bytes,6,rep,name=env,proto3" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Named cache configured on the builder to use as a cipd tool root cache.
	//
	// Required.
	ToolsCache string `protobuf:"bytes,7,opt,name=tools_cache,json=toolsCache,proto3" json:"tools_cache,omitempty"`
	// Named cache configured on the builder to use as a git clone cache.
	//
	// Required.
	GitCache string `protobuf:"bytes,8,opt,name=git_cache,json=gitCache,proto3" json:"git_cache,omitempty"`
	// On Macs, the version of Xcode to use. Because installing it is expensive,
	// it should be the same for all builders that run on a given host.
	XcodeVersion string `protobuf:"bytes,9,opt,name=xcode_version,json=xcodeVersion,proto3" json:"xcode_version,omitempty"`
	// Which mode to run golangbuild in. See the Mode enum for details.
	Mode Mode `protobuf:"varint,10,opt,name=mode,proto3,enum=golangbuildpb.Mode" json:"mode,omitempty"`
	// Properties specific to MODE_ALL.
	AllMode *AllMode `protobuf:"bytes,11,opt,name=all_mode,json=allMode,proto3" json:"all_mode,omitempty"`
	// Properties specific to MODE_COORDINATOR.
	CoordMode *CoordinatorMode `protobuf:"bytes,12,opt,name=coord_mode,json=coordMode,proto3" json:"coord_mode,omitempty"`
	// Properties specific to MODE_BUILD.
	BuildMode *BuildMode `protobuf:"bytes,13,opt,name=build_mode,json=buildMode,proto3" json:"build_mode,omitempty"`
	// Properties specific to MODE_TEST.
	TestMode *TestMode `protobuf:"bytes,14,opt,name=test_mode,json=testMode,proto3" json:"test_mode,omitempty"`
	// Test shard identity. This property is specific to MODE_TEST.
	//
	// Note: this property is mutable in the builder configuration.
	//
	// N.B. This cannot be part of "TestMode" without making "test_mode"
	// mutable too, because property mutability is only controllable at the top level.
	// See allowed_property_overrides in https://chromium.googlesource.com/infra/luci/luci-go/+/HEAD/lucicfg/doc/README.md#luci.builder-args.
	TestShard *TestShard `protobuf:"bytes,15,opt,name=test_shard,json=testShard,proto3" json:"test_shard,omitempty"`
}

func (x *Inputs) Reset() {
	*x = Inputs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Inputs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Inputs) ProtoMessage() {}

func (x *Inputs) ProtoReflect() protoreflect.Message {
	mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Inputs.ProtoReflect.Descriptor instead.
func (*Inputs) Descriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{0}
}

func (x *Inputs) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

func (x *Inputs) GetGoBranch() string {
	if x != nil {
		return x.GoBranch
	}
	return ""
}

func (x *Inputs) GetGoCommit() string {
	if x != nil {
		return x.GoCommit
	}
	return ""
}

func (x *Inputs) GetBootstrapVersion() string {
	if x != nil {
		return x.BootstrapVersion
	}
	return ""
}

func (x *Inputs) GetLongTest() bool {
	if x != nil {
		return x.LongTest
	}
	return false
}

func (x *Inputs) GetRaceMode() bool {
	if x != nil {
		return x.RaceMode
	}
	return false
}

func (x *Inputs) GetNoNetwork() bool {
	if x != nil {
		return x.NoNetwork
	}
	return false
}

func (x *Inputs) GetCompileOnly() bool {
	if x != nil {
		return x.CompileOnly
	}
	return false
}

func (x *Inputs) GetMiscPorts() bool {
	if x != nil {
		return x.MiscPorts
	}
	return false
}

func (x *Inputs) GetEnv() map[string]string {
	if x != nil {
		return x.Env
	}
	return nil
}

func (x *Inputs) GetToolsCache() string {
	if x != nil {
		return x.ToolsCache
	}
	return ""
}

func (x *Inputs) GetGitCache() string {
	if x != nil {
		return x.GitCache
	}
	return ""
}

func (x *Inputs) GetXcodeVersion() string {
	if x != nil {
		return x.XcodeVersion
	}
	return ""
}

func (x *Inputs) GetMode() Mode {
	if x != nil {
		return x.Mode
	}
	return Mode_MODE_ALL
}

func (x *Inputs) GetAllMode() *AllMode {
	if x != nil {
		return x.AllMode
	}
	return nil
}

func (x *Inputs) GetCoordMode() *CoordinatorMode {
	if x != nil {
		return x.CoordMode
	}
	return nil
}

func (x *Inputs) GetBuildMode() *BuildMode {
	if x != nil {
		return x.BuildMode
	}
	return nil
}

func (x *Inputs) GetTestMode() *TestMode {
	if x != nil {
		return x.TestMode
	}
	return nil
}

func (x *Inputs) GetTestShard() *TestShard {
	if x != nil {
		return x.TestShard
	}
	return nil
}

// AllMode contains properties specific to MODE_ALL.
type AllMode struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *AllMode) Reset() {
	*x = AllMode{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AllMode) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AllMode) ProtoMessage() {}

func (x *AllMode) ProtoReflect() protoreflect.Message {
	mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AllMode.ProtoReflect.Descriptor instead.
func (*AllMode) Descriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{1}
}

// CoordinatorMode contains properties specific to MODE_COORDINATOR.
type CoordinatorMode struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the builder to create a build for for building Go.
	BuildBuilder string `protobuf:"bytes,1,opt,name=build_builder,json=buildBuilder,proto3" json:"build_builder,omitempty"`
	// Name of the builder to create a build for to run tests.
	TestBuilder string `protobuf:"bytes,2,opt,name=test_builder,json=testBuilder,proto3" json:"test_builder,omitempty"`
	// Number of separate builds to spawn to run tests in shards.
	NumTestShards uint32 `protobuf:"varint,3,opt,name=num_test_shards,json=numTestShards,proto3" json:"num_test_shards,omitempty"`
	// Names of other builders to trigger for testing when the
	// build_builder completes successfully.
	//
	// This must be empty if project != "go".
	BuildersToTriggerAfterToolchainBuild []string `protobuf:"bytes,4,rep,name=builders_to_trigger_after_toolchain_build,json=buildersToTriggerAfterToolchainBuild,proto3" json:"builders_to_trigger_after_toolchain_build,omitempty"`
	// TargetGoos is the OS we're building for.
	TargetGoos string `protobuf:"bytes,5,opt,name=target_goos,json=targetGoos,proto3" json:"target_goos,omitempty"`
	// TargetGoarch is the CPU architecture we're building for.
	TargetGoarch string `protobuf:"bytes,6,opt,name=target_goarch,json=targetGoarch,proto3" json:"target_goarch,omitempty"`
}

func (x *CoordinatorMode) Reset() {
	*x = CoordinatorMode{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CoordinatorMode) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CoordinatorMode) ProtoMessage() {}

func (x *CoordinatorMode) ProtoReflect() protoreflect.Message {
	mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CoordinatorMode.ProtoReflect.Descriptor instead.
func (*CoordinatorMode) Descriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{2}
}

func (x *CoordinatorMode) GetBuildBuilder() string {
	if x != nil {
		return x.BuildBuilder
	}
	return ""
}

func (x *CoordinatorMode) GetTestBuilder() string {
	if x != nil {
		return x.TestBuilder
	}
	return ""
}

func (x *CoordinatorMode) GetNumTestShards() uint32 {
	if x != nil {
		return x.NumTestShards
	}
	return 0
}

func (x *CoordinatorMode) GetBuildersToTriggerAfterToolchainBuild() []string {
	if x != nil {
		return x.BuildersToTriggerAfterToolchainBuild
	}
	return nil
}

func (x *CoordinatorMode) GetTargetGoos() string {
	if x != nil {
		return x.TargetGoos
	}
	return ""
}

func (x *CoordinatorMode) GetTargetGoarch() string {
	if x != nil {
		return x.TargetGoarch
	}
	return ""
}

// BuildMode contains properties specific to MODE_BUILD.
type BuildMode struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *BuildMode) Reset() {
	*x = BuildMode{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildMode) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildMode) ProtoMessage() {}

func (x *BuildMode) ProtoReflect() protoreflect.Message {
	mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildMode.ProtoReflect.Descriptor instead.
func (*BuildMode) Descriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{3}
}

// TestMode contains properties specific to MODE_TEST.
type TestMode struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TestMode) Reset() {
	*x = TestMode{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestMode) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestMode) ProtoMessage() {}

func (x *TestMode) ProtoReflect() protoreflect.Message {
	mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestMode.ProtoReflect.Descriptor instead.
func (*TestMode) Descriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{4}
}

// TestShard is specific to MODE_TEST and represents the build's
// test shard identity.
//
// Note: this is mutable by ScheduleBuild, so add fields with care.
type TestShard struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID of the test shard. This is always less than num_shards.
	ShardId uint32 `protobuf:"varint,1,opt,name=shard_id,json=shardId,proto3" json:"shard_id,omitempty"`
	// Number of test shards.
	NumShards uint32 `protobuf:"varint,2,opt,name=num_shards,json=numShards,proto3" json:"num_shards,omitempty"`
}

func (x *TestShard) Reset() {
	*x = TestShard{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestShard) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestShard) ProtoMessage() {}

func (x *TestShard) ProtoReflect() protoreflect.Message {
	mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestShard.ProtoReflect.Descriptor instead.
func (*TestShard) Descriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{5}
}

func (x *TestShard) GetShardId() uint32 {
	if x != nil {
		return x.ShardId
	}
	return 0
}

func (x *TestShard) GetNumShards() uint32 {
	if x != nil {
		return x.NumShards
	}
	return 0
}

// Output properties.
type Outputs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Human-friendly failure description.
	Failure *FailureSummary `protobuf:"bytes,1,opt,name=failure,proto3" json:"failure,omitempty"`
}

func (x *Outputs) Reset() {
	*x = Outputs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Outputs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Outputs) ProtoMessage() {}

func (x *Outputs) ProtoReflect() protoreflect.Message {
	mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Outputs.ProtoReflect.Descriptor instead.
func (*Outputs) Descriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{6}
}

func (x *Outputs) GetFailure() *FailureSummary {
	if x != nil {
		return x.Failure
	}
	return nil
}

// FailureSummary summarizes a failure without all the build step structure.
//
// It's intended to be easily rendered as something meant for humans.
type FailureSummary struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Human-friendly one-line plain text description of the failure.
	Description string `protobuf:"bytes,1,opt,name=description,proto3" json:"description,omitempty"`
	// Links (typically to logs) that would be helpful in diagnosing the failure.
	Links []*Link `protobuf:"bytes,2,rep,name=links,proto3" json:"links,omitempty"`
}

func (x *FailureSummary) Reset() {
	*x = FailureSummary{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FailureSummary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FailureSummary) ProtoMessage() {}

func (x *FailureSummary) ProtoReflect() protoreflect.Message {
	mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FailureSummary.ProtoReflect.Descriptor instead.
func (*FailureSummary) Descriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{7}
}

func (x *FailureSummary) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *FailureSummary) GetLinks() []*Link {
	if x != nil {
		return x.Links
	}
	return nil
}

// Link is a URL with a human-friendly name.
type Link struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the link.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The link's URL.
	Url string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *Link) Reset() {
	*x = Link{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Link) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Link) ProtoMessage() {}

func (x *Link) ProtoReflect() protoreflect.Message {
	mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Link.ProtoReflect.Descriptor instead.
func (*Link) Descriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{8}
}

func (x *Link) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Link) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

var File_infra_experimental_golangbuild_golangbuildpb_params_proto protoreflect.FileDescriptor

var file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDesc = []byte{
	0x0a, 0x39, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6d, 0x65,
	0x6e, 0x74, 0x61, 0x6c, 0x2f, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x2f, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x2f, 0x70,
	0x61, 0x72, 0x61, 0x6d, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x67, 0x6f, 0x6c,
	0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x22, 0xb4, 0x06, 0x0a, 0x06, 0x49,
	0x6e, 0x70, 0x75, 0x74, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12,
	0x1b, 0x0a, 0x09, 0x67, 0x6f, 0x5f, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x67, 0x6f, 0x42, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x12, 0x1b, 0x0a, 0x09,
	0x67, 0x6f, 0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x18, 0x13, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x67, 0x6f, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x2b, 0x0a, 0x11, 0x62, 0x6f, 0x6f,
	0x74, 0x73, 0x74, 0x72, 0x61, 0x70, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x10,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x62, 0x6f, 0x6f, 0x74, 0x73, 0x74, 0x72, 0x61, 0x70, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x5f, 0x74,
	0x65, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x6c, 0x6f, 0x6e, 0x67, 0x54,
	0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x72, 0x61, 0x63, 0x65, 0x5f, 0x6d, 0x6f, 0x64, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x72, 0x61, 0x63, 0x65, 0x4d, 0x6f, 0x64, 0x65,
	0x12, 0x1d, 0x0a, 0x0a, 0x6e, 0x6f, 0x5f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x18, 0x11,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x6e, 0x6f, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x12,
	0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c, 0x65, 0x5f, 0x6f, 0x6e, 0x6c, 0x79, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c, 0x65, 0x4f, 0x6e,
	0x6c, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x6d, 0x69, 0x73, 0x63, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x73,
	0x18, 0x12, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x6d, 0x69, 0x73, 0x63, 0x50, 0x6f, 0x72, 0x74,
	0x73, 0x12, 0x30, 0x0a, 0x03, 0x65, 0x6e, 0x76, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e,
	0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x2e, 0x49,
	0x6e, 0x70, 0x75, 0x74, 0x73, 0x2e, 0x45, 0x6e, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03,
	0x65, 0x6e, 0x76, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x5f, 0x63, 0x61, 0x63,
	0x68, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x43,
	0x61, 0x63, 0x68, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x67, 0x69, 0x74, 0x5f, 0x63, 0x61, 0x63, 0x68,
	0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x67, 0x69, 0x74, 0x43, 0x61, 0x63, 0x68,
	0x65, 0x12, 0x23, 0x0a, 0x0d, 0x78, 0x63, 0x6f, 0x64, 0x65, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x78, 0x63, 0x6f, 0x64, 0x65, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x27, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x0a,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69,
	0x6c, 0x64, 0x70, 0x62, 0x2e, 0x4d, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12,
	0x31, 0x0a, 0x08, 0x61, 0x6c, 0x6c, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x70,
	0x62, 0x2e, 0x41, 0x6c, 0x6c, 0x4d, 0x6f, 0x64, 0x65, 0x52, 0x07, 0x61, 0x6c, 0x6c, 0x4d, 0x6f,
	0x64, 0x65, 0x12, 0x3d, 0x0a, 0x0a, 0x63, 0x6f, 0x6f, 0x72, 0x64, 0x5f, 0x6d, 0x6f, 0x64, 0x65,
	0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x2e, 0x43, 0x6f, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x74,
	0x6f, 0x72, 0x4d, 0x6f, 0x64, 0x65, 0x52, 0x09, 0x63, 0x6f, 0x6f, 0x72, 0x64, 0x4d, 0x6f, 0x64,
	0x65, 0x12, 0x37, 0x0a, 0x0a, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x18,
	0x0d, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75,
	0x69, 0x6c, 0x64, 0x70, 0x62, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x4d, 0x6f, 0x64, 0x65, 0x52,
	0x09, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x34, 0x0a, 0x09, 0x74, 0x65,
	0x73, 0x74, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e,
	0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x2e, 0x54, 0x65,
	0x73, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x52, 0x08, 0x74, 0x65, 0x73, 0x74, 0x4d, 0x6f, 0x64, 0x65,
	0x12, 0x37, 0x0a, 0x0a, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x73, 0x68, 0x61, 0x72, 0x64, 0x18, 0x0f,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69,
	0x6c, 0x64, 0x70, 0x62, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x53, 0x68, 0x61, 0x72, 0x64, 0x52, 0x09,
	0x74, 0x65, 0x73, 0x74, 0x53, 0x68, 0x61, 0x72, 0x64, 0x1a, 0x36, 0x0a, 0x08, 0x45, 0x6e, 0x76,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x22, 0x09, 0x0a, 0x07, 0x41, 0x6c, 0x6c, 0x4d, 0x6f, 0x64, 0x65, 0x22, 0xa0, 0x02, 0x0a,
	0x0f, 0x43, 0x6f, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x74, 0x6f, 0x72, 0x4d, 0x6f, 0x64, 0x65,
	0x12, 0x23, 0x0a, 0x0d, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x5f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x42, 0x75,
	0x69, 0x6c, 0x64, 0x65, 0x72, 0x12, 0x21, 0x0a, 0x0c, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x62, 0x75,
	0x69, 0x6c, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x74, 0x65, 0x73,
	0x74, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x12, 0x26, 0x0a, 0x0f, 0x6e, 0x75, 0x6d, 0x5f,
	0x74, 0x65, 0x73, 0x74, 0x5f, 0x73, 0x68, 0x61, 0x72, 0x64, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x0d, 0x6e, 0x75, 0x6d, 0x54, 0x65, 0x73, 0x74, 0x53, 0x68, 0x61, 0x72, 0x64, 0x73,
	0x12, 0x57, 0x0a, 0x29, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x73, 0x5f, 0x74, 0x6f, 0x5f,
	0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x5f, 0x61, 0x66, 0x74, 0x65, 0x72, 0x5f, 0x74, 0x6f,
	0x6f, 0x6c, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x18, 0x04, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x24, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x73, 0x54, 0x6f, 0x54,
	0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x41, 0x66, 0x74, 0x65, 0x72, 0x54, 0x6f, 0x6f, 0x6c, 0x63,
	0x68, 0x61, 0x69, 0x6e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x5f, 0x67, 0x6f, 0x6f, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x47, 0x6f, 0x6f, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x74, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x5f, 0x67, 0x6f, 0x61, 0x72, 0x63, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x47, 0x6f, 0x61, 0x72, 0x63, 0x68, 0x22,
	0x0b, 0x0a, 0x09, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x4d, 0x6f, 0x64, 0x65, 0x22, 0x0a, 0x0a, 0x08,
	0x54, 0x65, 0x73, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x22, 0x45, 0x0a, 0x09, 0x54, 0x65, 0x73, 0x74,
	0x53, 0x68, 0x61, 0x72, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x68, 0x61, 0x72, 0x64, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x73, 0x68, 0x61, 0x72, 0x64, 0x49, 0x64,
	0x12, 0x1d, 0x0a, 0x0a, 0x6e, 0x75, 0x6d, 0x5f, 0x73, 0x68, 0x61, 0x72, 0x64, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x6e, 0x75, 0x6d, 0x53, 0x68, 0x61, 0x72, 0x64, 0x73, 0x22,
	0x42, 0x0a, 0x07, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x12, 0x37, 0x0a, 0x07, 0x66, 0x61,
	0x69, 0x6c, 0x75, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x67, 0x6f,
	0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x2e, 0x46, 0x61, 0x69, 0x6c,
	0x75, 0x72, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x52, 0x07, 0x66, 0x61, 0x69, 0x6c,
	0x75, 0x72, 0x65, 0x22, 0x5d, 0x0a, 0x0e, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x53, 0x75,
	0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x29, 0x0a, 0x05, 0x6c, 0x69, 0x6e, 0x6b, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x2e, 0x4c, 0x69, 0x6e, 0x6b, 0x52, 0x05, 0x6c, 0x69, 0x6e,
	0x6b, 0x73, 0x22, 0x2c, 0x0a, 0x04, 0x4c, 0x69, 0x6e, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x10,
	0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c,
	0x2a, 0x49, 0x0a, 0x04, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x0c, 0x0a, 0x08, 0x4d, 0x4f, 0x44, 0x45,
	0x5f, 0x41, 0x4c, 0x4c, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x4d, 0x4f, 0x44, 0x45, 0x5f, 0x43,
	0x4f, 0x4f, 0x52, 0x44, 0x49, 0x4e, 0x41, 0x54, 0x4f, 0x52, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a,
	0x4d, 0x4f, 0x44, 0x45, 0x5f, 0x42, 0x55, 0x49, 0x4c, 0x44, 0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09,
	0x4d, 0x4f, 0x44, 0x45, 0x5f, 0x54, 0x45, 0x53, 0x54, 0x10, 0x03, 0x42, 0x2e, 0x5a, 0x2c, 0x69,
	0x6e, 0x66, 0x72, 0x61, 0x2f, 0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6d, 0x65, 0x6e, 0x74, 0x61,
	0x6c, 0x2f, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2f, 0x67, 0x6f,
	0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescOnce sync.Once
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescData = file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDesc
)

func file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP() []byte {
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescOnce.Do(func() {
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescData)
	})
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescData
}

var file_infra_experimental_golangbuild_golangbuildpb_params_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_infra_experimental_golangbuild_golangbuildpb_params_proto_goTypes = []interface{}{
	(Mode)(0),               // 0: golangbuildpb.Mode
	(*Inputs)(nil),          // 1: golangbuildpb.Inputs
	(*AllMode)(nil),         // 2: golangbuildpb.AllMode
	(*CoordinatorMode)(nil), // 3: golangbuildpb.CoordinatorMode
	(*BuildMode)(nil),       // 4: golangbuildpb.BuildMode
	(*TestMode)(nil),        // 5: golangbuildpb.TestMode
	(*TestShard)(nil),       // 6: golangbuildpb.TestShard
	(*Outputs)(nil),         // 7: golangbuildpb.Outputs
	(*FailureSummary)(nil),  // 8: golangbuildpb.FailureSummary
	(*Link)(nil),            // 9: golangbuildpb.Link
	nil,                     // 10: golangbuildpb.Inputs.EnvEntry
}
var file_infra_experimental_golangbuild_golangbuildpb_params_proto_depIdxs = []int32{
	10, // 0: golangbuildpb.Inputs.env:type_name -> golangbuildpb.Inputs.EnvEntry
	0,  // 1: golangbuildpb.Inputs.mode:type_name -> golangbuildpb.Mode
	2,  // 2: golangbuildpb.Inputs.all_mode:type_name -> golangbuildpb.AllMode
	3,  // 3: golangbuildpb.Inputs.coord_mode:type_name -> golangbuildpb.CoordinatorMode
	4,  // 4: golangbuildpb.Inputs.build_mode:type_name -> golangbuildpb.BuildMode
	5,  // 5: golangbuildpb.Inputs.test_mode:type_name -> golangbuildpb.TestMode
	6,  // 6: golangbuildpb.Inputs.test_shard:type_name -> golangbuildpb.TestShard
	8,  // 7: golangbuildpb.Outputs.failure:type_name -> golangbuildpb.FailureSummary
	9,  // 8: golangbuildpb.FailureSummary.links:type_name -> golangbuildpb.Link
	9,  // [9:9] is the sub-list for method output_type
	9,  // [9:9] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_infra_experimental_golangbuild_golangbuildpb_params_proto_init() }
func file_infra_experimental_golangbuild_golangbuildpb_params_proto_init() {
	if File_infra_experimental_golangbuild_golangbuildpb_params_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Inputs); i {
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
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AllMode); i {
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
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CoordinatorMode); i {
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
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuildMode); i {
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
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestMode); i {
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
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestShard); i {
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
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Outputs); i {
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
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FailureSummary); i {
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
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Link); i {
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
			RawDescriptor: file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_experimental_golangbuild_golangbuildpb_params_proto_goTypes,
		DependencyIndexes: file_infra_experimental_golangbuild_golangbuildpb_params_proto_depIdxs,
		EnumInfos:         file_infra_experimental_golangbuild_golangbuildpb_params_proto_enumTypes,
		MessageInfos:      file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes,
	}.Build()
	File_infra_experimental_golangbuild_golangbuildpb_params_proto = out.File
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDesc = nil
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_goTypes = nil
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_depIdxs = nil
}
