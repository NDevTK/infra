// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.7
// source: infra/libs/vmlab/api/instance.proto

package api

import (
	api "go.chromium.org/chromiumos/config/go/test/api"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// ProviderId is the ID of various VM service providers: GCloud, VM leaser
// service etc.
type ProviderId int32

const (
	ProviderId_UNKNOWN   ProviderId = 0
	ProviderId_GCLOUD    ProviderId = 1
	ProviderId_CLOUDSDK  ProviderId = 2
	ProviderId_VM_LEASER ProviderId = 3
)

// Enum value maps for ProviderId.
var (
	ProviderId_name = map[int32]string{
		0: "UNKNOWN",
		1: "GCLOUD",
		2: "CLOUDSDK",
		3: "VM_LEASER",
	}
	ProviderId_value = map[string]int32{
		"UNKNOWN":   0,
		"GCLOUD":    1,
		"CLOUDSDK":  2,
		"VM_LEASER": 3,
	}
)

func (x ProviderId) Enum() *ProviderId {
	p := new(ProviderId)
	*p = x
	return p
}

func (x ProviderId) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProviderId) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_libs_vmlab_api_instance_proto_enumTypes[0].Descriptor()
}

func (ProviderId) Type() protoreflect.EnumType {
	return &file_infra_libs_vmlab_api_instance_proto_enumTypes[0]
}

func (x ProviderId) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProviderId.Descriptor instead.
func (ProviderId) EnumDescriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_instance_proto_rawDescGZIP(), []int{0}
}

// The VM Leaser environment to connect to.
type Config_VmLeaserBackend_Environment int32

const (
	Config_VmLeaserBackend_ENV_LOCAL      Config_VmLeaserBackend_Environment = 0
	Config_VmLeaserBackend_ENV_STAGING    Config_VmLeaserBackend_Environment = 1
	Config_VmLeaserBackend_ENV_PRODUCTION Config_VmLeaserBackend_Environment = 2
)

// Enum value maps for Config_VmLeaserBackend_Environment.
var (
	Config_VmLeaserBackend_Environment_name = map[int32]string{
		0: "ENV_LOCAL",
		1: "ENV_STAGING",
		2: "ENV_PRODUCTION",
	}
	Config_VmLeaserBackend_Environment_value = map[string]int32{
		"ENV_LOCAL":      0,
		"ENV_STAGING":    1,
		"ENV_PRODUCTION": 2,
	}
)

func (x Config_VmLeaserBackend_Environment) Enum() *Config_VmLeaserBackend_Environment {
	p := new(Config_VmLeaserBackend_Environment)
	*p = x
	return p
}

func (x Config_VmLeaserBackend_Environment) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Config_VmLeaserBackend_Environment) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_libs_vmlab_api_instance_proto_enumTypes[1].Descriptor()
}

func (Config_VmLeaserBackend_Environment) Type() protoreflect.EnumType {
	return &file_infra_libs_vmlab_api_instance_proto_enumTypes[1]
}

func (x Config_VmLeaserBackend_Environment) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Config_VmLeaserBackend_Environment.Descriptor instead.
func (Config_VmLeaserBackend_Environment) EnumDescriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_instance_proto_rawDescGZIP(), []int{4, 1, 0}
}

// VMInstance represents a created VM instance.
type VmInstance struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A unique identifier of the VM that can identify the VM among all configs.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The AddressPort information for SSH.
	Ssh *AddressPort `protobuf:"bytes,2,opt,name=ssh,proto3" json:"ssh,omitempty"`
	// Configuration used to create the instance.
	Config *Config `protobuf:"bytes,3,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *VmInstance) Reset() {
	*x = VmInstance{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VmInstance) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VmInstance) ProtoMessage() {}

func (x *VmInstance) ProtoReflect() protoreflect.Message {
	mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VmInstance.ProtoReflect.Descriptor instead.
func (*VmInstance) Descriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_instance_proto_rawDescGZIP(), []int{0}
}

func (x *VmInstance) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *VmInstance) GetSsh() *AddressPort {
	if x != nil {
		return x.Ssh
	}
	return nil
}

func (x *VmInstance) GetConfig() *Config {
	if x != nil {
		return x.Config
	}
	return nil
}

// Request for the InstanceApi.Create endpoint.
type CreateVmInstanceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Configuration of the backend to start the instance.
	Config *Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	// Optional tags to be associated to the instance.
	Tags map[string]string `protobuf:"bytes,2,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *CreateVmInstanceRequest) Reset() {
	*x = CreateVmInstanceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateVmInstanceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateVmInstanceRequest) ProtoMessage() {}

