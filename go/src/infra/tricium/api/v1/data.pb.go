// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/tricium/api/v1/data.proto

package tricium

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

// Available data types should be listed in this enum and have a
// corresponding nested message with a mandatory platforms fields,
// see GitFileDetails for field details.
type Data_Type int32

const (
	Data_NONE             Data_Type = 0
	Data_GIT_FILE_DETAILS Data_Type = 1
	Data_FILES            Data_Type = 2
	Data_CLANG_DETAILS    Data_Type = 3
	Data_RESULTS          Data_Type = 4
)

var Data_Type_name = map[int32]string{
	0: "NONE",
	1: "GIT_FILE_DETAILS",
	2: "FILES",
	3: "CLANG_DETAILS",
	4: "RESULTS",
}

var Data_Type_value = map[string]int32{
	"NONE":             0,
	"GIT_FILE_DETAILS": 1,
	"FILES":            2,
	"CLANG_DETAILS":    3,
	"RESULTS":          4,
}

func (x Data_Type) String() string {
	return proto.EnumName(Data_Type_name, int32(x))
}

func (Data_Type) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 0}
}

// File change status.
//
// This corresponds to the status field provided by Gerrit in FileInfo:
// https://goo.gl/ABFHDg
type Data_Status int32

const (
	Data_MODIFIED  Data_Status = 0
	Data_ADDED     Data_Status = 1
	Data_DELETED   Data_Status = 2
	Data_RENAMED   Data_Status = 3
	Data_COPIED    Data_Status = 4
	Data_REWRITTEN Data_Status = 5
)

var Data_Status_name = map[int32]string{
	0: "MODIFIED",
	1: "ADDED",
	2: "DELETED",
	3: "RENAMED",
	4: "COPIED",
	5: "REWRITTEN",
}

var Data_Status_value = map[string]int32{
	"MODIFIED":  0,
	"ADDED":     1,
	"DELETED":   2,
	"RENAMED":   3,
	"COPIED":    4,
	"REWRITTEN": 5,
}

func (x Data_Status) String() string {
	return proto.EnumName(Data_Status_name, int32(x))
}

func (Data_Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 1}
}

// Tricium data types.
//
// Any data type provided or needed by a Tricium function.
type Data struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Data) Reset()         { *m = Data{} }
func (m *Data) String() string { return proto.CompactTextString(m) }
func (*Data) ProtoMessage()    {}
func (*Data) Descriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0}
}

func (m *Data) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data.Unmarshal(m, b)
}
func (m *Data) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data.Marshal(b, m, deterministic)
}
func (m *Data) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data.Merge(m, src)
}
func (m *Data) XXX_Size() int {
	return xxx_messageInfo_Data.Size(m)
}
func (m *Data) XXX_DiscardUnknown() {
	xxx_messageInfo_Data.DiscardUnknown(m)
}

var xxx_messageInfo_Data proto.InternalMessageInfo

