// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/appengine/cros/lab_inventory/app/config/config.proto

package config

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type LuciAuthGroup struct {
	Value                string   `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LuciAuthGroup) Reset()         { *m = LuciAuthGroup{} }
func (m *LuciAuthGroup) String() string { return proto.CompactTextString(m) }
func (*LuciAuthGroup) ProtoMessage()    {}
func (*LuciAuthGroup) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe4d2048e3798022, []int{0}
}

func (m *LuciAuthGroup) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LuciAuthGroup.Unmarshal(m, b)
}
func (m *LuciAuthGroup) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LuciAuthGroup.Marshal(b, m, deterministic)
}
func (m *LuciAuthGroup) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LuciAuthGroup.Merge(m, src)
}
func (m *LuciAuthGroup) XXX_Size() int {
	return xxx_messageInfo_LuciAuthGroup.Size(m)
}
func (m *LuciAuthGroup) XXX_DiscardUnknown() {
	xxx_messageInfo_LuciAuthGroup.DiscardUnknown(m)
}

var xxx_messageInfo_LuciAuthGroup proto.InternalMessageInfo

func (m *LuciAuthGroup) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

// Config is the configuration data served by luci-config for this app.
type Config struct {
	// AdminService contains information about the skylab admin instances.
	AdminService *AdminService `protobuf:"bytes,2,opt,name=admin_service,json=adminService,proto3" json:"admin_service,omitempty"`
	// The access groups of the inventory.
	Readers                   *LuciAuthGroup `protobuf:"bytes,3,opt,name=readers,proto3" json:"readers,omitempty"`
	StatusWriters             *LuciAuthGroup `protobuf:"bytes,4,opt,name=status_writers,json=statusWriters,proto3" json:"status_writers,omitempty"`
	SetupWriters              *LuciAuthGroup `protobuf:"bytes,5,opt,name=setup_writers,json=setupWriters,proto3" json:"setup_writers,omitempty"`
	PrivilegedWriters         *LuciAuthGroup `protobuf:"bytes,6,opt,name=privileged_writers,json=privilegedWriters,proto3" json:"privileged_writers,omitempty"`
	HwidSecret                string         `protobuf:"bytes,7,opt,name=hwid_secret,json=hwidSecret,proto3" json:"hwid_secret,omitempty"`
	DeviceConfigSource        *Gitiles       `protobuf:"bytes,8,opt,name=device_config_source,json=deviceConfigSource,proto3" json:"device_config_source,omitempty"`
	ManufacturingConfigSource *Gitiles       `protobuf:"bytes,9,opt,name=manufacturing_config_source,json=manufacturingConfigSource,proto3" json:"manufacturing_config_source,omitempty"`
	// The git repo information of inventory v1.
	// TODO(guocb) remove this after migration.
	Inventory *InventoryV1Repo `protobuf:"bytes,12,opt,name=inventory,proto3" json:"inventory,omitempty"`
	// Environment managed by this instance of app, e.g. ENVIRONMENT_STAGING,
	// ENVIRONMENT_PROD, etc.
	Environment string `protobuf:"bytes,10,opt,name=environment,proto3" json:"environment,omitempty"`
	// The hostname of drone-queen service to push inventory to.
	QueenService string `protobuf:"bytes,11,opt,name=queen_service,json=queenService,proto3" json:"queen_service,omitempty"`
	// Report the DUT utilization or not.
	EnableInventoryReporting bool `protobuf:"varint,13,opt,name=enable_inventory_reporting,json=enableInventoryReporting,proto3" json:"enable_inventory_reporting,omitempty"`
	// HaRT PubSub Configs
	Hart                 *HaRT    `protobuf:"bytes,14,opt,name=hart,proto3" json:"hart,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe4d2048e3798022, []int{1}
}