func (x *CreateVmInstanceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateVmInstanceRequest.ProtoReflect.Descriptor instead.
func (*CreateVmInstanceRequest) Descriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_instance_proto_rawDescGZIP(), []int{1}
}

func (x *CreateVmInstanceRequest) GetConfig() *Config {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *CreateVmInstanceRequest) GetTags() map[string]string {
	if x != nil {
		return x.Tags
	}
	return nil
}

// Request for the InstanceApi.List endpoint.
type ListVmInstancesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Configuration of the backend to list the instance.
	Config *Config `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	// Instances with matching tags will be filtered.
	TagFilters map[string]string `protobuf:"bytes,1,rep,name=tag_filters,json=tagFilters,proto3" json:"tag_filters,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *ListVmInstancesRequest) Reset() {
	*x = ListVmInstancesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListVmInstancesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListVmInstancesRequest) ProtoMessage() {}

func (x *ListVmInstancesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListVmInstancesRequest.ProtoReflect.Descriptor instead.
func (*ListVmInstancesRequest) Descriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_instance_proto_rawDescGZIP(), []int{2}
}

func (x *ListVmInstancesRequest) GetConfig() *Config {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *ListVmInstancesRequest) GetTagFilters() map[string]string {
	if x != nil {
		return x.TagFilters
	}
	return nil
}

// AddressPort represents the SSH address of an VMInstance.
type AddressPort struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// An accessible address: IP, domain, or instance name if in the same network.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// Port number for SSH.
	Port int32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
}

func (x *AddressPort) Reset() {
	*x = AddressPort{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddressPort) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddressPort) ProtoMessage() {}

func (x *AddressPort) ProtoReflect() protoreflect.Message {
	mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddressPort.ProtoReflect.Descriptor instead.
func (*AddressPort) Descriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_instance_proto_rawDescGZIP(), []int{3}
}

func (x *AddressPort) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *AddressPort) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

// TODO(b/250961857): finalize fields and add documentation
// Configuration to specify how to create an instance.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Backend:
	//
	//	*Config_GcloudBackend
	//	*Config_VmLeaserBackend_
	Backend isConfig_Backend `protobuf_oneof:"backend"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_instance_proto_rawDescGZIP(), []int{4}
}

func (m *Config) GetBackend() isConfig_Backend {
	if m != nil {
		return m.Backend
	}
	return nil
}

func (x *Config) GetGcloudBackend() *Config_GCloudBackend {
	if x, ok := x.GetBackend().(*Config_GcloudBackend); ok {
		return x.GcloudBackend
	}
	return nil
}

func (x *Config) GetVmLeaserBackend() *Config_VmLeaserBackend {
	if x, ok := x.GetBackend().(*Config_VmLeaserBackend_); ok {
		return x.VmLeaserBackend
	}
	return nil
}

type isConfig_Backend interface {
	isConfig_Backend()
}

type Config_GcloudBackend struct {
	GcloudBackend *Config_GCloudBackend `protobuf:"bytes,1,opt,name=gcloud_backend,json=gcloudBackend,proto3,oneof"`
}

type Config_VmLeaserBackend_ struct {
	VmLeaserBackend *Config_VmLeaserBackend `protobuf:"bytes,2,opt,name=vm_leaser_backend,json=vmLeaserBackend,proto3,oneof"`
}

func (*Config_GcloudBackend) isConfig_Backend() {}

func (*Config_VmLeaserBackend_) isConfig_Backend() {}

// Gcloud properties. Most properties are passed through to the corresponding
// flags of gcloud. A mandatory field is required when the config is used to
// create an instance.
type Config_GCloudBackend struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// GCP project id. Mandatory
	Project string `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	// GCE zone. Mandatory.
	Zone string `protobuf:"bytes,2,opt,name=zone,proto3" json:"zone,omitempty"`
	// GCE machine type. Mandatory.
	MachineType string `protobuf:"bytes,3,opt,name=machine_type,json=machineType,proto3" json:"machine_type,omitempty"`
	// A custom prefix to instance name. Mandatory.
	InstancePrefix string `protobuf:"bytes,4,opt,name=instance_prefix,json=instancePrefix,proto3" json:"instance_prefix,omitempty"`
	// Network, must be consistent to zone. Optional, fallback to default.
	Network string `protobuf:"bytes,7,opt,name=network,proto3" json:"network,omitempty"`
	// Subnet of network. Optional, fallback to default.
	Subnet string `protobuf:"bytes,8,opt,name=subnet,proto3" json:"subnet,omitempty"`
	// A boolean flag whether to request a public IPv4 address.
	// If requested, ssh target will be the public IPv4 address, otherwise it
	// will be the GCE internal IP address.
	PublicIp bool `protobuf:"varint,5,opt,name=public_ip,json=publicIp,proto3" json:"public_ip,omitempty"`
	// A boolean flag to determine what ip address to return in ssh target.
	// Default is false and the public_ip flag is used for that decision.
	// If true, ssh target will be the GCE internal IP address regardless of
	// whether public_ip is requested.
	AlwaysSshInternalIp bool `protobuf:"varint,9,opt,name=always_ssh_internal_ip,json=alwaysSshInternalIp,proto3" json:"always_ssh_internal_ip,omitempty"`
	// GCE Image to be used to create instance. Mandatory.
	Image *GceImage `protobuf:"bytes,6,opt,name=image,proto3" json:"image,omitempty"`
}