// Details for supported types, specifically whether a type is tied to
// a platform.
//
// These type details are used to resolve data dependencies when
// generating workflows.
type Data_TypeDetails struct {
	Type                 Data_Type `protobuf:"varint,1,opt,name=type,proto3,enum=tricium.Data_Type" json:"type,omitempty"`
	IsPlatformSpecific   bool      `protobuf:"varint,2,opt,name=is_platform_specific,json=isPlatformSpecific,proto3" json:"is_platform_specific,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *Data_TypeDetails) Reset()         { *m = Data_TypeDetails{} }
func (m *Data_TypeDetails) String() string { return proto.CompactTextString(m) }
func (*Data_TypeDetails) ProtoMessage()    {}
func (*Data_TypeDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 0}
}

func (m *Data_TypeDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data_TypeDetails.Unmarshal(m, b)
}
func (m *Data_TypeDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data_TypeDetails.Marshal(b, m, deterministic)
}
func (m *Data_TypeDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data_TypeDetails.Merge(m, src)
}
func (m *Data_TypeDetails) XXX_Size() int {
	return xxx_messageInfo_Data_TypeDetails.Size(m)
}
func (m *Data_TypeDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_Data_TypeDetails.DiscardUnknown(m)
}

var xxx_messageInfo_Data_TypeDetails proto.InternalMessageInfo

func (m *Data_TypeDetails) GetType() Data_Type {
	if m != nil {
		return m.Type
	}
	return Data_NONE
}

func (m *Data_TypeDetails) GetIsPlatformSpecific() bool {
	if m != nil {
		return m.IsPlatformSpecific
	}
	return false
}

// Details for retrieval of file content from a Git repository.
//
// In practice this was only used as an input to GitFileDetails,
// and is now DEPRECATED.
//
// ISOLATED PATH: tricium/data/git_file_details.json
type Data_GitFileDetails struct {
	// The platforms this data is tied to encoded as a bitmap.
	//
	// The bit number for each platform should correspond to the enum
	// position number of the same platform in the Platform.Name enum.
	//
	// This includes the ANY platform, encoded as zero, which should
	// be used for any data that is not platform-specific.
	Platforms            int64        `protobuf:"varint,1,opt,name=platforms,proto3" json:"platforms,omitempty"`
	Repository           string       `protobuf:"bytes,2,opt,name=repository,proto3" json:"repository,omitempty"`
	Ref                  string       `protobuf:"bytes,3,opt,name=ref,proto3" json:"ref,omitempty"`
	Files                []*Data_File `protobuf:"bytes,4,rep,name=files,proto3" json:"files,omitempty"`
	CommitMessage        string       `protobuf:"bytes,5,opt,name=commit_message,json=commitMessage,proto3" json:"commit_message,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Data_GitFileDetails) Reset()         { *m = Data_GitFileDetails{} }
func (m *Data_GitFileDetails) String() string { return proto.CompactTextString(m) }
func (*Data_GitFileDetails) ProtoMessage()    {}
func (*Data_GitFileDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 1}
}

func (m *Data_GitFileDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data_GitFileDetails.Unmarshal(m, b)
}
func (m *Data_GitFileDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data_GitFileDetails.Marshal(b, m, deterministic)
}
func (m *Data_GitFileDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data_GitFileDetails.Merge(m, src)
}
func (m *Data_GitFileDetails) XXX_Size() int {
	return xxx_messageInfo_Data_GitFileDetails.Size(m)
}
func (m *Data_GitFileDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_Data_GitFileDetails.DiscardUnknown(m)
}

var xxx_messageInfo_Data_GitFileDetails proto.InternalMessageInfo

func (m *Data_GitFileDetails) GetPlatforms() int64 {
	if m != nil {
		return m.Platforms
	}
	return 0
}

func (m *Data_GitFileDetails) GetRepository() string {
	if m != nil {
		return m.Repository
	}
	return ""
}

func (m *Data_GitFileDetails) GetRef() string {
	if m != nil {
		return m.Ref
	}
	return ""
}

func (m *Data_GitFileDetails) GetFiles() []*Data_File {
	if m != nil {
		return m.Files
	}
	return nil
}

func (m *Data_GitFileDetails) GetCommitMessage() string {
	if m != nil {
		return m.CommitMessage
	}
	return ""
}

// List of paths included in the isolated input.
//
// Files in the isolate should be laid out with the same file system
// structure as in the repository, with the root of the isolate input mapped
// to the root of the repository.
//
// ISOLATED PATH: tricium/data/files.json
type Data_Files struct {
	Platforms int64 `protobuf:"varint,1,opt,name=platforms,proto3" json:"platforms,omitempty"`
	// Files from the root of the isolated input.
	Files                []*Data_File `protobuf:"bytes,3,rep,name=files,proto3" json:"files,omitempty"`
	CommitMessage        string       `protobuf:"bytes,4,opt,name=commit_message,json=commitMessage,proto3" json:"commit_message,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Data_Files) Reset()         { *m = Data_Files{} }
func (m *Data_Files) String() string { return proto.CompactTextString(m) }
func (*Data_Files) ProtoMessage()    {}
func (*Data_Files) Descriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 2}
}

func (m *Data_Files) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data_Files.Unmarshal(m, b)
}
func (m *Data_Files) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data_Files.Marshal(b, m, deterministic)
}
func (m *Data_Files) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data_Files.Merge(m, src)
}
func (m *Data_Files) XXX_Size() int {
	return xxx_messageInfo_Data_Files.Size(m)
}
func (m *Data_Files) XXX_DiscardUnknown() {
	xxx_messageInfo_Data_Files.DiscardUnknown(m)
}

var xxx_messageInfo_Data_Files proto.InternalMessageInfo

func (m *Data_Files) GetPlatforms() int64 {
	if m != nil {
		return m.Platforms
	}
	return 0
}

func (m *Data_Files) GetFiles() []*Data_File {
	if m != nil {
		return m.Files
	}
	return nil
}

func (m *Data_Files) GetCommitMessage() string {
	if m != nil {
		return m.CommitMessage
	}
	return ""
}

type Data_File struct {
	// Path to the file from the root of the isolated input.
	//
	// The path is relative to the root of the repository being analyzed,
	// and the path separator character is "/".
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// Whether or not this file contains binary content (not text).
	IsBinary bool `protobuf:"varint,2,opt,name=is_binary,json=isBinary,proto3" json:"is_binary,omitempty"`
	// How the file was changed.
	Status               Data_Status `protobuf:"varint,3,opt,name=status,proto3,enum=tricium.Data_Status" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Data_File) Reset()         { *m = Data_File{} }
