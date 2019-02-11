// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/api_proto/project_objects.proto

package monorail

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

// Next available tag: 4
type Project struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Summary              string   `protobuf:"bytes,2,opt,name=summary,proto3" json:"summary,omitempty"`
	Description          string   `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Project) Reset()         { *m = Project{} }
func (m *Project) String() string { return proto.CompactTextString(m) }
func (*Project) ProtoMessage()    {}
func (*Project) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f680a8ed8804f88, []int{0}
}

func (m *Project) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Project.Unmarshal(m, b)
}
func (m *Project) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Project.Marshal(b, m, deterministic)
}
func (m *Project) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Project.Merge(m, src)
}
func (m *Project) XXX_Size() int {
	return xxx_messageInfo_Project.Size(m)
}
func (m *Project) XXX_DiscardUnknown() {
	xxx_messageInfo_Project.DiscardUnknown(m)
}

var xxx_messageInfo_Project proto.InternalMessageInfo

func (m *Project) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Project) GetSummary() string {
	if m != nil {
		return m.Summary
	}
	return ""
}

func (m *Project) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

// Next available tag: 6
type StatusDef struct {
	Status               string   `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	MeansOpen            bool     `protobuf:"varint,2,opt,name=means_open,json=meansOpen,proto3" json:"means_open,omitempty"`
	Rank                 uint32   `protobuf:"varint,3,opt,name=rank,proto3" json:"rank,omitempty"`
	Docstring            string   `protobuf:"bytes,4,opt,name=docstring,proto3" json:"docstring,omitempty"`
	Deprecated           bool     `protobuf:"varint,5,opt,name=deprecated,proto3" json:"deprecated,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StatusDef) Reset()         { *m = StatusDef{} }
func (m *StatusDef) String() string { return proto.CompactTextString(m) }
func (*StatusDef) ProtoMessage()    {}
func (*StatusDef) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f680a8ed8804f88, []int{1}
}

func (m *StatusDef) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatusDef.Unmarshal(m, b)
}
func (m *StatusDef) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatusDef.Marshal(b, m, deterministic)
}
func (m *StatusDef) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatusDef.Merge(m, src)
}
func (m *StatusDef) XXX_Size() int {
	return xxx_messageInfo_StatusDef.Size(m)
}
func (m *StatusDef) XXX_DiscardUnknown() {
	xxx_messageInfo_StatusDef.DiscardUnknown(m)
}

var xxx_messageInfo_StatusDef proto.InternalMessageInfo

func (m *StatusDef) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *StatusDef) GetMeansOpen() bool {
	if m != nil {
		return m.MeansOpen
	}
	return false
}

func (m *StatusDef) GetRank() uint32 {
	if m != nil {
		return m.Rank
	}
	return 0
}

func (m *StatusDef) GetDocstring() string {
	if m != nil {
		return m.Docstring
	}
	return ""
}

func (m *StatusDef) GetDeprecated() bool {
	if m != nil {
		return m.Deprecated
	}
	return false
}

// Next available tag: 5
type LabelDef struct {
	Label                string   `protobuf:"bytes,1,opt,name=label,proto3" json:"label,omitempty"`
	Docstring            string   `protobuf:"bytes,3,opt,name=docstring,proto3" json:"docstring,omitempty"`
	Deprecated           bool     `protobuf:"varint,4,opt,name=deprecated,proto3" json:"deprecated,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LabelDef) Reset()         { *m = LabelDef{} }
func (m *LabelDef) String() string { return proto.CompactTextString(m) }
func (*LabelDef) ProtoMessage()    {}
func (*LabelDef) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f680a8ed8804f88, []int{2}
}

func (m *LabelDef) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LabelDef.Unmarshal(m, b)
}
func (m *LabelDef) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LabelDef.Marshal(b, m, deterministic)
}
func (m *LabelDef) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LabelDef.Merge(m, src)
}
func (m *LabelDef) XXX_Size() int {
	return xxx_messageInfo_LabelDef.Size(m)
}
func (m *LabelDef) XXX_DiscardUnknown() {
	xxx_messageInfo_LabelDef.DiscardUnknown(m)
}