func (x *Config_GCloudBackend) Reset() {
	*x = Config_GCloudBackend{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config_GCloudBackend) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config_GCloudBackend) ProtoMessage() {}

func (x *Config_GCloudBackend) ProtoReflect() protoreflect.Message {
	mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config_GCloudBackend.ProtoReflect.Descriptor instead.
func (*Config_GCloudBackend) Descriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_instance_proto_rawDescGZIP(), []int{4, 0}
}

func (x *Config_GCloudBackend) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

func (x *Config_GCloudBackend) GetZone() string {
	if x != nil {
		return x.Zone
	}
	return ""
}

func (x *Config_GCloudBackend) GetMachineType() string {
	if x != nil {
		return x.MachineType
	}
	return ""
}

func (x *Config_GCloudBackend) GetInstancePrefix() string {
	if x != nil {
		return x.InstancePrefix
	}
	return ""
}

func (x *Config_GCloudBackend) GetNetwork() string {
	if x != nil {
		return x.Network
	}
	return ""
}

func (x *Config_GCloudBackend) GetSubnet() string {
	if x != nil {
		return x.Subnet
	}
	return ""
}

func (x *Config_GCloudBackend) GetPublicIp() bool {
	if x != nil {
		return x.PublicIp
	}
	return false
}

func (x *Config_GCloudBackend) GetAlwaysSshInternalIp() bool {
	if x != nil {
		return x.AlwaysSshInternalIp
	}
	return false
}

func (x *Config_GCloudBackend) GetImage() *GceImage {
	if x != nil {
		return x.Image
	}
	return nil
}

// VM Leaser properties. The fields will be passed to VM Leaser service for
// VM creation and deletion. Required fields must be passed.
//
// NEXT TAG: 4
type Config_VmLeaserBackend struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Env Config_VmLeaserBackend_Environment `protobuf:"varint,3,opt,name=env,proto3,enum=vmlab.api.Config_VmLeaserBackend_Environment" json:"env,omitempty"`
	// The populated fields will specify the requirements for operations on a VM
	// lease. Required.
	VmRequirements *api.VMRequirements `protobuf:"bytes,1,opt,name=vm_requirements,json=vmRequirements,proto3" json:"vm_requirements,omitempty"`
	// Duration of a VM lease. Optional, fallback to service default.
	// This will put a ceiling on time wasted if the client dies.
	LeaseDuration *durationpb.Duration `protobuf:"bytes,2,opt,name=lease_duration,json=leaseDuration,proto3" json:"lease_duration,omitempty"`
}

func (x *Config_VmLeaserBackend) Reset() {
	*x = Config_VmLeaserBackend{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config_VmLeaserBackend) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config_VmLeaserBackend) ProtoMessage() {}