func (m *Data_File) String() string { return proto.CompactTextString(m) }
func (*Data_File) ProtoMessage()    {}
func (*Data_File) Descriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 3}
}

func (m *Data_File) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data_File.Unmarshal(m, b)
}
func (m *Data_File) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data_File.Marshal(b, m, deterministic)
}
func (m *Data_File) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data_File.Merge(m, src)
}
func (m *Data_File) XXX_Size() int {
	return xxx_messageInfo_Data_File.Size(m)
}
func (m *Data_File) XXX_DiscardUnknown() {
	xxx_messageInfo_Data_File.DiscardUnknown(m)
}

var xxx_messageInfo_Data_File proto.InternalMessageInfo

func (m *Data_File) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *Data_File) GetIsBinary() bool {
	if m != nil {
		return m.IsBinary
	}
	return false
}

func (m *Data_File) GetStatus() Data_Status {
	if m != nil {
		return m.Status
	}
	return Data_MODIFIED
}

// Details needed to replay a clang compilation.
//
// Note: This was never used in practice, and is DEPRECATED.
//
// ISOLATED PATH: tricium/data/clang_details.json
type Data_ClangDetails struct {
	Platforms int64 `protobuf:"varint,1,opt,name=platforms,proto3" json:"platforms,omitempty"`
	// Path to the compilation database. Typically, in the build root.
	CompilationDb string `protobuf:"bytes,2,opt,name=compilation_db,json=compilationDb,proto3" json:"compilation_db,omitempty"`
	// Paths to files needed to compile cpp files to analyze.
	CompDepPaths         []string `protobuf:"bytes,3,rep,name=comp_dep_paths,json=compDepPaths,proto3" json:"comp_dep_paths,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Data_ClangDetails) Reset()         { *m = Data_ClangDetails{} }
func (m *Data_ClangDetails) String() string { return proto.CompactTextString(m) }
func (*Data_ClangDetails) ProtoMessage()    {}
func (*Data_ClangDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 4}
}

func (m *Data_ClangDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data_ClangDetails.Unmarshal(m, b)
}
func (m *Data_ClangDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data_ClangDetails.Marshal(b, m, deterministic)
}
func (m *Data_ClangDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data_ClangDetails.Merge(m, src)
}
func (m *Data_ClangDetails) XXX_Size() int {
	return xxx_messageInfo_Data_ClangDetails.Size(m)
}
func (m *Data_ClangDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_Data_ClangDetails.DiscardUnknown(m)
}

var xxx_messageInfo_Data_ClangDetails proto.InternalMessageInfo

func (m *Data_ClangDetails) GetPlatforms() int64 {
	if m != nil {
		return m.Platforms
	}
	return 0
}

func (m *Data_ClangDetails) GetCompilationDb() string {
	if m != nil {
		return m.CompilationDb
	}
	return ""
}

func (m *Data_ClangDetails) GetCompDepPaths() []string {
	if m != nil {
		return m.CompDepPaths
	}
	return nil
}

// Results from running a Tricium analyzer.
//
// Results are returned to the Tricium service via isolated output from
// swarming tasks executing Tricium workers or from Buildbucket presentation
// properties on executed Tricium recipes.
//
// ISOLATED PATH: tricium/data/results.json
// BUILDBUCKET PROPERTIES: output.properties.comments
//                         output.properties.num_comments
type Data_Results struct {
	Platforms int64 `protobuf:"varint,1,opt,name=platforms,proto3" json:"platforms,omitempty"`
	// Zero or more results found as comments, either inline comments or change
	// comments (comments without line positions).
	Comments             []*Data_Comment `protobuf:"bytes,2,rep,name=comments,proto3" json:"comments,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *Data_Results) Reset()         { *m = Data_Results{} }
func (m *Data_Results) String() string { return proto.CompactTextString(m) }
func (*Data_Results) ProtoMessage()    {}
func (*Data_Results) Descriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 5}
}