var xxx_messageInfo_LabelDef proto.InternalMessageInfo

func (m *LabelDef) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *LabelDef) GetDocstring() string {
	if m != nil {
		return m.Docstring
	}
	return ""
}

func (m *LabelDef) GetDeprecated() bool {
	if m != nil {
		return m.Deprecated
	}
	return false
}

// Next available tag: 11
type ComponentDef struct {
	Path                 string      `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	Docstring            string      `protobuf:"bytes,2,opt,name=docstring,proto3" json:"docstring,omitempty"`
	AdminRefs            []*UserRef  `protobuf:"bytes,3,rep,name=admin_refs,json=adminRefs,proto3" json:"admin_refs,omitempty"`
	CcRefs               []*UserRef  `protobuf:"bytes,4,rep,name=cc_refs,json=ccRefs,proto3" json:"cc_refs,omitempty"`
	Deprecated           bool        `protobuf:"varint,5,opt,name=deprecated,proto3" json:"deprecated,omitempty"`
	Created              uint32      `protobuf:"fixed32,6,opt,name=created,proto3" json:"created,omitempty"`
	CreatorRef           *UserRef    `protobuf:"bytes,7,opt,name=creator_ref,json=creatorRef,proto3" json:"creator_ref,omitempty"`
	Modified             uint32      `protobuf:"fixed32,8,opt,name=modified,proto3" json:"modified,omitempty"`
	ModifierRef          *UserRef    `protobuf:"bytes,9,opt,name=modifier_ref,json=modifierRef,proto3" json:"modifier_ref,omitempty"`
	LabelRefs            []*LabelRef `protobuf:"bytes,10,rep,name=label_refs,json=labelRefs,proto3" json:"label_refs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *ComponentDef) Reset()         { *m = ComponentDef{} }
func (m *ComponentDef) String() string { return proto.CompactTextString(m) }
func (*ComponentDef) ProtoMessage()    {}
func (*ComponentDef) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f680a8ed8804f88, []int{3}
}

func (m *ComponentDef) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ComponentDef.Unmarshal(m, b)
}
func (m *ComponentDef) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ComponentDef.Marshal(b, m, deterministic)
}
func (m *ComponentDef) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ComponentDef.Merge(m, src)
}
func (m *ComponentDef) XXX_Size() int {
	return xxx_messageInfo_ComponentDef.Size(m)
}
func (m *ComponentDef) XXX_DiscardUnknown() {
	xxx_messageInfo_ComponentDef.DiscardUnknown(m)
}

var xxx_messageInfo_ComponentDef proto.InternalMessageInfo

func (m *ComponentDef) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *ComponentDef) GetDocstring() string {
	if m != nil {
		return m.Docstring
	}
	return ""
}

func (m *ComponentDef) GetAdminRefs() []*UserRef {
	if m != nil {
		return m.AdminRefs
	}
	return nil
}

func (m *ComponentDef) GetCcRefs() []*UserRef {
	if m != nil {
		return m.CcRefs
	}
	return nil
}

func (m *ComponentDef) GetDeprecated() bool {
	if m != nil {
		return m.Deprecated
	}
	return false
}

func (m *ComponentDef) GetCreated() uint32 {
	if m != nil {
		return m.Created
	}
	return 0
}

func (m *ComponentDef) GetCreatorRef() *UserRef {
	if m != nil {
		return m.CreatorRef
	}
	return nil
}

func (m *ComponentDef) GetModified() uint32 {
	if m != nil {
		return m.Modified
	}
	return 0
}

func (m *ComponentDef) GetModifierRef() *UserRef {
	if m != nil {
		return m.ModifierRef
	}
	return nil
}

func (m *ComponentDef) GetLabelRefs() []*LabelRef {
	if m != nil {
		return m.LabelRefs
	}
	return nil
}