func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (m *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(m, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetAdminService() *AdminService {
	if m != nil {
		return m.AdminService
	}
	return nil
}

func (m *Config) GetReaders() *LuciAuthGroup {
	if m != nil {
		return m.Readers
	}
	return nil
}

func (m *Config) GetStatusWriters() *LuciAuthGroup {
	if m != nil {
		return m.StatusWriters
	}
	return nil
}

func (m *Config) GetSetupWriters() *LuciAuthGroup {
	if m != nil {
		return m.SetupWriters
	}
	return nil
}

func (m *Config) GetPrivilegedWriters() *LuciAuthGroup {
	if m != nil {
		return m.PrivilegedWriters
	}
	return nil
}

func (m *Config) GetHwidSecret() string {
	if m != nil {
		return m.HwidSecret
	}
	return ""
}

func (m *Config) GetDeviceConfigSource() *Gitiles {
	if m != nil {
		return m.DeviceConfigSource
	}
	return nil
}

func (m *Config) GetManufacturingConfigSource() *Gitiles {
	if m != nil {
		return m.ManufacturingConfigSource
	}
	return nil
}

func (m *Config) GetInventory() *InventoryV1Repo {
	if m != nil {
		return m.Inventory
	}
	return nil
}

func (m *Config) GetEnvironment() string {
	if m != nil {
		return m.Environment
	}
	return ""
}

func (m *Config) GetQueenService() string {
	if m != nil {
		return m.QueenService
	}
	return ""
}

func (m *Config) GetEnableInventoryReporting() bool {
	if m != nil {
		return m.EnableInventoryReporting
	}
	return false
}

func (m *Config) GetHart() *HaRT {
	if m != nil {
		return m.Hart
	}
	return nil
}

type AdminService struct {
	// The skylab admin GAE server hosting the admin services.
	Host                 string   `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AdminService) Reset()         { *m = AdminService{} }
func (m *AdminService) String() string { return proto.CompactTextString(m) }
func (*AdminService) ProtoMessage()    {}
func (*AdminService) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe4d2048e3798022, []int{2}
}

func (m *AdminService) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AdminService.Unmarshal(m, b)
}
func (m *AdminService) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AdminService.Marshal(b, m, deterministic)
}
func (m *AdminService) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AdminService.Merge(m, src)
}
func (m *AdminService) XXX_Size() int {
	return xxx_messageInfo_AdminService.Size(m)
}
func (m *AdminService) XXX_DiscardUnknown() {
	xxx_messageInfo_AdminService.DiscardUnknown(m)
}

var xxx_messageInfo_AdminService proto.InternalMessageInfo

func (m *AdminService) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

type Gitiles struct {
	// The gitiles host name, e.g. 'chrome-internal.googlesource.com'.
	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	// The project (repo) name, e.g. 'chromeos/infra/config'.
	Project string `protobuf:"bytes,2,opt,name=project,proto3" json:"project,omitempty"`
	// The commit hash/branch to be checked out, e.g. 'refs/heads/master'.
	Committish string `protobuf:"bytes,3,opt,name=committish,proto3" json:"committish,omitempty"`
	// The path of the file to be downloaded, e.g. 'path/to/file.cfg'.
	Path                 string   `protobuf:"bytes,4,opt,name=path,proto3" json:"path,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Gitiles) Reset()         { *m = Gitiles{} }
func (m *Gitiles) String() string { return proto.CompactTextString(m) }
func (*Gitiles) ProtoMessage()    {}
func (*Gitiles) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe4d2048e3798022, []int{3}
}

func (m *Gitiles) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Gitiles.Unmarshal(m, b)
}
func (m *Gitiles) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Gitiles.Marshal(b, m, deterministic)
}
func (m *Gitiles) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Gitiles.Merge(m, src)
}
func (m *Gitiles) XXX_Size() int {
	return xxx_messageInfo_Gitiles.Size(m)
}
func (m *Gitiles) XXX_DiscardUnknown() {
	xxx_messageInfo_Gitiles.DiscardUnknown(m)
}

var xxx_messageInfo_Gitiles proto.InternalMessageInfo

func (m *Gitiles) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *Gitiles) GetProject() string {
	if m != nil {
		return m.Project
	}
	return ""
}

func (m *Gitiles) GetCommittish() string {
	if m != nil {
		return m.Committish
	}
	return ""
}

func (m *Gitiles) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