func (m *Data_Results) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data_Results.Unmarshal(m, b)
}
func (m *Data_Results) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data_Results.Marshal(b, m, deterministic)
}
func (m *Data_Results) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data_Results.Merge(m, src)
}
func (m *Data_Results) XXX_Size() int {
	return xxx_messageInfo_Data_Results.Size(m)
}
func (m *Data_Results) XXX_DiscardUnknown() {
	xxx_messageInfo_Data_Results.DiscardUnknown(m)
}

var xxx_messageInfo_Data_Results proto.InternalMessageInfo

func (m *Data_Results) GetPlatforms() int64 {
	if m != nil {
		return m.Platforms
	}
	return 0
}

func (m *Data_Results) GetComments() []*Data_Comment {
	if m != nil {
		return m.Comments
	}
	return nil
}

// Results.Comment, results as comments.
//
// Similar content as that needed to provide robot comments in Gerrit,
// https://gerrit-review.googlesource.com/Documentation/config-robot-comments.html
type Data_Comment struct {
	// Comment ID.
	//
	// This is an UUID generated by the Tricium service and used for tracking
	// of comment feedback. Analyzers should leave this field empty.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Category of the result, encoded as a path with the analyzer name as the
	// root, followed by an arbitrary number of subcategories, for example
	// "ClangTidy/llvm-header-guard".
	Category string `protobuf:"bytes,2,opt,name=category,proto3" json:"category,omitempty"`
	// Comment message. This should be a short message suitable as a code
	// review comment.
	Message string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	// Path to the file this comment is for.
	//
	// If this path is the empty string, then the comment is on the commit
	// message text, rather than an actual file.
	Path string `protobuf:"bytes,5,opt,name=path,proto3" json:"path,omitempty"`
	// Position information. If start_line is omitted, then the comment
	// will be a file-level comment.
	StartLine int32 `protobuf:"varint,6,opt,name=start_line,json=startLine,proto3" json:"start_line,omitempty"`
	EndLine   int32 `protobuf:"varint,7,opt,name=end_line,json=endLine,proto3" json:"end_line,omitempty"`
	StartChar int32 `protobuf:"varint,8,opt,name=start_char,json=startChar,proto3" json:"start_char,omitempty"`
	EndChar   int32 `protobuf:"varint,9,opt,name=end_char,json=endChar,proto3" json:"end_char,omitempty"`
	// Suggested fixes for the identified issue.
	Suggestions []*Data_Suggestion `protobuf:"bytes,10,rep,name=suggestions,proto3" json:"suggestions,omitempty"`
	// When true, show on both changed and unchanged lines.
	// When false, only show on changed lines.
	ShowOnUnchangedLines bool     `protobuf:"varint,11,opt,name=show_on_unchanged_lines,json=showOnUnchangedLines,proto3" json:"show_on_unchanged_lines,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Data_Comment) Reset()         { *m = Data_Comment{} }
func (m *Data_Comment) String() string { return proto.CompactTextString(m) }
func (*Data_Comment) ProtoMessage()    {}
func (*Data_Comment) Descriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 6}
}

func (m *Data_Comment) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data_Comment.Unmarshal(m, b)
}
func (m *Data_Comment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data_Comment.Marshal(b, m, deterministic)
}
func (m *Data_Comment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data_Comment.Merge(m, src)
}
func (m *Data_Comment) XXX_Size() int {
	return xxx_messageInfo_Data_Comment.Size(m)
}
func (m *Data_Comment) XXX_DiscardUnknown() {
	xxx_messageInfo_Data_Comment.DiscardUnknown(m)
}

var xxx_messageInfo_Data_Comment proto.InternalMessageInfo

func (m *Data_Comment) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Data_Comment) GetCategory() string {
	if m != nil {
		return m.Category
	}
	return ""
}

func (m *Data_Comment) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *Data_Comment) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *Data_Comment) GetStartLine() int32 {
	if m != nil {
		return m.StartLine
	}
	return 0
}

func (m *Data_Comment) GetEndLine() int32 {
	if m != nil {
		return m.EndLine
	}
	return 0
}

func (m *Data_Comment) GetStartChar() int32 {
	if m != nil {
		return m.StartChar
	}
	return 0
}

func (m *Data_Comment) GetEndChar() int32 {
	if m != nil {
		return m.EndChar
	}
	return 0
}

func (m *Data_Comment) GetSuggestions() []*Data_Suggestion {
	if m != nil {
		return m.Suggestions
	}
	return nil
}

func (m *Data_Comment) GetShowOnUnchangedLines() bool {
	if m != nil {
		return m.ShowOnUnchangedLines
	}
	return false
}

// Suggested fix.
//
// A fix may include replacements in any file in the same repo as the file of
// the corresponding comment.
type Data_Suggestion struct {
	// A brief description of the suggested fix.
	Description string `protobuf:"bytes,1,opt,name=description,proto3" json:"description,omitempty"`
	// Fix as a list of replacements.
	Replacements         []*Data_Replacement `protobuf:"bytes,2,rep,name=replacements,proto3" json:"replacements,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *Data_Suggestion) Reset()         { *m = Data_Suggestion{} }
func (m *Data_Suggestion) String() string { return proto.CompactTextString(m) }
func (*Data_Suggestion) ProtoMessage()    {}
func (*Data_Suggestion) Descriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 7}
}

