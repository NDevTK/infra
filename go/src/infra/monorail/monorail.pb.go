// Code generated by protoc-gen-go.
// source: monorail.proto
// DO NOT EDIT!

/*
Package monorail is a generated protocol buffer package.

It is generated from these files:
	monorail.proto

It has these top-level messages:
	Issue
	IssueRef
	InsertIssueRequest
	InsertIssueResponse
	InsertCommentRequest
	InsertCommentResponse
	IssuesListRequest
	ErrorMessage
	IssuesListResponse
	Update
	AtomPerson
*/
package monorail

import prpccommon "github.com/luci/luci-go/common/prpc"
import prpc "github.com/luci/luci-go/server/prpc"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type IssuesListRequest_CannedQuery int32

const (
	IssuesListRequest_ALL       IssuesListRequest_CannedQuery = 0
	IssuesListRequest_NEW       IssuesListRequest_CannedQuery = 1
	IssuesListRequest_OPEN      IssuesListRequest_CannedQuery = 2
	IssuesListRequest_OWNED     IssuesListRequest_CannedQuery = 3
	IssuesListRequest_REPORTED  IssuesListRequest_CannedQuery = 4
	IssuesListRequest_STARRED   IssuesListRequest_CannedQuery = 5
	IssuesListRequest_TO_VERIFY IssuesListRequest_CannedQuery = 6
)

var IssuesListRequest_CannedQuery_name = map[int32]string{
	0: "ALL",
	1: "NEW",
	2: "OPEN",
	3: "OWNED",
	4: "REPORTED",
	5: "STARRED",
	6: "TO_VERIFY",
}
var IssuesListRequest_CannedQuery_value = map[string]int32{
	"ALL":       0,
	"NEW":       1,
	"OPEN":      2,
	"OWNED":     3,
	"REPORTED":  4,
	"STARRED":   5,
	"TO_VERIFY": 6,
}

func (x IssuesListRequest_CannedQuery) String() string {
	return proto.EnumName(IssuesListRequest_CannedQuery_name, int32(x))
}
func (IssuesListRequest_CannedQuery) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{6, 0}
}

// A monorail issue.
type Issue struct {
	// Reporter of the issue.
	Author *AtomPerson `protobuf:"bytes,1,opt,name=author" json:"author,omitempty"`
	// Issues that must be fixed before this one can be fixed.
	BlockedOn []*IssueRef `protobuf:"bytes,2,rep,name=blockedOn" json:"blockedOn,omitempty"`
	// People participating in the issue discussion.
	Cc []*AtomPerson `protobuf:"bytes,6,rep,name=cc" json:"cc,omitempty"`
	// The text body of the issue.
	Description string `protobuf:"bytes,8,opt,name=description" json:"description,omitempty"`
	// Identifier of the issue, unique within the appengine app.
	Id int32 `protobuf:"varint,9,opt,name=id" json:"id,omitempty"`
	// Monorail components for this issue.
	Components []string `protobuf:"bytes,10,rep,name=components" json:"components,omitempty"`
	// Arbitrary indexed strings visible to users,
	// usually of form "Key-Value" or "Key-Value-SubValue",
	Labels []string `protobuf:"bytes,11,rep,name=labels" json:"labels,omitempty"`
	// Who is currently responsible for closing the issue.
	Owner *AtomPerson `protobuf:"bytes,12,opt,name=owner" json:"owner,omitempty"`
	// Current status of issue. Standard values:
	//
	// Open statuses:
	// "Unconrimed" - New, has been not verified or reproduced.
	// "Untriaged" - Confirmed, not reviews for priority of assignment
	// "Available" - Triaged, but no owner assigned
	// "Started" - Work in progress.
	// "ExternalDependency" - Requires action from a third party
	// Closed statuses:
	// "Fixed" - Work completed, needs verificaiton
	// "Verified" - Test or reporter verified that the fix works
	// "Duplicate" - Same root cause as another issue
	// "WontFix" -  Cannot reproduce, works as intended, invalid or absolete.
	// "Archived" - Old issue with no activity.
	Status string `protobuf:"bytes,17,opt,name=status" json:"status,omitempty"`
	// A one line description of the issue.
	Summary string `protobuf:"bytes,18,opt,name=summary" json:"summary,omitempty"`
}