func (x *Config_VmLeaserBackend) ProtoReflect() protoreflect.Message {
	mi := &file_infra_libs_vmlab_api_instance_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config_VmLeaserBackend.ProtoReflect.Descriptor instead.
func (*Config_VmLeaserBackend) Descriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_instance_proto_rawDescGZIP(), []int{4, 1}
}

func (x *Config_VmLeaserBackend) GetEnv() Config_VmLeaserBackend_Environment {
	if x != nil {
		return x.Env
	}
	return Config_VmLeaserBackend_ENV_LOCAL
}

func (x *Config_VmLeaserBackend) GetVmRequirements() *api.VMRequirements {
	if x != nil {
		return x.VmRequirements
	}
	return nil
}

func (x *Config_VmLeaserBackend) GetLeaseDuration() *durationpb.Duration {
	if x != nil {
		return x.LeaseDuration
	}
	return nil
}

var File_infra_libs_vmlab_api_instance_proto protoreflect.FileDescriptor

var file_infra_libs_vmlab_api_instance_proto_rawDesc = []byte{
	0x0a, 0x23, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x6c, 0x69, 0x62, 0x73, 0x2f, 0x76, 0x6d, 0x6c,
	0x61, 0x62, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69,
	0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x20, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x6c, 0x69, 0x62, 0x73, 0x2f, 0x76, 0x6d, 0x6c,
	0x61, 0x62, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x4b, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x2e,
	0x6f, 0x72, 0x67, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x6f, 0x73, 0x2f, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x68, 0x72, 0x6f,
	0x6d, 0x69, 0x75, 0x6d, 0x6f, 0x73, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x6d, 0x5f, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x75, 0x0a, 0x0a, 0x56, 0x6d, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x28, 0x0a, 0x03, 0x73, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16,
	0x2e, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x50, 0x6f, 0x72, 0x74, 0x52, 0x03, 0x73, 0x73, 0x68, 0x12, 0x29, 0x0a, 0x06, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x76, 0x6d,
	0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0xbf, 0x01, 0x0a, 0x17, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x56, 0x6d, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x29, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x11, 0x2e, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x40, 0x0a,
	0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x76, 0x6d,
	0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x56, 0x6d,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e,
	0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x1a,
	0x37, 0x0a, 0x09, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xd6, 0x01, 0x0a, 0x16, 0x4c, 0x69, 0x73,
	0x74, 0x56, 0x6d, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x29, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x52,
	0x0a, 0x0b, 0x74, 0x61, 0x67, 0x5f, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x31, 0x2e, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x4c, 0x69, 0x73, 0x74, 0x56, 0x6d, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x54, 0x61, 0x67, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x74, 0x61, 0x67, 0x46, 0x69, 0x6c, 0x74, 0x65,
	0x72, 0x73, 0x1a, 0x3d, 0x0a, 0x0f, 0x54, 0x61, 0x67, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x22, 0x3b, 0x0a, 0x0b, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x50, 0x6f, 0x72, 0x74,
	0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f,
	0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x91,
	0x06, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x48, 0x0a, 0x0e, 0x67, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1f, 0x2e, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x47, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x42, 0x61, 0x63, 0x6b, 0x65,
	0x6e, 0x64, 0x48, 0x00, 0x52, 0x0d, 0x67, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x42, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x12, 0x4f, 0x0a, 0x11, 0x76, 0x6d, 0x5f, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x72,
	0x5f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21,
	0x2e, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2e, 0x56, 0x6d, 0x4c, 0x65, 0x61, 0x73, 0x65, 0x72, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e,
	0x64, 0x48, 0x00, 0x52, 0x0f, 0x76, 0x6d, 0x4c, 0x65, 0x61, 0x73, 0x65, 0x72, 0x42, 0x61, 0x63,
	0x6b, 0x65, 0x6e, 0x64, 0x1a, 0xb8, 0x02, 0x0a, 0x0d, 0x47, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x42,
	0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x7a, 0x6f, 0x6e, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x7a, 0x6f, 0x6e, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x5f,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6d, 0x61, 0x63, 0x68,
	0x69, 0x6e, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78,
	0x12, 0x18, 0x0a, 0x07, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x75,
	0x62, 0x6e, 0x65, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x75, 0x62, 0x6e,
	0x65, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x69, 0x70, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x49, 0x70, 0x12,
	0x33, 0x0a, 0x16, 0x61, 0x6c, 0x77, 0x61, 0x79, 0x73, 0x5f, 0x73, 0x73, 0x68, 0x5f, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x5f, 0x69, 0x70, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x13, 0x61, 0x6c, 0x77, 0x61, 0x79, 0x73, 0x53, 0x73, 0x68, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x49, 0x70, 0x12, 0x29, 0x0a, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x47, 0x63, 0x65, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x1a,
	0xa5, 0x02, 0x0a, 0x0f, 0x56, 0x6d, 0x4c, 0x65, 0x61, 0x73, 0x65, 0x72, 0x42, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x12, 0x3f, 0x0a, 0x03, 0x65, 0x6e, 0x76, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x2d, 0x2e, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x2e, 0x56, 0x6d, 0x4c, 0x65, 0x61, 0x73, 0x65, 0x72, 0x42, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x2e, 0x45, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x52,
	0x03, 0x65, 0x6e, 0x76, 0x12, 0x4c, 0x0a, 0x0f, 0x76, 0x6d, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69,
	0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e,
	0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x6f, 0x73, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x56, 0x4d, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x6d, 0x65, 0x6e,
	0x74, 0x73, 0x52, 0x0e, 0x76, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x6d, 0x65, 0x6e,
	0x74, 0x73, 0x12, 0x40, 0x0a, 0x0e, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x5f, 0x64, 0x75, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0d, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x44, 0x75, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0x41, 0x0a, 0x0b, 0x45, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d,
	0x65, 0x6e, 0x74, 0x12, 0x0d, 0x0a, 0x09, 0x45, 0x4e, 0x56, 0x5f, 0x4c, 0x4f, 0x43, 0x41, 0x4c,
	0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b, 0x45, 0x4e, 0x56, 0x5f, 0x53, 0x54, 0x41, 0x47, 0x49, 0x4e,
	0x47, 0x10, 0x01, 0x12, 0x12, 0x0a, 0x0e, 0x45, 0x4e, 0x56, 0x5f, 0x50, 0x52, 0x4f, 0x44, 0x55,
	0x43, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x02, 0x42, 0x09, 0x0a, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x65,
	0x6e, 0x64, 0x2a, 0x42, 0x0a, 0x0a, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x49, 0x64,
	0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0a, 0x0a,
	0x06, 0x47, 0x43, 0x4c, 0x4f, 0x55, 0x44, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x43, 0x4c, 0x4f,
	0x55, 0x44, 0x53, 0x44, 0x4b, 0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09, 0x56, 0x4d, 0x5f, 0x4c, 0x45,
	0x41, 0x53, 0x45, 0x52, 0x10, 0x03, 0x42, 0x16, 0x5a, 0x14, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f,
	0x6c, 0x69, 0x62, 0x73, 0x2f, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_libs_vmlab_api_instance_proto_rawDescOnce sync.Once
	file_infra_libs_vmlab_api_instance_proto_rawDescData = file_infra_libs_vmlab_api_instance_proto_rawDesc
)