func (m *Data_Suggestion) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data_Suggestion.Unmarshal(m, b)
}
func (m *Data_Suggestion) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data_Suggestion.Marshal(b, m, deterministic)
}
func (m *Data_Suggestion) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data_Suggestion.Merge(m, src)
}
func (m *Data_Suggestion) XXX_Size() int {
	return xxx_messageInfo_Data_Suggestion.Size(m)
}
func (m *Data_Suggestion) XXX_DiscardUnknown() {
	xxx_messageInfo_Data_Suggestion.DiscardUnknown(m)
}

var xxx_messageInfo_Data_Suggestion proto.InternalMessageInfo

func (m *Data_Suggestion) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Data_Suggestion) GetReplacements() []*Data_Replacement {
	if m != nil {
		return m.Replacements
	}
	return nil
}

// A suggested replacement.
//
// The replacement should be for one continuous section of a file.
type Data_Replacement struct {
	// Path to the file for this replacement.
	//
	// An empty string indicates the commit message.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// A replacement string.
	Replacement string `protobuf:"bytes,2,opt,name=replacement,proto3" json:"replacement,omitempty"`
	// A continuous section of the file to replace.
	StartLine            int32    `protobuf:"varint,3,opt,name=start_line,json=startLine,proto3" json:"start_line,omitempty"`
	EndLine              int32    `protobuf:"varint,4,opt,name=end_line,json=endLine,proto3" json:"end_line,omitempty"`
	StartChar            int32    `protobuf:"varint,5,opt,name=start_char,json=startChar,proto3" json:"start_char,omitempty"`
	EndChar              int32    `protobuf:"varint,6,opt,name=end_char,json=endChar,proto3" json:"end_char,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Data_Replacement) Reset()         { *m = Data_Replacement{} }
func (m *Data_Replacement) String() string { return proto.CompactTextString(m) }
func (*Data_Replacement) ProtoMessage()    {}
func (*Data_Replacement) Descriptor() ([]byte, []int) {
	return fileDescriptor_5414e8d9ed493bf0, []int{0, 8}
}

func (m *Data_Replacement) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data_Replacement.Unmarshal(m, b)
}
func (m *Data_Replacement) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data_Replacement.Marshal(b, m, deterministic)
}
func (m *Data_Replacement) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data_Replacement.Merge(m, src)
}
func (m *Data_Replacement) XXX_Size() int {
	return xxx_messageInfo_Data_Replacement.Size(m)
}
func (m *Data_Replacement) XXX_DiscardUnknown() {
	xxx_messageInfo_Data_Replacement.DiscardUnknown(m)
}

var xxx_messageInfo_Data_Replacement proto.InternalMessageInfo

func (m *Data_Replacement) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *Data_Replacement) GetReplacement() string {
	if m != nil {
		return m.Replacement
	}
	return ""
}

func (m *Data_Replacement) GetStartLine() int32 {
	if m != nil {
		return m.StartLine
	}
	return 0
}

func (m *Data_Replacement) GetEndLine() int32 {
	if m != nil {
		return m.EndLine
	}
	return 0
}

func (m *Data_Replacement) GetStartChar() int32 {
	if m != nil {
		return m.StartChar
	}
	return 0
}

func (m *Data_Replacement) GetEndChar() int32 {
	if m != nil {
		return m.EndChar
	}
	return 0
}

func init() {
	proto.RegisterEnum("tricium.Data_Type", Data_Type_name, Data_Type_value)
	proto.RegisterEnum("tricium.Data_Status", Data_Status_name, Data_Status_value)
	proto.RegisterType((*Data)(nil), "tricium.Data")
	proto.RegisterType((*Data_TypeDetails)(nil), "tricium.Data.TypeDetails")
	proto.RegisterType((*Data_GitFileDetails)(nil), "tricium.Data.GitFileDetails")
	proto.RegisterType((*Data_Files)(nil), "tricium.Data.Files")
	proto.RegisterType((*Data_File)(nil), "tricium.Data.File")
	proto.RegisterType((*Data_ClangDetails)(nil), "tricium.Data.ClangDetails")
	proto.RegisterType((*Data_Results)(nil), "tricium.Data.Results")
	proto.RegisterType((*Data_Comment)(nil), "tricium.Data.Comment")
	proto.RegisterType((*Data_Suggestion)(nil), "tricium.Data.Suggestion")
	proto.RegisterType((*Data_Replacement)(nil), "tricium.Data.Replacement")
}

func init() {
	proto.RegisterFile("infra/tricium/api/v1/data.proto", fileDescriptor_5414e8d9ed493bf0)
}

var fileDescriptor_5414e8d9ed493bf0 = []byte{
	// 758 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x55, 0x5f, 0x8f, 0xab, 0x44,
	0x1c, 0xbd, 0x14, 0x68, 0xe1, 0x47, 0xb7, 0xc1, 0xc9, 0x1a, 0xb9, 0xf8, 0xaf, 0xd9, 0xa8, 0xe9,
	0x83, 0xd9, 0xf5, 0x5e, 0xe3, 0x8b, 0x89, 0x0f, 0x6b, 0x61, 0x37, 0x35, 0xdd, 0xee, 0x3a, 0xed,
	0xd5, 0xc4, 0x17, 0x32, 0x85, 0x69, 0x3b, 0x09, 0x05, 0xc2, 0x4c, 0xaf, 0xe9, 0x9b, 0x5f, 0xc9,
	0x27, 0xbf, 0x86, 0x1f, 0xc9, 0xcc, 0x40, 0x29, 0xdd, 0x6c, 0xaa, 0xbe, 0x31, 0xe7, 0x9c, 0xdf,
	0x9f, 0x9e, 0x33, 0x14, 0xf8, 0x9c, 0x65, 0xab, 0x92, 0xdc, 0x88, 0x92, 0xc5, 0x6c, 0xb7, 0xbd,
	0x21, 0x05, 0xbb, 0x79, 0xff, 0xe6, 0x26, 0x21, 0x82, 0x5c, 0x17, 0x65, 0x2e, 0x72, 0xd4, 0xab,
	0xa9, 0xab, 0x3f, 0xfa, 0x60, 0x04, 0x44, 0x10, 0x7f, 0x0d, 0xce, 0x62, 0x5f, 0xd0, 0x80, 0x0a,
	0xc2, 0x52, 0x8e, 0xbe, 0x02, 0x43, 0xec, 0x0b, 0xea, 0x69, 0x43, 0x6d, 0x34, 0x78, 0x8b, 0xae,
	0x6b, 0xfd, 0xb5, 0xd4, 0x5e, 0x4b, 0x21, 0x56, 0x3c, 0xfa, 0x06, 0x2e, 0x19, 0x8f, 0x8a, 0x94,
	0x88, 0x55, 0x5e, 0x6e, 0x23, 0x5e, 0xd0, 0x98, 0xad, 0x58, 0xec, 0x75, 0x86, 0xda, 0xc8, 0xc2,
	0x88, 0xf1, 0xa7, 0x9a, 0x9a, 0xd7, 0x8c, 0xff, 0xa7, 0x06, 0x83, 0x7b, 0x26, 0xee, 0x58, 0xda,
	0x0c, 0xfb, 0x04, 0xec, 0x43, 0x07, 0xae, 0x26, 0xea, 0xf8, 0x08, 0xa0, 0xcf, 0x00, 0x4a, 0x5a,
	0xe4, 0x9c, 0x89, 0xbc, 0xdc, 0xab, 0xc6, 0x36, 0x6e, 0x21, 0xc8, 0x05, 0xbd, 0xa4, 0x2b, 0x4f,
	0x57, 0x84, 0x7c, 0x44, 0x23, 0x30, 0x57, 0x2c, 0xa5, 0xdc, 0x33, 0x86, 0xfa, 0xc8, 0x79, 0xbe,
	0xbd, 0x9c, 0x8c, 0x2b, 0x01, 0xfa, 0x12, 0x06, 0x71, 0xbe, 0xdd, 0x32, 0x11, 0x6d, 0x29, 0xe7,
	0x64, 0x4d, 0x3d, 0x53, 0xb5, 0xb9, 0xa8, 0xd0, 0x87, 0x0a, 0xf4, 0xdf, 0x83, 0x79, 0xa7, 0xf4,
	0xe7, 0x37, 0x6d, 0xe6, 0xea, 0xff, 0x7f, 0xae, 0xf1, 0xd2, 0x5c, 0x0a, 0x86, 0xac, 0x42, 0x08,
	0x8c, 0x82, 0x88, 0x8d, 0x9a, 0x68, 0x63, 0xf5, 0x8c, 0x3e, 0x06, 0x9b, 0xf1, 0x68, 0xc9, 0x32,
	0x52, 0xbb, 0x62, 0x61, 0x8b, 0xf1, 0x1f, 0xd5, 0x19, 0x7d, 0x0d, 0x5d, 0x2e, 0x88, 0xd8, 0x71,
	0x65, 0xcb, 0xe0, 0xed, 0xe5, 0xe9, 0x2a, 0x73, 0xc5, 0xe1, 0x5a, 0xe3, 0xef, 0xa1, 0x3f, 0x4e,
	0x49, 0xb6, 0xfe, 0x6f, 0x79, 0x54, 0xbb, 0x17, 0x2c, 0x25, 0x82, 0xe5, 0x59, 0x94, 0x2c, 0xeb,
	0x4c, 0x2e, 0x5a, 0x68, 0xb0, 0x44, 0x5f, 0x54, 0xb2, 0x28, 0xa1, 0x45, 0x24, 0x17, 0xae, 0x5c,
	0xb1, 0x71, 0x5f, 0xa2, 0x01, 0x2d, 0x9e, 0x24, 0xe6, 0xff, 0x06, 0x3d, 0x4c, 0xf9, 0x2e, 0x15,
	0xff, 0x36, 0xf5, 0x0d, 0x58, 0xd2, 0x1b, 0x9a, 0x09, 0xee, 0x75, 0x94, 0xbd, 0x1f, 0x9e, 0xfe,
	0xa6, 0x71, 0xc5, 0xe2, 0x46, 0xe6, 0xff, 0xdd, 0x81, 0x5e, 0x8d, 0xa2, 0x01, 0x74, 0x58, 0x52,
	0xfb, 0xd7, 0x61, 0x09, 0xf2, 0xc1, 0x8a, 0x89, 0xa0, 0xeb, 0xe3, 0x95, 0x6a, 0xce, 0xc8, 0x83,
	0xde, 0x21, 0x95, 0xea, 0x52, 0x1d, 0x8e, 0x4d, 0x0e, 0x66, 0x2b, 0x87, 0x4f, 0x01, 0xb8, 0x20,
	0xa5, 0x88, 0x52, 0x96, 0x51, 0xaf, 0x3b, 0xd4, 0x46, 0x26, 0xb6, 0x15, 0x32, 0x65, 0x19, 0x45,
	0xaf, 0xc1, 0xa2, 0x59, 0x52, 0x91, 0x3d, 0x45, 0xf6, 0x68, 0x96, 0x28, 0xaa, 0xa9, 0x8c, 0x37,
	0xa4, 0xf4, 0xac, 0x56, 0xe5, 0x78, 0x43, 0xca, 0x43, 0xa5, 0x22, 0xed, 0xa6, 0x52, 0x51, 0xdf,
	0x83, 0xc3, 0x77, 0xeb, 0x35, 0xe5, 0xd2, 0x6b, 0xee, 0x81, 0xf2, 0xc3, 0x7b, 0x96, 0x71, 0x23,
	0xc0, 0x6d, 0x31, 0xfa, 0x0e, 0x3e, 0xe2, 0x9b, 0xfc, 0xf7, 0x28, 0xcf, 0xa2, 0x5d, 0x16, 0x6f,
	0x48, 0xb6, 0xa6, 0xd5, 0x7a, 0xdc, 0x73, 0xd4, 0x2d, 0xba, 0x94, 0xf4, 0x63, 0xf6, 0xee, 0x40,
	0xca, 0x5d, 0xf9, 0x4f, 0x86, 0x65, 0xb8, 0xa6, 0xbf, 0x05, 0x38, 0xf6, 0x45, 0x43, 0x70, 0x12,
	0xca, 0xe3, 0x92, 0x15, 0xf2, 0x58, 0xbb, 0xdb, 0x86, 0xd0, 0x0f, 0xd0, 0x2f, 0x69, 0x91, 0x92,
	0x98, 0xb6, 0x93, 0x7b, 0x7d, 0xba, 0x29, 0x3e, 0x2a, 0xf0, 0x89, 0xdc, 0xff, 0x4b, 0x03, 0xa7,
	0xc5, 0xbe, 0xf8, 0x1e, 0x0c, 0xc1, 0x69, 0xd5, 0xd4, 0x61, 0xb6, 0xa1, 0x67, 0x09, 0xe9, 0xe7,
	0x12, 0x32, 0xce, 0x25, 0x64, 0x9e, 0x4b, 0xa8, 0x7b, 0x92, 0xd0, 0xd5, 0xcf, 0x60, 0xc8, 0x7f,
	0x49, 0x64, 0x81, 0x31, 0x7b, 0x9c, 0x85, 0xee, 0x2b, 0x74, 0x09, 0xee, 0xfd, 0x64, 0x11, 0xdd,
	0x4d, 0xa6, 0x61, 0x14, 0x84, 0x8b, 0xdb, 0xc9, 0x74, 0xee, 0x6a, 0xc8, 0x06, 0x53, 0x22, 0x73,
	0xb7, 0x83, 0x3e, 0x80, 0x8b, 0xf1, 0xf4, 0x76, 0x76, 0xdf, 0xb0, 0x3a, 0x72, 0xa0, 0x87, 0xc3,
	0xf9, 0xbb, 0xe9, 0x62, 0xee, 0x1a, 0x57, 0xbf, 0x40, 0xb7, 0x7a, 0x6f, 0x51, 0x1f, 0xac, 0x87,
	0xc7, 0x60, 0x72, 0x37, 0x09, 0x03, 0xf7, 0x95, 0x6c, 0x71, 0x1b, 0x04, 0x61, 0xe0, 0x6a, 0x52,
	0x1f, 0x84, 0xd3, 0x70, 0x11, 0x06, 0x6e, 0xa7, 0x2a, 0x9e, 0xdd, 0x3e, 0x84, 0x81, 0xab, 0x23,
	0x80, 0xee, 0xf8, 0xf1, 0x49, 0x16, 0x18, 0xe8, 0x02, 0x6c, 0x1c, 0xfe, 0x8a, 0x27, 0x8b, 0x45,
	0x38, 0x73, 0xcd, 0x65, 0x57, 0x7d, 0x12, 0xbe, 0xfd, 0x27, 0x00, 0x00, 0xff, 0xff, 0x6e, 0x10,
	0x3d, 0x60, 0x35, 0x06, 0x00, 0x00,
}