type InventoryV1Repo struct {
	Host                   string   `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Project                string   `protobuf:"bytes,2,opt,name=project,proto3" json:"project,omitempty"`
	Branch                 string   `protobuf:"bytes,3,opt,name=branch,proto3" json:"branch,omitempty"`
	LabDataPath            string   `protobuf:"bytes,4,opt,name=lab_data_path,json=labDataPath,proto3" json:"lab_data_path,omitempty"`
	InfrastructureDataPath string   `protobuf:"bytes,5,opt,name=infrastructure_data_path,json=infrastructureDataPath,proto3" json:"infrastructure_data_path,omitempty"`
	Multifile              bool     `protobuf:"varint,6,opt,name=multifile,proto3" json:"multifile,omitempty"`
	XXX_NoUnkeyedLiteral   struct{} `json:"-"`
	XXX_unrecognized       []byte   `json:"-"`
	XXX_sizecache          int32    `json:"-"`
}

func (m *InventoryV1Repo) Reset()         { *m = InventoryV1Repo{} }
func (m *InventoryV1Repo) String() string { return proto.CompactTextString(m) }
func (*InventoryV1Repo) ProtoMessage()    {}
func (*InventoryV1Repo) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe4d2048e3798022, []int{4}
}

func (m *InventoryV1Repo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InventoryV1Repo.Unmarshal(m, b)
}
func (m *InventoryV1Repo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InventoryV1Repo.Marshal(b, m, deterministic)
}
func (m *InventoryV1Repo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InventoryV1Repo.Merge(m, src)
}
func (m *InventoryV1Repo) XXX_Size() int {
	return xxx_messageInfo_InventoryV1Repo.Size(m)
}
func (m *InventoryV1Repo) XXX_DiscardUnknown() {
	xxx_messageInfo_InventoryV1Repo.DiscardUnknown(m)
}

var xxx_messageInfo_InventoryV1Repo proto.InternalMessageInfo

func (m *InventoryV1Repo) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *InventoryV1Repo) GetProject() string {
	if m != nil {
		return m.Project
	}
	return ""
}

func (m *InventoryV1Repo) GetBranch() string {
	if m != nil {
		return m.Branch
	}
	return ""
}

func (m *InventoryV1Repo) GetLabDataPath() string {
	if m != nil {
		return m.LabDataPath
	}
	return ""
}

func (m *InventoryV1Repo) GetInfrastructureDataPath() string {
	if m != nil {
		return m.InfrastructureDataPath
	}
	return ""
}

func (m *InventoryV1Repo) GetMultifile() bool {
	if m != nil {
		return m.Multifile
	}
	return false
}

type HaRT struct {
	Project              string   `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	Topic                string   `protobuf:"bytes,2,opt,name=topic,proto3" json:"topic,omitempty"`
	Subscription         string   `protobuf:"bytes,3,opt,name=subscription,proto3" json:"subscription,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HaRT) Reset()         { *m = HaRT{} }
func (m *HaRT) String() string { return proto.CompactTextString(m) }
func (*HaRT) ProtoMessage()    {}
func (*HaRT) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe4d2048e3798022, []int{5}
}

func (m *HaRT) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HaRT.Unmarshal(m, b)
}
func (m *HaRT) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HaRT.Marshal(b, m, deterministic)
}
func (m *HaRT) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HaRT.Merge(m, src)
}
func (m *HaRT) XXX_Size() int {
	return xxx_messageInfo_HaRT.Size(m)
}
func (m *HaRT) XXX_DiscardUnknown() {
	xxx_messageInfo_HaRT.DiscardUnknown(m)
}

var xxx_messageInfo_HaRT proto.InternalMessageInfo

func (m *HaRT) GetProject() string {
	if m != nil {
		return m.Project
	}
	return ""
}

func (m *HaRT) GetTopic() string {
	if m != nil {
		return m.Topic
	}
	return ""
}

func (m *HaRT) GetSubscription() string {
	if m != nil {
		return m.Subscription
	}
	return ""
}

func init() {
	proto.RegisterType((*LuciAuthGroup)(nil), "lab_inventory.config.LuciAuthGroup")
	proto.RegisterType((*Config)(nil), "lab_inventory.config.Config")
	proto.RegisterType((*AdminService)(nil), "lab_inventory.config.AdminService")
	proto.RegisterType((*Gitiles)(nil), "lab_inventory.config.Gitiles")
	proto.RegisterType((*InventoryV1Repo)(nil), "lab_inventory.config.InventoryV1Repo")
	proto.RegisterType((*HaRT)(nil), "lab_inventory.config.HaRT")
}

func init() {
	proto.RegisterFile("infra/appengine/cros/lab_inventory/app/config/config.proto", fileDescriptor_fe4d2048e3798022)
}

var fileDescriptor_fe4d2048e3798022 = []byte{
	// 628 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x94, 0x4b, 0x6b, 0xdb, 0x4a,
	0x14, 0xc7, 0x51, 0xae, 0x5f, 0x3a, 0xb6, 0x72, 0xef, 0x1d, 0x4c, 0xd0, 0xcd, 0xed, 0xc3, 0x28,
	0x04, 0xb2, 0xb2, 0x69, 0xbb, 0x29, 0xa5, 0x5d, 0xa4, 0x29, 0x24, 0x0d, 0x85, 0x96, 0x49, 0x69,
	0x21, 0x50, 0xc4, 0x58, 0x3e, 0xb6, 0xa6, 0x95, 0x67, 0xd4, 0x79, 0x38, 0xf4, 0x33, 0x76, 0xd9,
	0x2f, 0x54, 0x34, 0xb2, 0x6c, 0x29, 0x78, 0xe1, 0xae, 0xa4, 0xf3, 0xf8, 0xff, 0xce, 0xcc, 0x99,
	0x39, 0x03, 0x2f, 0xb8, 0x98, 0x2b, 0x36, 0x61, 0x79, 0x8e, 0x62, 0xc1, 0x05, 0x4e, 0x12, 0x25,
	0xf5, 0x24, 0x63, 0xd3, 0x98, 0x8b, 0x15, 0x0a, 0x23, 0xd5, 0x8f, 0x22, 0x38, 0x49, 0xa4, 0x98,
	0xf3, 0xc5, 0xfa, 0x33, 0xce, 0x95, 0x34, 0x92, 0x0c, 0x1b, 0x69, 0xe3, 0x32, 0x16, 0x9d, 0x42,
	0xf0, 0xce, 0x26, 0xfc, 0xdc, 0x9a, 0xf4, 0x52, 0x49, 0x9b, 0x93, 0x21, 0xb4, 0x57, 0x2c, 0xb3,
	0x18, 0x7a, 0x23, 0xef, 0xcc, 0xa7, 0xa5, 0x11, 0xfd, 0xec, 0x40, 0xe7, 0xc2, 0x29, 0xc8, 0x25,
	0x04, 0x6c, 0xb6, 0xe4, 0x22, 0xd6, 0xa8, 0x56, 0x3c, 0xc1, 0xf0, 0x60, 0xe4, 0x9d, 0xf5, 0x9f,
	0x46, 0xe3, 0x5d, 0xfc, 0xf1, 0x79, 0x91, 0x7a, 0x53, 0x66, 0xd2, 0x01, 0xab, 0x59, 0xe4, 0x15,
	0x74, 0x15, 0xb2, 0x19, 0x2a, 0x1d, 0xfe, 0xe5, 0x10, 0x27, 0xbb, 0x11, 0x8d, 0xf5, 0xd1, 0x4a,
	0x43, 0xae, 0xe1, 0x50, 0x1b, 0x66, 0xac, 0x8e, 0xef, 0x14, 0x37, 0x05, 0xa5, 0xb5, 0x3f, 0x25,
	0x28, 0xa5, 0x9f, 0x4b, 0x25, 0xb9, 0x82, 0x40, 0xa3, 0xb1, 0xf9, 0x06, 0xd5, 0xde, 0x1f, 0x35,
	0x70, 0xca, 0x8a, 0x44, 0x81, 0xe4, 0x8a, 0xaf, 0x78, 0x86, 0x0b, 0x9c, 0x6d, 0x70, 0x9d, 0xfd,
	0x71, 0xff, 0x6e, 0xe5, 0x15, 0xf3, 0x31, 0xf4, 0xd3, 0x3b, 0x3e, 0x8b, 0x35, 0x26, 0x0a, 0x4d,
	0xd8, 0x75, 0x07, 0x03, 0x85, 0xeb, 0xc6, 0x79, 0xc8, 0x7b, 0x18, 0xce, 0xb0, 0xe8, 0x69, 0x5c,
	0x22, 0x63, 0x2d, 0xad, 0x4a, 0x30, 0xec, 0xb9, 0xb2, 0x0f, 0x77, 0x97, 0xbd, 0xe4, 0x86, 0x67,
	0xa8, 0x29, 0x29, 0xa5, 0xe5, 0xe9, 0xde, 0x38, 0x21, 0xf9, 0x02, 0xff, 0x2f, 0x99, 0xb0, 0x73,
	0x96, 0x18, 0xab, 0xb8, 0x58, 0xdc, 0xe3, 0xfa, 0xfb, 0x70, 0xff, 0x6b, 0x10, 0x1a, 0xf8, 0x0b,
	0xf0, 0x37, 0xb2, 0x70, 0xe0, 0x60, 0xa7, 0xbb, 0x61, 0x6f, 0x2b, 0xc7, 0xa7, 0x27, 0x14, 0x73,
	0x49, 0xb7, 0x3a, 0x32, 0x82, 0x3e, 0x8a, 0x15, 0x57, 0x52, 0x2c, 0x51, 0x98, 0x10, 0x5c, 0x57,
	0xea, 0x2e, 0x72, 0x02, 0xc1, 0x77, 0x8b, 0xb8, 0xbd, 0xa9, 0x7d, 0x97, 0x33, 0x70, 0xce, 0xea,
	0x16, 0xbe, 0x84, 0x63, 0x14, 0x6c, 0x9a, 0xe1, 0xb6, 0x78, 0xac, 0x30, 0x97, 0xca, 0x70, 0xb1,
	0x08, 0x83, 0x91, 0x77, 0xd6, 0xa3, 0x61, 0x99, 0xb1, 0x59, 0x0c, 0xad, 0xe2, 0x64, 0x0c, 0xad,
	0x94, 0x29, 0x13, 0x1e, 0xba, 0x4d, 0x1c, 0xef, 0xde, 0xc4, 0x15, 0xa3, 0x1f, 0xa9, 0xcb, 0xbb,
	0x6e, 0xf5, 0xbc, 0x7f, 0x0e, 0xa2, 0x08, 0x06, 0xf5, 0xb9, 0x20, 0x04, 0x5a, 0xa9, 0xd4, 0x66,
	0x3d, 0x72, 0xee, 0x3f, 0xfa, 0x06, 0xdd, 0x75, 0x27, 0x77, 0x85, 0x49, 0x08, 0xdd, 0x5c, 0xc9,
	0xaf, 0x98, 0x18, 0x37, 0x7f, 0x3e, 0xad, 0x4c, 0xf2, 0x08, 0x20, 0x91, 0xcb, 0x25, 0x37, 0x86,
	0xeb, 0xd4, 0x4d, 0x96, 0x4f, 0x6b, 0x9e, 0x82, 0x96, 0x33, 0x93, 0xba, 0x69, 0xf1, 0xa9, 0xfb,
	0x8f, 0x7e, 0x79, 0xf0, 0xf7, 0xbd, 0x56, 0xff, 0x61, 0xd5, 0x23, 0xe8, 0x4c, 0x15, 0x13, 0x49,
	0x55, 0x71, 0x6d, 0x91, 0x08, 0x82, 0xa2, 0x27, 0x33, 0x66, 0x58, 0x5c, 0x2b, 0xdb, 0xcf, 0xd8,
	0xf4, 0x0d, 0x33, 0xec, 0x03, 0x33, 0x29, 0x79, 0x0e, 0xa1, 0x7b, 0xd7, 0xb4, 0x51, 0xb6, 0xb8,
	0x2e, 0x58, 0x4b, 0x6f, 0xbb, 0xf4, 0xa3, 0x66, 0x7c, 0xa3, 0x7c, 0x00, 0xfe, 0xd2, 0x66, 0x86,
	0xcf, 0x79, 0x86, 0x6e, 0xc8, 0x7a, 0x74, 0xeb, 0x88, 0x6e, 0xa1, 0x55, 0xb4, 0xbe, 0xbe, 0x6a,
	0xaf, 0xb9, 0xea, 0x21, 0xb4, 0x8d, 0xcc, 0x79, 0xb2, 0xde, 0x4d, 0x69, 0x90, 0x08, 0x06, 0xda,
	0x4e, 0x75, 0xa2, 0x78, 0x6e, 0xb8, 0x14, 0xeb, 0x1d, 0x35, 0x7c, 0xaf, 0x7b, 0xb7, 0x9d, 0xf2,
	0x74, 0xa7, 0x1d, 0xf7, 0xbc, 0x3e, 0xfb, 0x1d, 0x00, 0x00, 0xff, 0xff, 0xd6, 0xa5, 0xe1, 0xbc,
	0x9c, 0x05, 0x00, 0x00,
}