func file_infra_libs_vmlab_api_instance_proto_rawDescGZIP() []byte {
	file_infra_libs_vmlab_api_instance_proto_rawDescOnce.Do(func() {
		file_infra_libs_vmlab_api_instance_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_libs_vmlab_api_instance_proto_rawDescData)
	})
	return file_infra_libs_vmlab_api_instance_proto_rawDescData
}

var file_infra_libs_vmlab_api_instance_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_infra_libs_vmlab_api_instance_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_infra_libs_vmlab_api_instance_proto_goTypes = []interface{}{
	(ProviderId)(0),                         // 0: vmlab.api.ProviderId
	(Config_VmLeaserBackend_Environment)(0), // 1: vmlab.api.Config.VmLeaserBackend.Environment
	(*VmInstance)(nil),                      // 2: vmlab.api.VmInstance
	(*CreateVmInstanceRequest)(nil),         // 3: vmlab.api.CreateVmInstanceRequest
	(*ListVmInstancesRequest)(nil),          // 4: vmlab.api.ListVmInstancesRequest
	(*AddressPort)(nil),                     // 5: vmlab.api.AddressPort
	(*Config)(nil),                          // 6: vmlab.api.Config
	nil,                                     // 7: vmlab.api.CreateVmInstanceRequest.TagsEntry
	nil,                                     // 8: vmlab.api.ListVmInstancesRequest.TagFiltersEntry
	(*Config_GCloudBackend)(nil),            // 9: vmlab.api.Config.GCloudBackend
	(*Config_VmLeaserBackend)(nil),          // 10: vmlab.api.Config.VmLeaserBackend
	(*GceImage)(nil),                        // 11: vmlab.api.GceImage
	(*api.VMRequirements)(nil),              // 12: chromiumos.test.api.VMRequirements
	(*durationpb.Duration)(nil),             // 13: google.protobuf.Duration
}
var file_infra_libs_vmlab_api_instance_proto_depIdxs = []int32{
	5,  // 0: vmlab.api.VmInstance.ssh:type_name -> vmlab.api.AddressPort
	6,  // 1: vmlab.api.VmInstance.config:type_name -> vmlab.api.Config
	6,  // 2: vmlab.api.CreateVmInstanceRequest.config:type_name -> vmlab.api.Config
	7,  // 3: vmlab.api.CreateVmInstanceRequest.tags:type_name -> vmlab.api.CreateVmInstanceRequest.TagsEntry
	6,  // 4: vmlab.api.ListVmInstancesRequest.config:type_name -> vmlab.api.Config
	8,  // 5: vmlab.api.ListVmInstancesRequest.tag_filters:type_name -> vmlab.api.ListVmInstancesRequest.TagFiltersEntry
	9,  // 6: vmlab.api.Config.gcloud_backend:type_name -> vmlab.api.Config.GCloudBackend
	10, // 7: vmlab.api.Config.vm_leaser_backend:type_name -> vmlab.api.Config.VmLeaserBackend
	11, // 8: vmlab.api.Config.GCloudBackend.image:type_name -> vmlab.api.GceImage
	1,  // 9: vmlab.api.Config.VmLeaserBackend.env:type_name -> vmlab.api.Config.VmLeaserBackend.Environment
	12, // 10: vmlab.api.Config.VmLeaserBackend.vm_requirements:type_name -> chromiumos.test.api.VMRequirements
	13, // 11: vmlab.api.Config.VmLeaserBackend.lease_duration:type_name -> google.protobuf.Duration
	12, // [12:12] is the sub-list for method output_type
	12, // [12:12] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_infra_libs_vmlab_api_instance_proto_init() }
func file_infra_libs_vmlab_api_instance_proto_init() {
	if File_infra_libs_vmlab_api_instance_proto != nil {
		return
	}
	file_infra_libs_vmlab_api_image_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_infra_libs_vmlab_api_instance_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VmInstance); i {
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
		file_infra_libs_vmlab_api_instance_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateVmInstanceRequest); i {
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
		file_infra_libs_vmlab_api_instance_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListVmInstancesRequest); i {
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
		file_infra_libs_vmlab_api_instance_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddressPort); i {
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
		file_infra_libs_vmlab_api_instance_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
		file_infra_libs_vmlab_api_instance_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config_GCloudBackend); i {
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
		file_infra_libs_vmlab_api_instance_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config_VmLeaserBackend); i {
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
	file_infra_libs_vmlab_api_instance_proto_msgTypes[4].OneofWrappers = []interface{}{
		(*Config_GcloudBackend)(nil),
		(*Config_VmLeaserBackend_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_infra_libs_vmlab_api_instance_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_libs_vmlab_api_instance_proto_goTypes,
		DependencyIndexes: file_infra_libs_vmlab_api_instance_proto_depIdxs,
		EnumInfos:         file_infra_libs_vmlab_api_instance_proto_enumTypes,
		MessageInfos:      file_infra_libs_vmlab_api_instance_proto_msgTypes,
	}.Build()
	File_infra_libs_vmlab_api_instance_proto = out.File
	file_infra_libs_vmlab_api_instance_proto_rawDesc = nil
	file_infra_libs_vmlab_api_instance_proto_goTypes = nil
	file_infra_libs_vmlab_api_instance_proto_depIdxs = nil
}
