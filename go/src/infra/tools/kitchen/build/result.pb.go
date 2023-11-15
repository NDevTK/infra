// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.1
// source: infra/tools/kitchen/build/result.proto

package build

import (
	proto "github.com/golang/protobuf/proto"
	proto1 "go.chromium.org/luci/luciexe/legacy/annotee/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	recipe_engine "infra/tools/kitchen/third_party/recipe_engine"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type InfraFailure_Type int32

const (
	// An unexpected internal error happened inside build bootstrapper.
	InfraFailure_BOOTSTRAPPER_ERROR InfraFailure_Type = 0
	// Build bootstrapper input was invalid, for example flags.
	InfraFailure_INVALID_INPUT InfraFailure_Type = 1
	// Build run was cancelled.
	InfraFailure_CANCELED InfraFailure_Type = 2
	// Recipe has infra-failed.
	InfraFailure_RECIPE_INFRA_FAILURE InfraFailure_Type = 3
)

// Enum value maps for InfraFailure_Type.
var (
	InfraFailure_Type_name = map[int32]string{
		0: "BOOTSTRAPPER_ERROR",
		1: "INVALID_INPUT",
		2: "CANCELED",
		3: "RECIPE_INFRA_FAILURE",
	}
	InfraFailure_Type_value = map[string]int32{
		"BOOTSTRAPPER_ERROR":   0,
		"INVALID_INPUT":        1,
		"CANCELED":             2,
		"RECIPE_INFRA_FAILURE": 3,
	}
)

func (x InfraFailure_Type) Enum() *InfraFailure_Type {
	p := new(InfraFailure_Type)
	*p = x
	return p
}

func (x InfraFailure_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (InfraFailure_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_tools_kitchen_build_result_proto_enumTypes[0].Descriptor()
}

func (InfraFailure_Type) Type() protoreflect.EnumType {
	return &file_infra_tools_kitchen_build_result_proto_enumTypes[0]
}

func (x InfraFailure_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use InfraFailure_Type.Descriptor instead.
func (InfraFailure_Type) EnumDescriptor() ([]byte, []int) {
	return file_infra_tools_kitchen_build_result_proto_rawDescGZIP(), []int{2, 0}
}

// Result of a build run from the perspective of a build bootstrapper.
type BuildRunResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A failure of the build infrastructure.
	// A build with an infra failure cannot be used to judge the build input
	// (e.g. a CL).
	InfraFailure *InfraFailure `protobuf:"bytes,1,opt,name=infra_failure,json=infraFailure,proto3" json:"infra_failure,omitempty"`
	// recipes.py's exit code.
	RecipeExitCode *OptionalInt32 `protobuf:"bytes,2,opt,name=recipe_exit_code,json=recipeExitCode,proto3" json:"recipe_exit_code,omitempty"`
	// Results returned by recipe engine.
	RecipeResult *recipe_engine.Result `protobuf:"bytes,3,opt,name=recipe_result,json=recipeResult,proto3" json:"recipe_result,omitempty"`
	// Final state of the annotation proto.
	Annotations *proto1.Step `protobuf:"bytes,4,opt,name=annotations,proto3" json:"annotations,omitempty"`
	// URL where the annotations (above) can be fetched from.
	// In practice, a LogDog URL.
	AnnotationUrl string                 `protobuf:"bytes,5,opt,name=annotation_url,json=annotationUrl,proto3" json:"annotation_url,omitempty"`
	Recipe        *BuildRunResult_Recipe `protobuf:"bytes,6,opt,name=recipe,proto3" json:"recipe,omitempty"`
}

func (x *BuildRunResult) Reset() {
	*x = BuildRunResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_kitchen_build_result_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildRunResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildRunResult) ProtoMessage() {}

func (x *BuildRunResult) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_kitchen_build_result_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildRunResult.ProtoReflect.Descriptor instead.
func (*BuildRunResult) Descriptor() ([]byte, []int) {
	return file_infra_tools_kitchen_build_result_proto_rawDescGZIP(), []int{0}
}

func (x *BuildRunResult) GetInfraFailure() *InfraFailure {
	if x != nil {
		return x.InfraFailure
	}
	return nil
}

func (x *BuildRunResult) GetRecipeExitCode() *OptionalInt32 {
	if x != nil {
		return x.RecipeExitCode
	}
	return nil
}

func (x *BuildRunResult) GetRecipeResult() *recipe_engine.Result {
	if x != nil {
		return x.RecipeResult
	}
	return nil
}

func (x *BuildRunResult) GetAnnotations() *proto1.Step {
	if x != nil {
		return x.Annotations
	}
	return nil
}

func (x *BuildRunResult) GetAnnotationUrl() string {
	if x != nil {
		return x.AnnotationUrl
	}
	return ""
}

func (x *BuildRunResult) GetRecipe() *BuildRunResult_Recipe {
	if x != nil {
		return x.Recipe
	}
	return nil
}

// An int32 that can be distinguished from absence of the value.
type OptionalInt32 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value int32 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *OptionalInt32) Reset() {
	*x = OptionalInt32{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_kitchen_build_result_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OptionalInt32) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OptionalInt32) ProtoMessage() {}

func (x *OptionalInt32) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_kitchen_build_result_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OptionalInt32.ProtoReflect.Descriptor instead.
func (*OptionalInt32) Descriptor() ([]byte, []int) {
	return file_infra_tools_kitchen_build_result_proto_rawDescGZIP(), []int{1}
}

func (x *OptionalInt32) GetValue() int32 {
	if x != nil {
		return x.Value
	}
	return 0
}

type InfraFailure struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type InfraFailure_Type `protobuf:"varint,1,opt,name=type,proto3,enum=luci.build.InfraFailure_Type" json:"type,omitempty"`
	Text string            `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	// Build bootstrapper call stack is populated on bootstrapper errors.
	BootstrapperCallStack []string `protobuf:"bytes,3,rep,name=bootstrapper_call_stack,json=bootstrapperCallStack,proto3" json:"bootstrapper_call_stack,omitempty"`
}

func (x *InfraFailure) Reset() {
	*x = InfraFailure{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_kitchen_build_result_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InfraFailure) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InfraFailure) ProtoMessage() {}

func (x *InfraFailure) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_kitchen_build_result_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InfraFailure.ProtoReflect.Descriptor instead.
func (*InfraFailure) Descriptor() ([]byte, []int) {
	return file_infra_tools_kitchen_build_result_proto_rawDescGZIP(), []int{2}
}

func (x *InfraFailure) GetType() InfraFailure_Type {
	if x != nil {
		return x.Type
	}
	return InfraFailure_BOOTSTRAPPER_ERROR
}

func (x *InfraFailure) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *InfraFailure) GetBootstrapperCallStack() []string {
	if x != nil {
		return x.BootstrapperCallStack
	}
	return nil
}

// Information about the recipe checkout that was used. If the recipes
// weren't checked out (e.g., isolated bundle, local), this will be empty.
type BuildRunResult_Recipe struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the recipe that was executed.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Repository is the recipe repository that was checked out.
	// will not have a revision field.
	Repository string `protobuf:"bytes,2,opt,name=repository,proto3" json:"repository,omitempty"`
	// The revision of the recipe repository.
	Revision string `protobuf:"bytes,3,opt,name=revision,proto3" json:"revision,omitempty"`
}

func (x *BuildRunResult_Recipe) Reset() {
	*x = BuildRunResult_Recipe{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_kitchen_build_result_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildRunResult_Recipe) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildRunResult_Recipe) ProtoMessage() {}

func (x *BuildRunResult_Recipe) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_kitchen_build_result_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildRunResult_Recipe.ProtoReflect.Descriptor instead.
func (*BuildRunResult_Recipe) Descriptor() ([]byte, []int) {
	return file_infra_tools_kitchen_build_result_proto_rawDescGZIP(), []int{0, 0}
}

func (x *BuildRunResult_Recipe) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *BuildRunResult_Recipe) GetRepository() string {
	if x != nil {
		return x.Repository
	}
	return ""
}

func (x *BuildRunResult_Recipe) GetRevision() string {
	if x != nil {
		return x.Revision
	}
	return ""
}

var File_infra_tools_kitchen_build_result_proto protoreflect.FileDescriptor

var file_infra_tools_kitchen_build_result_proto_rawDesc = []byte{
	0x0a, 0x26, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x6b, 0x69,
	0x74, 0x63, 0x68, 0x65, 0x6e, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2f, 0x72, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x6c, 0x75, 0x63, 0x69, 0x2e, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x1a, 0x3a, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74, 0x6f, 0x6f, 0x6c,
	0x73, 0x2f, 0x6b, 0x69, 0x74, 0x63, 0x68, 0x65, 0x6e, 0x2f, 0x74, 0x68, 0x69, 0x72, 0x64, 0x5f,
	0x70, 0x61, 0x72, 0x74, 0x79, 0x2f, 0x72, 0x65, 0x63, 0x69, 0x70, 0x65, 0x5f, 0x65, 0x6e, 0x67,
	0x69, 0x6e, 0x65, 0x2f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x43, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x2e, 0x6f, 0x72,
	0x67, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x65, 0x78, 0x65, 0x2f, 0x6c,
	0x65, 0x67, 0x61, 0x63, 0x79, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x65, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc0, 0x03, 0x0a, 0x0e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x52,
	0x75, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x3d, 0x0a, 0x0d, 0x69, 0x6e, 0x66, 0x72,
	0x61, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x18, 0x2e, 0x6c, 0x75, 0x63, 0x69, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x49, 0x6e, 0x66,
	0x72, 0x61, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x52, 0x0c, 0x69, 0x6e, 0x66, 0x72, 0x61,
	0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x12, 0x43, 0x0a, 0x10, 0x72, 0x65, 0x63, 0x69, 0x70,
	0x65, 0x5f, 0x65, 0x78, 0x69, 0x74, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x6c, 0x75, 0x63, 0x69, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x52, 0x0e, 0x72, 0x65,
	0x63, 0x69, 0x70, 0x65, 0x45, 0x78, 0x69, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x3a, 0x0a, 0x0d,
	0x72, 0x65, 0x63, 0x69, 0x70, 0x65, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x72, 0x65, 0x63, 0x69, 0x70, 0x65, 0x5f, 0x65, 0x6e, 0x67,
	0x69, 0x6e, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x0c, 0x72, 0x65, 0x63, 0x69,
	0x70, 0x65, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x32, 0x0a, 0x0b, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e,
	0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x53, 0x74, 0x65, 0x70, 0x52,
	0x0b, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x25, 0x0a, 0x0e,
	0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x55, 0x72, 0x6c, 0x12, 0x39, 0x0a, 0x06, 0x72, 0x65, 0x63, 0x69, 0x70, 0x65, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6c, 0x75, 0x63, 0x69, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x2e,
	0x52, 0x65, 0x63, 0x69, 0x70, 0x65, 0x52, 0x06, 0x72, 0x65, 0x63, 0x69, 0x70, 0x65, 0x1a, 0x58,
	0x0a, 0x06, 0x52, 0x65, 0x63, 0x69, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a,
	0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x1a, 0x0a, 0x08,
	0x72, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x72, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x25, 0x0a, 0x0d, 0x4f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x61, 0x6c, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22,
	0xe8, 0x01, 0x0a, 0x0c, 0x49, 0x6e, 0x66, 0x72, 0x61, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65,
	0x12, 0x31, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1d,
	0x2e, 0x6c, 0x75, 0x63, 0x69, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x49, 0x6e, 0x66, 0x72,
	0x61, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x12, 0x36, 0x0a, 0x17, 0x62, 0x6f, 0x6f, 0x74, 0x73,
	0x74, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x5f, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x73, 0x74, 0x61,
	0x63, 0x6b, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x15, 0x62, 0x6f, 0x6f, 0x74, 0x73, 0x74,
	0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x43, 0x61, 0x6c, 0x6c, 0x53, 0x74, 0x61, 0x63, 0x6b, 0x22,
	0x59, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x12, 0x42, 0x4f, 0x4f, 0x54, 0x53,
	0x54, 0x52, 0x41, 0x50, 0x50, 0x45, 0x52, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x00, 0x12,
	0x11, 0x0a, 0x0d, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x49, 0x4e, 0x50, 0x55, 0x54,
	0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x45, 0x44, 0x10, 0x02,
	0x12, 0x18, 0x0a, 0x14, 0x52, 0x45, 0x43, 0x49, 0x50, 0x45, 0x5f, 0x49, 0x4e, 0x46, 0x52, 0x41,
	0x5f, 0x46, 0x41, 0x49, 0x4c, 0x55, 0x52, 0x45, 0x10, 0x03, 0x42, 0x21, 0x5a, 0x1f, 0x69, 0x6e,
	0x66, 0x72, 0x61, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x6b, 0x69, 0x74, 0x63, 0x68, 0x65,
	0x6e, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x3b, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_tools_kitchen_build_result_proto_rawDescOnce sync.Once
	file_infra_tools_kitchen_build_result_proto_rawDescData = file_infra_tools_kitchen_build_result_proto_rawDesc
)

func file_infra_tools_kitchen_build_result_proto_rawDescGZIP() []byte {
	file_infra_tools_kitchen_build_result_proto_rawDescOnce.Do(func() {
		file_infra_tools_kitchen_build_result_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_tools_kitchen_build_result_proto_rawDescData)
	})
	return file_infra_tools_kitchen_build_result_proto_rawDescData
}

var file_infra_tools_kitchen_build_result_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_infra_tools_kitchen_build_result_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_infra_tools_kitchen_build_result_proto_goTypes = []interface{}{
	(InfraFailure_Type)(0),        // 0: luci.build.InfraFailure.Type
	(*BuildRunResult)(nil),        // 1: luci.build.BuildRunResult
	(*OptionalInt32)(nil),         // 2: luci.build.OptionalInt32
	(*InfraFailure)(nil),          // 3: luci.build.InfraFailure
	(*BuildRunResult_Recipe)(nil), // 4: luci.build.BuildRunResult.Recipe
	(*recipe_engine.Result)(nil),  // 5: recipe_engine.Result
	(*proto1.Step)(nil),           // 6: annotation.Step
}
var file_infra_tools_kitchen_build_result_proto_depIdxs = []int32{
	3, // 0: luci.build.BuildRunResult.infra_failure:type_name -> luci.build.InfraFailure
	2, // 1: luci.build.BuildRunResult.recipe_exit_code:type_name -> luci.build.OptionalInt32
	5, // 2: luci.build.BuildRunResult.recipe_result:type_name -> recipe_engine.Result
	6, // 3: luci.build.BuildRunResult.annotations:type_name -> annotation.Step
	4, // 4: luci.build.BuildRunResult.recipe:type_name -> luci.build.BuildRunResult.Recipe
	0, // 5: luci.build.InfraFailure.type:type_name -> luci.build.InfraFailure.Type
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_infra_tools_kitchen_build_result_proto_init() }
func file_infra_tools_kitchen_build_result_proto_init() {
	if File_infra_tools_kitchen_build_result_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_tools_kitchen_build_result_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuildRunResult); i {
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
		file_infra_tools_kitchen_build_result_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OptionalInt32); i {
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
		file_infra_tools_kitchen_build_result_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InfraFailure); i {
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
		file_infra_tools_kitchen_build_result_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuildRunResult_Recipe); i {
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
			RawDescriptor: file_infra_tools_kitchen_build_result_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_tools_kitchen_build_result_proto_goTypes,
		DependencyIndexes: file_infra_tools_kitchen_build_result_proto_depIdxs,
		EnumInfos:         file_infra_tools_kitchen_build_result_proto_enumTypes,
		MessageInfos:      file_infra_tools_kitchen_build_result_proto_msgTypes,
	}.Build()
	File_infra_tools_kitchen_build_result_proto = out.File
	file_infra_tools_kitchen_build_result_proto_rawDesc = nil
	file_infra_tools_kitchen_build_result_proto_goTypes = nil
	file_infra_tools_kitchen_build_result_proto_depIdxs = nil
}