func (m *Issue) Reset()                    { *m = Issue{} }
func (m *Issue) String() string            { return proto.CompactTextString(m) }
func (*Issue) ProtoMessage()               {}
func (*Issue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Issue) GetAuthor() *AtomPerson {
	if m != nil {
		return m.Author
	}
	return nil
}

func (m *Issue) GetBlockedOn() []*IssueRef {
	if m != nil {
		return m.BlockedOn
	}
	return nil
}

func (m *Issue) GetCc() []*AtomPerson {
	if m != nil {
		return m.Cc
	}
	return nil
}

func (m *Issue) GetOwner() *AtomPerson {
	if m != nil {
		return m.Owner
	}
	return nil
}

// IssueRef references another issue in the same Monorail instance.
type IssueRef struct {
	// ID of the issue.
	IssueId int32 `protobuf:"varint,1,opt,name=issueId" json:"issueId,omitempty"`
	// ID of the project containing the issue.
	ProjectId string `protobuf:"bytes,2,opt,name=projectId" json:"projectId,omitempty"`
}

func (m *IssueRef) Reset()                    { *m = IssueRef{} }
func (m *IssueRef) String() string            { return proto.CompactTextString(m) }
func (*IssueRef) ProtoMessage()               {}
func (*IssueRef) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// Request for Monorail.InsertIssue().
type InsertIssueRequest struct {
	// Target project id.
	ProjectId string `protobuf:"bytes,1,opt,name=projectId" json:"projectId,omitempty"`
	// Definition of the issue.
	// issue.id must be empty.
	Issue *Issue `protobuf:"bytes,2,opt,name=issue" json:"issue,omitempty"`
	// Whether to send email to participants.
	SendEmail bool `protobuf:"varint,3,opt,name=sendEmail" json:"sendEmail,omitempty"`
}

func (m *InsertIssueRequest) Reset()                    { *m = InsertIssueRequest{} }
func (m *InsertIssueRequest) String() string            { return proto.CompactTextString(m) }
func (*InsertIssueRequest) ProtoMessage()               {}
func (*InsertIssueRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *InsertIssueRequest) GetIssue() *Issue {
	if m != nil {
		return m.Issue
	}
	return nil
}

// Response for Monorail.InsertIssue()
type InsertIssueResponse struct {
	// Created issue.
	Issue *Issue `protobuf:"bytes,1,opt,name=issue" json:"issue,omitempty"`
}

func (m *InsertIssueResponse) Reset()                    { *m = InsertIssueResponse{} }
func (m *InsertIssueResponse) String() string            { return proto.CompactTextString(m) }
func (*InsertIssueResponse) ProtoMessage()               {}
func (*InsertIssueResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *InsertIssueResponse) GetIssue() *Issue {
	if m != nil {
		return m.Issue
	}
	return nil
}

// Request for Monorail.InsertComment()
type InsertCommentRequest struct {
	// Definition of the comment.
	Comment *InsertCommentRequest_Comment `protobuf:"bytes,1,opt,name=comment" json:"comment,omitempty"`
	// The reference to post the comment to.
	Issue *IssueRef `protobuf:"bytes,2,opt,name=issue" json:"issue,omitempty"`
}

func (m *InsertCommentRequest) Reset()                    { *m = InsertCommentRequest{} }
func (m *InsertCommentRequest) String() string            { return proto.CompactTextString(m) }
func (*InsertCommentRequest) ProtoMessage()               {}
func (*InsertCommentRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *InsertCommentRequest) GetComment() *InsertCommentRequest_Comment {
	if m != nil {
		return m.Comment
	}
	return nil
}