// Next available tag: 9
type FieldDef struct {
	FieldRef       *FieldRef `protobuf:"bytes,1,opt,name=field_ref,json=fieldRef,proto3" json:"field_ref,omitempty"`
	ApplicableType string    `protobuf:"bytes,2,opt,name=applicable_type,json=applicableType,proto3" json:"applicable_type,omitempty"`
	// TODO(jrobbins): applicable_predicate
	IsRequired    bool       `protobuf:"varint,3,opt,name=is_required,json=isRequired,proto3" json:"is_required,omitempty"`
	IsNiche       bool       `protobuf:"varint,4,opt,name=is_niche,json=isNiche,proto3" json:"is_niche,omitempty"`
	IsMultivalued bool       `protobuf:"varint,5,opt,name=is_multivalued,json=isMultivalued,proto3" json:"is_multivalued,omitempty"`
	Docstring     string     `protobuf:"bytes,6,opt,name=docstring,proto3" json:"docstring,omitempty"`
	AdminRefs     []*UserRef `protobuf:"bytes,7,rep,name=admin_refs,json=adminRefs,proto3" json:"admin_refs,omitempty"`
	// TODO(jrobbins): validation, permission granting, and notification options.
	IsPhaseField         bool        `protobuf:"varint,8,opt,name=is_phase_field,json=isPhaseField,proto3" json:"is_phase_field,omitempty"`
	UserChoices          []*UserRef  `protobuf:"bytes,9,rep,name=user_choices,json=userChoices,proto3" json:"user_choices,omitempty"`
	EnumChoices          []*LabelDef `protobuf:"bytes,10,rep,name=enum_choices,json=enumChoices,proto3" json:"enum_choices,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *FieldDef) Reset()         { *m = FieldDef{} }
func (m *FieldDef) String() string { return proto.CompactTextString(m) }
func (*FieldDef) ProtoMessage()    {}
func (*FieldDef) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f680a8ed8804f88, []int{4}
}

func (m *FieldDef) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FieldDef.Unmarshal(m, b)
}
func (m *FieldDef) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FieldDef.Marshal(b, m, deterministic)
}
func (m *FieldDef) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FieldDef.Merge(m, src)
}
func (m *FieldDef) XXX_Size() int {
	return xxx_messageInfo_FieldDef.Size(m)
}
func (m *FieldDef) XXX_DiscardUnknown() {
	xxx_messageInfo_FieldDef.DiscardUnknown(m)
}

var xxx_messageInfo_FieldDef proto.InternalMessageInfo

func (m *FieldDef) GetFieldRef() *FieldRef {
	if m != nil {
		return m.FieldRef
	}
	return nil
}

func (m *FieldDef) GetApplicableType() string {
	if m != nil {
		return m.ApplicableType
	}
	return ""
}

func (m *FieldDef) GetIsRequired() bool {
	if m != nil {
		return m.IsRequired
	}
	return false
}

func (m *FieldDef) GetIsNiche() bool {
	if m != nil {
		return m.IsNiche
	}
	return false
}

func (m *FieldDef) GetIsMultivalued() bool {
	if m != nil {
		return m.IsMultivalued
	}
	return false
}

func (m *FieldDef) GetDocstring() string {
	if m != nil {
		return m.Docstring
	}
	return ""
}

func (m *FieldDef) GetAdminRefs() []*UserRef {
	if m != nil {
		return m.AdminRefs
	}
	return nil
}

func (m *FieldDef) GetIsPhaseField() bool {
	if m != nil {
		return m.IsPhaseField
	}
	return false
}

func (m *FieldDef) GetUserChoices() []*UserRef {
	if m != nil {
		return m.UserChoices
	}
	return nil
}

func (m *FieldDef) GetEnumChoices() []*LabelDef {
	if m != nil {
		return m.EnumChoices
	}
	return nil
}

// Next available tag: 3
type FieldOptions struct {
	FieldRef             *FieldRef  `protobuf:"bytes,1,opt,name=field_ref,json=fieldRef,proto3" json:"field_ref,omitempty"`
	UserRefs             []*UserRef `protobuf:"bytes,2,rep,name=user_refs,json=userRefs,proto3" json:"user_refs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *FieldOptions) Reset()         { *m = FieldOptions{} }
func (m *FieldOptions) String() string { return proto.CompactTextString(m) }
func (*FieldOptions) ProtoMessage()    {}
func (*FieldOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f680a8ed8804f88, []int{5}
}

func (m *FieldOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FieldOptions.Unmarshal(m, b)
}
func (m *FieldOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FieldOptions.Marshal(b, m, deterministic)
}
func (m *FieldOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FieldOptions.Merge(m, src)
}
func (m *FieldOptions) XXX_Size() int {
	return xxx_messageInfo_FieldOptions.Size(m)
}
func (m *FieldOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_FieldOptions.DiscardUnknown(m)
}

var xxx_messageInfo_FieldOptions proto.InternalMessageInfo

func (m *FieldOptions) GetFieldRef() *FieldRef {
	if m != nil {
		return m.FieldRef
	}
	return nil
}

func (m *FieldOptions) GetUserRefs() []*UserRef {
	if m != nil {
		return m.UserRefs
	}
	return nil
}

// Next available tag: 4
type ApprovalDef struct {
	FieldRef             *FieldRef  `protobuf:"bytes,1,opt,name=field_ref,json=fieldRef,proto3" json:"field_ref,omitempty"`
	ApproverRefs         []*UserRef `protobuf:"bytes,2,rep,name=approver_refs,json=approverRefs,proto3" json:"approver_refs,omitempty"`
	Survey               string     `protobuf:"bytes,3,opt,name=survey,proto3" json:"survey,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *ApprovalDef) Reset()         { *m = ApprovalDef{} }
func (m *ApprovalDef) String() string { return proto.CompactTextString(m) }
func (*ApprovalDef) ProtoMessage()    {}
func (*ApprovalDef) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f680a8ed8804f88, []int{6}
}

func (m *ApprovalDef) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ApprovalDef.Unmarshal(m, b)
}
func (m *ApprovalDef) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ApprovalDef.Marshal(b, m, deterministic)
}
func (m *ApprovalDef) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ApprovalDef.Merge(m, src)
}
func (m *ApprovalDef) XXX_Size() int {
	return xxx_messageInfo_ApprovalDef.Size(m)
}
func (m *ApprovalDef) XXX_DiscardUnknown() {
	xxx_messageInfo_ApprovalDef.DiscardUnknown(m)
}

var xxx_messageInfo_ApprovalDef proto.InternalMessageInfo

func (m *ApprovalDef) GetFieldRef() *FieldRef {
	if m != nil {
		return m.FieldRef
	}
	return nil
}

func (m *ApprovalDef) GetApproverRefs() []*UserRef {
	if m != nil {
		return m.ApproverRefs
	}
	return nil
}

func (m *ApprovalDef) GetSurvey() string {
	if m != nil {
		return m.Survey
	}
	return ""
}

// Next available tag: 11
type Config struct {
	ProjectName            string          `protobuf:"bytes,1,opt,name=project_name,json=projectName,proto3" json:"project_name,omitempty"`
	StatusDefs             []*StatusDef    `protobuf:"bytes,2,rep,name=status_defs,json=statusDefs,proto3" json:"status_defs,omitempty"`
	StatusesOfferMerge     []*StatusRef    `protobuf:"bytes,3,rep,name=statuses_offer_merge,json=statusesOfferMerge,proto3" json:"statuses_offer_merge,omitempty"`
	LabelDefs              []*LabelDef     `protobuf:"bytes,4,rep,name=label_defs,json=labelDefs,proto3" json:"label_defs,omitempty"`
	ExclusiveLabelPrefixes []string        `protobuf:"bytes,5,rep,name=exclusive_label_prefixes,json=exclusiveLabelPrefixes,proto3" json:"exclusive_label_prefixes,omitempty"`
	ComponentDefs          []*ComponentDef `protobuf:"bytes,6,rep,name=component_defs,json=componentDefs,proto3" json:"component_defs,omitempty"`
	FieldDefs              []*FieldDef     `protobuf:"bytes,7,rep,name=field_defs,json=fieldDefs,proto3" json:"field_defs,omitempty"`
	ApprovalDefs           []*ApprovalDef  `protobuf:"bytes,8,rep,name=approval_defs,json=approvalDefs,proto3" json:"approval_defs,omitempty"`
	RestrictToKnown        bool            `protobuf:"varint,9,opt,name=restrict_to_known,json=restrictToKnown,proto3" json:"restrict_to_known,omitempty"`
	XXX_NoUnkeyedLiteral   struct{}        `json:"-"`
	XXX_unrecognized       []byte          `json:"-"`
	XXX_sizecache          int32           `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f680a8ed8804f88, []int{7}
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

func (m *Config) GetProjectName() string {
	if m != nil {
		return m.ProjectName
	}
	return ""
}

func (m *Config) GetStatusDefs() []*StatusDef {
	if m != nil {
		return m.StatusDefs
	}
	return nil
}

func (m *Config) GetStatusesOfferMerge() []*StatusRef {
	if m != nil {
		return m.StatusesOfferMerge
	}
	return nil
}

func (m *Config) GetLabelDefs() []*LabelDef {
	if m != nil {
		return m.LabelDefs
	}
	return nil
}

func (m *Config) GetExclusiveLabelPrefixes() []string {
	if m != nil {
		return m.ExclusiveLabelPrefixes
	}
	return nil
}

func (m *Config) GetComponentDefs() []*ComponentDef {
	if m != nil {
		return m.ComponentDefs
	}
	return nil
}

func (m *Config) GetFieldDefs() []*FieldDef {
	if m != nil {
		return m.FieldDefs
	}
	return nil
}

func (m *Config) GetApprovalDefs() []*ApprovalDef {
	if m != nil {
		return m.ApprovalDefs
	}
	return nil
}

func (m *Config) GetRestrictToKnown() bool {
	if m != nil {
		return m.RestrictToKnown
	}
	return false
}

// Next available tag: 2
type TemplateDef struct {
	TemplateName         string   `protobuf:"bytes,1,opt,name=template_name,json=templateName,proto3" json:"template_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TemplateDef) Reset()         { *m = TemplateDef{} }
func (m *TemplateDef) String() string { return proto.CompactTextString(m) }
func (*TemplateDef) ProtoMessage()    {}
func (*TemplateDef) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f680a8ed8804f88, []int{8}
}

func (m *TemplateDef) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TemplateDef.Unmarshal(m, b)
}
func (m *TemplateDef) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TemplateDef.Marshal(b, m, deterministic)
}
func (m *TemplateDef) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TemplateDef.Merge(m, src)
}
func (m *TemplateDef) XXX_Size() int {
	return xxx_messageInfo_TemplateDef.Size(m)
}
func (m *TemplateDef) XXX_DiscardUnknown() {
	xxx_messageInfo_TemplateDef.DiscardUnknown(m)
}

var xxx_messageInfo_TemplateDef proto.InternalMessageInfo

func (m *TemplateDef) GetTemplateName() string {
	if m != nil {
		return m.TemplateName
	}
	return ""
}

func init() {
	proto.RegisterType((*Project)(nil), "monorail.Project")
	proto.RegisterType((*StatusDef)(nil), "monorail.StatusDef")
	proto.RegisterType((*LabelDef)(nil), "monorail.LabelDef")
	proto.RegisterType((*ComponentDef)(nil), "monorail.ComponentDef")
	proto.RegisterType((*FieldDef)(nil), "monorail.FieldDef")
	proto.RegisterType((*FieldOptions)(nil), "monorail.FieldOptions")
	proto.RegisterType((*ApprovalDef)(nil), "monorail.ApprovalDef")
	proto.RegisterType((*Config)(nil), "monorail.Config")
	proto.RegisterType((*TemplateDef)(nil), "monorail.TemplateDef")
}

func init() {
	proto.RegisterFile("api/api_proto/project_objects.proto", fileDescriptor_4f680a8ed8804f88)
}

var fileDescriptor_4f680a8ed8804f88 = []byte{
	// 873 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x55, 0x5f, 0x6f, 0xdc, 0x44,
	0x10, 0xd7, 0xf5, 0x2e, 0x77, 0xf6, 0xf8, 0x2e, 0x55, 0x97, 0x12, 0x99, 0x88, 0x3f, 0xc7, 0x15,
	0x44, 0xd4, 0x87, 0x04, 0x42, 0x41, 0x08, 0x89, 0x07, 0x94, 0xc2, 0x0b, 0xb4, 0x89, 0x4c, 0x78,
	0xe0, 0x05, 0x6b, 0xb3, 0x1e, 0x27, 0x4b, 0xed, 0xdd, 0x65, 0xd7, 0x0e, 0xcd, 0x97, 0x40, 0x42,
	0xe2, 0x53, 0xf0, 0xe5, 0xf8, 0x0a, 0x68, 0xc7, 0xeb, 0xf3, 0xa5, 0x4d, 0x5a, 0xe0, 0xe9, 0x66,
	0x66, 0x7f, 0x33, 0xbf, 0xd9, 0xfd, 0x8d, 0xe7, 0xe0, 0x01, 0x37, 0xf2, 0x80, 0x1b, 0x99, 0x1b,
	0xab, 0x1b, 0x7d, 0x60, 0xac, 0xfe, 0x05, 0x45, 0x93, 0xeb, 0x33, 0xff, 0xe3, 0xf6, 0x29, 0xca,
	0xa2, 0x5a, 0x2b, 0x6d, 0xb9, 0xac, 0x76, 0x77, 0xaf, 0xc3, 0x85, 0xae, 0x6b, 0xad, 0x3a, 0xd4,
	0xea, 0x27, 0x98, 0x9d, 0x74, 0xe9, 0x8c, 0xc1, 0x44, 0xf1, 0x1a, 0xd3, 0xd1, 0x72, 0xb4, 0x17,
	0x67, 0x64, 0xb3, 0x14, 0x66, 0xae, 0xad, 0x6b, 0x6e, 0xaf, 0xd2, 0x3b, 0x14, 0xee, 0x5d, 0xb6,
	0x84, 0xa4, 0x40, 0x27, 0xac, 0x34, 0x8d, 0xd4, 0x2a, 0x1d, 0xd3, 0xe9, 0x66, 0x68, 0xf5, 0xe7,
	0x08, 0xe2, 0x1f, 0x1a, 0xde, 0xb4, 0xee, 0x31, 0x96, 0x6c, 0x07, 0xa6, 0x8e, 0x9c, 0x50, 0x3f,
	0x78, 0xec, 0x1d, 0x80, 0x1a, 0xb9, 0x72, 0xb9, 0x36, 0xa8, 0x88, 0x24, 0xca, 0x62, 0x8a, 0x1c,
	0x1b, 0x54, 0xbe, 0x29, 0xcb, 0xd5, 0x33, 0xaa, 0xbf, 0xc8, 0xc8, 0x66, 0x6f, 0x43, 0x5c, 0x68,
	0xe1, 0x1a, 0x2b, 0xd5, 0x79, 0x3a, 0xa1, 0x6a, 0x43, 0x80, 0xbd, 0x0b, 0x50, 0xa0, 0xb1, 0x28,
	0x78, 0x83, 0x45, 0xba, 0x45, 0x05, 0x37, 0x22, 0xab, 0x9f, 0x21, 0xfa, 0x9e, 0x9f, 0x61, 0xe5,
	0x9b, 0xba, 0x0f, 0x5b, 0x95, 0xb7, 0x43, 0x4f, 0x9d, 0x73, 0xbd, 0xfe, 0xf8, 0xd5, 0xf5, 0x27,
	0x2f, 0xd5, 0xff, 0x63, 0x0c, 0xf3, 0x23, 0x5d, 0x1b, 0xad, 0x50, 0x35, 0x9e, 0x84, 0xc1, 0xc4,
	0xf0, 0xe6, 0xa2, 0x7f, 0x57, 0x6f, 0x5f, 0xa7, 0xb8, 0xf3, 0x22, 0xc5, 0xc7, 0x00, 0xbc, 0xa8,
	0xa5, 0xca, 0x2d, 0x96, 0x2e, 0x1d, 0x2f, 0xc7, 0x7b, 0xc9, 0xe1, 0xbd, 0xfd, 0x5e, 0xcf, 0xfd,
	0x1f, 0x1d, 0xda, 0x0c, 0xcb, 0x2c, 0x26, 0x50, 0x86, 0xa5, 0x63, 0x0f, 0x61, 0x26, 0x44, 0x07,
	0x9f, 0xdc, 0x06, 0x9f, 0x0a, 0x41, 0xd8, 0xd7, 0x3c, 0x90, 0xd7, 0x5c, 0x58, 0xa4, 0xc3, 0xe9,
	0x72, 0xb4, 0x37, 0xcb, 0x7a, 0x97, 0x1d, 0x42, 0x42, 0xa6, 0xb6, 0x9e, 0x2a, 0x9d, 0x2d, 0x47,
	0x37, 0x33, 0x41, 0x40, 0x65, 0x58, 0xb2, 0x5d, 0x88, 0x6a, 0x5d, 0xc8, 0x52, 0x62, 0x91, 0x46,
	0x54, 0x6e, 0xed, 0xb3, 0x47, 0x30, 0x0f, 0x76, 0x57, 0x30, 0xbe, 0xad, 0x60, 0xd2, 0xc3, 0x7c,
	0xc5, 0x4f, 0x00, 0x48, 0xa7, 0xee, 0xba, 0x40, 0xd7, 0x65, 0x43, 0x0e, 0x89, 0x4b, 0xcf, 0x53,
	0x05, 0xcb, 0xad, 0xfe, 0x1a, 0x43, 0xf4, 0xad, 0xc4, 0xaa, 0xf0, 0x7a, 0x1c, 0x40, 0x5c, 0x7a,
	0x9b, 0x28, 0x47, 0x44, 0xb9, 0x91, 0x4e, 0x30, 0x9f, 0x1e, 0x95, 0xc1, 0x62, 0x1f, 0xc1, 0x5d,
	0x6e, 0x4c, 0x25, 0x05, 0x3f, 0xab, 0x30, 0x6f, 0xae, 0x0c, 0x06, 0xc9, 0xb6, 0x87, 0xf0, 0xe9,
	0x95, 0x41, 0xf6, 0x1e, 0x24, 0xd2, 0xe5, 0x16, 0x7f, 0x6d, 0xa5, 0xc5, 0x82, 0x46, 0x27, 0xca,
	0x40, 0xba, 0x2c, 0x44, 0xd8, 0x5b, 0x10, 0x49, 0x97, 0x2b, 0x29, 0x2e, 0x30, 0x4c, 0xce, 0x4c,
	0xba, 0xa7, 0xde, 0x65, 0x1f, 0xc2, 0xb6, 0x74, 0x79, 0xdd, 0x56, 0x8d, 0xbc, 0xe4, 0x55, 0xbb,
	0x56, 0x66, 0x21, 0xdd, 0x93, 0x21, 0x78, 0x7d, 0x70, 0xa6, 0xaf, 0x1e, 0x9c, 0xd9, 0xbf, 0x18,
	0x9c, 0x0f, 0x88, 0xd6, 0x5c, 0x70, 0x87, 0x39, 0x5d, 0x98, 0x44, 0x8a, 0xb2, 0xb9, 0x74, 0x27,
	0x3e, 0x48, 0xcf, 0xe1, 0x85, 0x6a, 0x1d, 0xda, 0x5c, 0x5c, 0x68, 0x29, 0xd0, 0xa5, 0xf1, 0x6d,
	0x95, 0x13, 0x0f, 0x3b, 0xea, 0x50, 0xec, 0x33, 0x98, 0xa3, 0x6a, 0xeb, 0x75, 0xd6, 0xcd, 0x52,
	0x3d, 0xf6, 0x69, 0x1e, 0x17, 0xd2, 0x56, 0x1a, 0xe6, 0xc4, 0x7a, 0x4c, 0x6b, 0xc4, 0xfd, 0x77,
	0xbd, 0xf6, 0x21, 0xa6, 0x6e, 0xe9, 0x11, 0xee, 0xdc, 0xd6, 0x6a, 0xd4, 0x76, 0x86, 0x5b, 0xfd,
	0x3e, 0x82, 0xe4, 0x6b, 0x63, 0xac, 0xbe, 0xe4, 0xd5, 0xff, 0x1a, 0x90, 0xcf, 0x61, 0xc1, 0x29,
	0xff, 0xb5, 0xa4, 0xf3, 0x1e, 0x47, 0x8f, 0xef, 0x77, 0x62, 0x6b, 0x2f, 0xf1, 0x2a, 0x6c, 0x99,
	0xe0, 0xad, 0xfe, 0x1e, 0xc3, 0xf4, 0x48, 0xab, 0x52, 0x9e, 0xb3, 0xf7, 0x61, 0xde, 0xaf, 0xf7,
	0x8d, 0xe5, 0x9c, 0x84, 0xd8, 0x53, 0xbf, 0xa3, 0x1f, 0x41, 0xd2, 0xed, 0xd2, 0xbc, 0x18, 0xb8,
	0xdf, 0x18, 0xb8, 0xd7, 0x3b, 0x38, 0x03, 0xd7, 0x9b, 0x8e, 0x7d, 0x03, 0xf7, 0x3b, 0x0f, 0x5d,
	0xae, 0xcb, 0x12, 0x6d, 0x5e, 0xa3, 0x3d, 0xc7, 0xb0, 0x6d, 0x5e, 0x4a, 0xf7, 0xcd, 0xb3, 0x3e,
	0xe1, 0xd8, 0xe3, 0x9f, 0x78, 0xf8, 0xf0, 0x31, 0x16, 0xc3, 0xee, 0xb9, 0x49, 0xe1, 0xee, 0x63,
	0x24, 0xe6, 0x2f, 0x20, 0xc5, 0xe7, 0xa2, 0x6a, 0x9d, 0xbc, 0xc4, 0xbc, 0x4b, 0x36, 0x16, 0x4b,
	0xf9, 0x1c, 0x5d, 0xba, 0xb5, 0x1c, 0xef, 0xc5, 0xd9, 0xce, 0xfa, 0x9c, 0xf2, 0x4f, 0xc2, 0x29,
	0xfb, 0x0a, 0xb6, 0x45, 0xbf, 0x59, 0x3b, 0xc2, 0x29, 0x11, 0xee, 0x0c, 0x84, 0x9b, 0x9b, 0x37,
	0x5b, 0x88, 0x0d, 0xcf, 0xf9, 0x5e, 0x3b, 0x5d, 0x8b, 0xe1, 0xeb, 0x78, 0x51, 0x58, 0xea, 0xb5,
	0x0c, 0x96, 0x63, 0x5f, 0xf6, 0xca, 0xf2, 0x70, 0xc3, 0x88, 0xb2, 0xde, 0x1c, 0xb2, 0x36, 0x06,
	0xa7, 0x57, 0x97, 0x77, 0xf7, 0x7c, 0x08, 0xf7, 0x2c, 0xfa, 0x0f, 0x53, 0x34, 0x79, 0xa3, 0xf3,
	0x67, 0x4a, 0xff, 0xa6, 0x68, 0xc5, 0x45, 0xd9, 0xdd, 0xfe, 0xe0, 0x54, 0x7f, 0xe7, 0xc3, 0xab,
	0x43, 0x48, 0x4e, 0xb1, 0x36, 0x15, 0x6f, 0xd0, 0x4f, 0xe0, 0x03, 0x58, 0x34, 0xc1, 0xdd, 0x94,
	0x7d, 0xde, 0x07, 0xbd, 0xee, 0x67, 0x53, 0xfa, 0x07, 0xff, 0xf4, 0x9f, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x03, 0xbf, 0x89, 0xd0, 0x0e, 0x08, 0x00, 0x00,
}