func (m *InsertCommentRequest) GetIssue() *IssueRef {
	if m != nil {
		return m.Issue
	}
	return nil
}

// Defines the comment.
// This message is partial.
// Derived from IssueCommentWrapper type in api_pb2_v1.py.
type InsertCommentRequest_Comment struct {
	Content string  `protobuf:"bytes,4,opt,name=content" json:"content,omitempty"`
	Updates *Update `protobuf:"bytes,8,opt,name=updates" json:"updates,omitempty"`
}

func (m *InsertCommentRequest_Comment) Reset()                    { *m = InsertCommentRequest_Comment{} }
func (m *InsertCommentRequest_Comment) String() string            { return proto.CompactTextString(m) }
func (*InsertCommentRequest_Comment) ProtoMessage()               {}
func (*InsertCommentRequest_Comment) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4, 0} }

func (m *InsertCommentRequest_Comment) GetUpdates() *Update {
	if m != nil {
		return m.Updates
	}
	return nil
}

type InsertCommentResponse struct {
}

func (m *InsertCommentResponse) Reset()                    { *m = InsertCommentResponse{} }
func (m *InsertCommentResponse) String() string            { return proto.CompactTextString(m) }
func (*InsertCommentResponse) ProtoMessage()               {}
func (*InsertCommentResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

// Request for a list of Issues.
type IssuesListRequest struct {
	// String name of the project, e.g. "chromium".
	ProjectId string `protobuf:"bytes,1,opt,name=projectId" json:"projectId,omitempty"`
	// Additional projects to search.
	AdditionalProject []string `protobuf:"bytes,2,rep,name=additionalProject" json:"additionalProject,omitempty"`
	// Use a canned query.
	Can IssuesListRequest_CannedQuery `protobuf:"varint,3,opt,name=can,enum=monorail.IssuesListRequest_CannedQuery" json:"can,omitempty"`
	// Issue label or space separated list of labels.
	Label string `protobuf:"bytes,4,opt,name=label" json:"label,omitempty"`
	// Maximum results to retrieve.
	MaxResults int32 `protobuf:"varint,5,opt,name=maxResults" json:"maxResults,omitempty"`
	// Issue owner.
	Owner string `protobuf:"bytes,6,opt,name=owner" json:"owner,omitempty"`
	// Search for Issues published before this timestamp.
	PublishedMax int64 `protobuf:"varint,7,opt,name=publishedMax" json:"publishedMax,omitempty"`
	// Search for Issues published after this timestamp.
	PublishedMin int64 `protobuf:"varint,8,opt,name=publishedMin" json:"publishedMin,omitempty"`
	// Free-form text query.
	Q string `protobuf:"bytes,9,opt,name=q" json:"q,omitempty"`
	// Sort-by field or fields, space separated terms with leading - to
	// indicate decreasing direction. e.g. "estdays -milestone" to sort by
	// estdays increasing, then milestone decreasing.
	Sort string `protobuf:"bytes,10,opt,name=sort" json:"sort,omitempty"`
	// Starting index for pagination.
	StartIndex int32 `protobuf:"varint,11,opt,name=startIndex" json:"startIndex,omitempty"`
	// Issue status.
	Status string `protobuf:"bytes,12,opt,name=status" json:"status,omitempty"`
	// Search for Issues most recently updated before this timestamp.
	UpdatedMax int64 `protobuf:"varint,13,opt,name=updatedMax" json:"updatedMax,omitempty"`
	// Search for Issues most recently updated after this timestamp.
	UpdatedMin int64 `protobuf:"varint,14,opt,name=updatedMin" json:"updatedMin,omitempty"`
}

func (m *IssuesListRequest) Reset()                    { *m = IssuesListRequest{} }
func (m *IssuesListRequest) String() string            { return proto.CompactTextString(m) }
func (*IssuesListRequest) ProtoMessage()               {}
func (*IssuesListRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type ErrorMessage struct {
	Code    int32  `protobuf:"varint,1,opt,name=code" json:"code,omitempty"`
	Reason  string `protobuf:"bytes,2,opt,name=reason" json:"reason,omitempty"`
	Message string `protobuf:"bytes,3,opt,name=message" json:"message,omitempty"`
}

func (m *ErrorMessage) Reset()                    { *m = ErrorMessage{} }
func (m *ErrorMessage) String() string            { return proto.CompactTextString(m) }
func (*ErrorMessage) ProtoMessage()               {}
func (*ErrorMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type IssuesListResponse struct {
	Error *ErrorMessage `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
	// Search results.
	Items []*Issue `protobuf:"bytes,2,rep,name=items" json:"items,omitempty"`
	// Total size of matching result set, regardless of how many are included
	// in this response.
	TotalResults int32 `protobuf:"varint,3,opt,name=totalResults" json:"totalResults,omitempty"`
}

func (m *IssuesListResponse) Reset()                    { *m = IssuesListResponse{} }
func (m *IssuesListResponse) String() string            { return proto.CompactTextString(m) }
func (*IssuesListResponse) ProtoMessage()               {}
func (*IssuesListResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *IssuesListResponse) GetError() *ErrorMessage {
	if m != nil {
		return m.Error
	}
	return nil
}

func (m *IssuesListResponse) GetItems() []*Issue {
	if m != nil {
		return m.Items
	}
	return nil
}

// Defines a mutation to an issue.
// This message is partial.
// Derived from Update type in api_pb2_v1.py.
type Update struct {
	// If set, the new status of the issue.
	Status string `protobuf:"bytes,2,opt,name=status" json:"status,omitempty"`
}

func (m *Update) Reset()                    { *m = Update{} }
func (m *Update) String() string            { return proto.CompactTextString(m) }
func (*Update) ProtoMessage()               {}
func (*Update) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

// Identifies a Monorail user.
type AtomPerson struct {
	// User email.
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *AtomPerson) Reset()                    { *m = AtomPerson{} }
func (m *AtomPerson) String() string            { return proto.CompactTextString(m) }
func (*AtomPerson) ProtoMessage()               {}
func (*AtomPerson) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func init() {
	proto.RegisterType((*Issue)(nil), "monorail.Issue")
	proto.RegisterType((*IssueRef)(nil), "monorail.IssueRef")
	proto.RegisterType((*InsertIssueRequest)(nil), "monorail.InsertIssueRequest")
	proto.RegisterType((*InsertIssueResponse)(nil), "monorail.InsertIssueResponse")
	proto.RegisterType((*InsertCommentRequest)(nil), "monorail.InsertCommentRequest")
	proto.RegisterType((*InsertCommentRequest_Comment)(nil), "monorail.InsertCommentRequest.Comment")
	proto.RegisterType((*InsertCommentResponse)(nil), "monorail.InsertCommentResponse")
	proto.RegisterType((*IssuesListRequest)(nil), "monorail.IssuesListRequest")
	proto.RegisterType((*ErrorMessage)(nil), "monorail.ErrorMessage")
	proto.RegisterType((*IssuesListResponse)(nil), "monorail.IssuesListResponse")
	proto.RegisterType((*Update)(nil), "monorail.Update")
	proto.RegisterType((*AtomPerson)(nil), "monorail.AtomPerson")
	proto.RegisterEnum("monorail.IssuesListRequest_CannedQuery", IssuesListRequest_CannedQuery_name, IssuesListRequest_CannedQuery_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for Monorail service

type MonorailClient interface {
	// Creates an issue.
	InsertIssue(ctx context.Context, in *InsertIssueRequest, opts ...grpc.CallOption) (*InsertIssueResponse, error)
	// Posts a comment to an issue. Can update issue attributes, such as status.
	InsertComment(ctx context.Context, in *InsertCommentRequest, opts ...grpc.CallOption) (*InsertCommentResponse, error)
	// Lists issues from a project.
	IssuesList(ctx context.Context, in *IssuesListRequest, opts ...grpc.CallOption) (*IssuesListResponse, error)
}
type monorailPRPCClient struct {
	client *prpccommon.Client
}

func NewMonorailPRPCClient(client *prpccommon.Client) MonorailClient {
	return &monorailPRPCClient{client}
}

func (c *monorailPRPCClient) InsertIssue(ctx context.Context, in *InsertIssueRequest, opts ...grpc.CallOption) (*InsertIssueResponse, error) {
	out := new(InsertIssueResponse)
	err := c.client.Call(ctx, "monorail.Monorail", "InsertIssue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monorailPRPCClient) InsertComment(ctx context.Context, in *InsertCommentRequest, opts ...grpc.CallOption) (*InsertCommentResponse, error) {
	out := new(InsertCommentResponse)
	err := c.client.Call(ctx, "monorail.Monorail", "InsertComment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monorailPRPCClient) IssuesList(ctx context.Context, in *IssuesListRequest, opts ...grpc.CallOption) (*IssuesListResponse, error) {
	out := new(IssuesListResponse)
	err := c.client.Call(ctx, "monorail.Monorail", "IssuesList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type monorailClient struct {
	cc *grpc.ClientConn
}

func NewMonorailClient(cc *grpc.ClientConn) MonorailClient {
	return &monorailClient{cc}
}

func (c *monorailClient) InsertIssue(ctx context.Context, in *InsertIssueRequest, opts ...grpc.CallOption) (*InsertIssueResponse, error) {
	out := new(InsertIssueResponse)
	err := grpc.Invoke(ctx, "/monorail.Monorail/InsertIssue", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monorailClient) InsertComment(ctx context.Context, in *InsertCommentRequest, opts ...grpc.CallOption) (*InsertCommentResponse, error) {
	out := new(InsertCommentResponse)
	err := grpc.Invoke(ctx, "/monorail.Monorail/InsertComment", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monorailClient) IssuesList(ctx context.Context, in *IssuesListRequest, opts ...grpc.CallOption) (*IssuesListResponse, error) {
	out := new(IssuesListResponse)
	err := grpc.Invoke(ctx, "/monorail.Monorail/IssuesList", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Monorail service

type MonorailServer interface {
	// Creates an issue.
	InsertIssue(context.Context, *InsertIssueRequest) (*InsertIssueResponse, error)
	// Posts a comment to an issue. Can update issue attributes, such as status.
	InsertComment(context.Context, *InsertCommentRequest) (*InsertCommentResponse, error)
	// Lists issues from a project.
	IssuesList(context.Context, *IssuesListRequest) (*IssuesListResponse, error)
}

func RegisterMonorailServer(s prpc.Registrar, srv MonorailServer) {
	s.RegisterService(&_Monorail_serviceDesc, srv)
}

func _Monorail_InsertIssue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InsertIssueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonorailServer).InsertIssue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monorail.Monorail/InsertIssue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonorailServer).InsertIssue(ctx, req.(*InsertIssueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Monorail_InsertComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InsertCommentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonorailServer).InsertComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monorail.Monorail/InsertComment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonorailServer).InsertComment(ctx, req.(*InsertCommentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Monorail_IssuesList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IssuesListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonorailServer).IssuesList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monorail.Monorail/IssuesList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonorailServer).IssuesList(ctx, req.(*IssuesListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Monorail_serviceDesc = grpc.ServiceDesc{
	ServiceName: "monorail.Monorail",
	HandlerType: (*MonorailServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InsertIssue",
			Handler:    _Monorail_InsertIssue_Handler,
		},
		{
			MethodName: "InsertComment",
			Handler:    _Monorail_InsertComment_Handler,
		},
		{
			MethodName: "IssuesList",
			Handler:    _Monorail_IssuesList_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("monorail.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 857 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x55, 0x5f, 0x6f, 0xe3, 0x44,
	0x10, 0x3f, 0x27, 0x71, 0x12, 0x8f, 0xd3, 0x92, 0x2e, 0xe5, 0xb0, 0x4a, 0x39, 0x2a, 0x8b, 0x3f,
	0x27, 0x54, 0x55, 0x28, 0x3c, 0x21, 0xf1, 0x40, 0x39, 0x82, 0x14, 0xa9, 0x6d, 0xca, 0x12, 0x38,
	0xf1, 0x02, 0xda, 0xd8, 0x0b, 0x67, 0xb0, 0x77, 0x53, 0xef, 0x5a, 0xd7, 0xfb, 0x10, 0x7c, 0x0f,
	0xbe, 0x0f, 0xe2, 0xb3, 0xf0, 0xca, 0xfe, 0xf3, 0xd9, 0x0e, 0x69, 0x74, 0x6f, 0x9e, 0xf9, 0xcd,
	0xfc, 0x76, 0x76, 0x67, 0x7e, 0x63, 0x38, 0x2c, 0x38, 0xe3, 0x25, 0xc9, 0xf2, 0x8b, 0x4d, 0xc9,
	0x25, 0x47, 0xe3, 0xda, 0x8e, 0xff, 0xe9, 0x81, 0xbf, 0x10, 0xa2, 0xa2, 0xe8, 0x1c, 0x86, 0xa4,
	0x92, 0x2f, 0x78, 0x19, 0x79, 0x67, 0xde, 0xd3, 0x70, 0x76, 0x7c, 0xf1, 0x3a, 0xe9, 0x52, 0xf2,
	0xe2, 0x96, 0x96, 0x82, 0x33, 0xec, 0x62, 0xd0, 0x67, 0x10, 0xac, 0x73, 0x9e, 0xfc, 0x41, 0xd3,
	0x25, 0x8b, 0x7a, 0x67, 0x7d, 0x95, 0x80, 0x9a, 0x04, 0xc3, 0x88, 0xe9, 0xaf, 0xb8, 0x09, 0x42,
	0x1f, 0x42, 0x2f, 0x49, 0xa2, 0xa1, 0x09, 0xdd, 0xcd, 0xad, 0x70, 0x74, 0x06, 0x61, 0x4a, 0x45,
	0x52, 0x66, 0x1b, 0x99, 0x71, 0x16, 0x8d, 0x55, 0x29, 0x01, 0x6e, 0xbb, 0xd0, 0x21, 0xf4, 0xb2,
	0x34, 0x0a, 0x14, 0xe0, 0x63, 0xf5, 0x85, 0x9e, 0x00, 0x24, 0xbc, 0xd8, 0x70, 0x46, 0x99, 0x14,
	0x11, 0x28, 0xfe, 0x00, 0xb7, 0x3c, 0xe8, 0x31, 0x0c, 0x73, 0xb2, 0xa6, 0xb9, 0x88, 0x42, 0x83,
	0x39, 0x0b, 0x7d, 0x0a, 0x3e, 0x7f, 0xc9, 0x68, 0x19, 0x4d, 0xf6, 0x5c, 0xd7, 0x86, 0x68, 0x0e,
	0x21, 0x89, 0xac, 0x44, 0x74, 0x64, 0x0a, 0x72, 0x16, 0x8a, 0x60, 0x24, 0xaa, 0xa2, 0x20, 0xe5,
	0xab, 0x08, 0x19, 0xa0, 0x36, 0xe3, 0xaf, 0x61, 0x5c, 0x3f, 0x82, 0x8e, 0xca, 0xf4, 0xf7, 0x22,
	0x35, 0x4f, 0xeb, 0xe3, 0xda, 0x44, 0xa7, 0x10, 0xa8, 0x86, 0xfc, 0x4e, 0x13, 0xa9, 0xb0, 0x9e,
	0x61, 0x68, 0x1c, 0xf1, 0x4b, 0x40, 0x0b, 0x26, 0x68, 0x29, 0x1d, 0xd3, 0x5d, 0x45, 0x85, 0xec,
	0xe6, 0x78, 0x5b, 0x39, 0xe8, 0x23, 0xf0, 0x0d, 0xb9, 0x61, 0x0b, 0x67, 0x6f, 0x6d, 0xf7, 0xc4,
	0xa2, 0x9a, 0x44, 0x50, 0x96, 0xce, 0x0b, 0x85, 0x44, 0x7d, 0x15, 0x3a, 0xc6, 0x8d, 0x23, 0xfe,
	0x12, 0xde, 0xee, 0x1c, 0x2c, 0xd4, 0x5b, 0x0a, 0xda, 0x70, 0x7b, 0xfb, 0xb8, 0xe3, 0xbf, 0x3d,
	0x38, 0xb6, 0xe9, 0xcf, 0x78, 0x51, 0xa8, 0x1e, 0xd4, 0x95, 0x7f, 0x05, 0xa3, 0xc4, 0x7a, 0x1c,
	0xc3, 0xc7, 0x2d, 0x86, 0x1d, 0x09, 0x17, 0xb5, 0x59, 0xa7, 0xa1, 0xa7, 0xdd, 0xdb, 0xed, 0x9a,
	0x38, 0x1b, 0x70, 0xb2, 0x84, 0x91, 0xcb, 0xd6, 0xcf, 0x9f, 0x70, 0x26, 0xf5, 0xb1, 0x03, 0xdb,
	0x24, 0x67, 0xaa, 0x11, 0x18, 0x55, 0x9b, 0x94, 0x48, 0x2a, 0xcc, 0xa0, 0x85, 0xb3, 0x69, 0x43,
	0xf8, 0x83, 0x01, 0x70, 0x1d, 0x10, 0xbf, 0x0b, 0xef, 0x6c, 0xd5, 0x68, 0x5f, 0x25, 0xfe, 0x6b,
	0x00, 0x47, 0xe6, 0x74, 0x71, 0x95, 0x09, 0xf9, 0x66, 0x5d, 0x3a, 0x87, 0x23, 0x92, 0xa6, 0x99,
	0x9e, 0x67, 0x92, 0xdf, 0x5a, 0xb7, 0x51, 0x51, 0x80, 0xff, 0x0f, 0xa0, 0x2f, 0xa0, 0x9f, 0x10,
	0x66, 0xda, 0x74, 0x38, 0xfb, 0x64, 0xeb, 0xce, 0xed, 0x53, 0x2f, 0x9e, 0x11, 0xc6, 0x68, 0xfa,
	0x5d, 0x45, 0xcb, 0x57, 0x58, 0xe7, 0xa0, 0x63, 0xf0, 0xcd, 0xb8, 0xbb, 0x9b, 0x5b, 0x43, 0x4b,
	0xa6, 0x20, 0xf7, 0xea, 0x06, 0x55, 0xae, 0x24, 0xe3, 0x9b, 0x99, 0x6c, 0x79, 0x74, 0x96, 0x95,
	0xc6, 0xd0, 0x66, 0x59, 0x11, 0xc4, 0x30, 0xd9, 0x54, 0xeb, 0x3c, 0x13, 0x2f, 0x68, 0x7a, 0x4d,
	0xee, 0xa3, 0x91, 0x02, 0xfb, 0xb8, 0xe3, 0xeb, 0xc6, 0x64, 0x56, 0xbf, 0x9d, 0x98, 0x8c, 0xa1,
	0x09, 0x78, 0x77, 0x46, 0xbf, 0x01, 0xf6, 0xee, 0x10, 0x82, 0x81, 0xe0, 0xa5, 0x54, 0xc2, 0xd5,
	0x0e, 0xf3, 0xad, 0xeb, 0x53, 0x02, 0x53, 0xe3, 0xc7, 0x52, 0x7a, 0xaf, 0x64, 0x6b, 0xea, 0x6b,
	0x3c, 0x2d, 0x39, 0x4e, 0x3a, 0x72, 0x54, 0x79, 0xb6, 0x5d, 0xa6, 0xbe, 0x03, 0x73, 0x76, 0xcb,
	0xd3, 0xc6, 0x55, 0x6d, 0x87, 0x5d, 0x3c, 0x63, 0xf1, 0xcf, 0x10, 0xb6, 0x5e, 0x10, 0x8d, 0xa0,
	0x7f, 0x79, 0x75, 0x35, 0x7d, 0xa4, 0x3f, 0x6e, 0xe6, 0xcf, 0xa7, 0x1e, 0x1a, 0xc3, 0x60, 0x79,
	0x3b, 0xbf, 0x99, 0xf6, 0x50, 0x00, 0xfe, 0xf2, 0xf9, 0xcd, 0xfc, 0x9b, 0x69, 0x5f, 0xdd, 0x67,
	0x8c, 0xe7, 0xb7, 0x4b, 0xbc, 0x52, 0xd6, 0x00, 0x85, 0x30, 0xfa, 0x7e, 0x75, 0x89, 0xb1, 0x32,
	0x7c, 0x74, 0x00, 0xc1, 0x6a, 0xf9, 0xcb, 0x8f, 0x73, 0xbc, 0xf8, 0xf6, 0xa7, 0xe9, 0x30, 0x5e,
	0xc1, 0x64, 0x5e, 0x96, 0xbc, 0xbc, 0xa6, 0x42, 0x90, 0xdf, 0xa8, 0xbe, 0x7b, 0xc2, 0x53, 0xea,
	0xb6, 0x82, 0xf9, 0xd6, 0x77, 0x2b, 0x29, 0x51, 0xbb, 0xc7, 0xed, 0x03, 0x67, 0xe9, 0x29, 0x2e,
	0x6c, 0x9a, 0x19, 0x04, 0x35, 0xc5, 0xce, 0x8c, 0xff, 0xf4, 0xd4, 0x9e, 0x68, 0x8d, 0x82, 0x53,
	0xeb, 0x39, 0xf8, 0x54, 0x1f, 0xe6, 0xb4, 0xf6, 0xb8, 0x99, 0x9b, 0x76, 0x0d, 0xd8, 0x06, 0x19,
	0x6d, 0x4b, 0x5a, 0x08, 0xb7, 0xcb, 0x77, 0x68, 0x5b, 0xa3, 0xba, 0xbf, 0x92, 0x4b, 0x92, 0xd7,
	0xb3, 0xd3, 0x37, 0x95, 0x77, 0x7c, 0xf1, 0x19, 0x0c, 0xad, 0x78, 0x5a, 0x7d, 0xea, 0xb5, 0xfb,
	0xa4, 0x22, 0xa0, 0xd9, 0xb1, 0xfa, 0x15, 0x18, 0x29, 0xa8, 0x53, 0x89, 0xf9, 0x9e, 0xfd, 0xeb,
	0xc1, 0xf8, 0xda, 0x55, 0x80, 0xae, 0x20, 0x6c, 0xad, 0x23, 0x74, 0xba, 0xbd, 0x35, 0xda, 0xeb,
	0xf1, 0xe4, 0xfd, 0x07, 0x50, 0xa7, 0xd6, 0x47, 0x08, 0xc3, 0x41, 0x47, 0xc8, 0xe8, 0xc9, 0xfe,
	0x2d, 0x74, 0xf2, 0xc1, 0x83, 0xf8, 0x6b, 0xce, 0x05, 0x40, 0xd3, 0x01, 0xf4, 0xde, 0x1e, 0x89,
	0x9e, 0x9c, 0xee, 0x06, 0x6b, 0xaa, 0xf5, 0xd0, 0xfc, 0xa1, 0x3f, 0xff, 0x2f, 0x00, 0x00, 0xff,
	0xff, 0xf0, 0x43, 0x48, 0x3f, 0xb3, 0x07, 0x00, 0x00,
}
