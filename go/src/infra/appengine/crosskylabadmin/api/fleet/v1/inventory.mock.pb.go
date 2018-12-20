// Code generated by MockGen. DO NOT EDIT.
// Source: inventory.pb.go

// Package fleet is a generated GoMock package.
package fleet

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockInventoryClient is a mock of InventoryClient interface
type MockInventoryClient struct {
	ctrl     *gomock.Controller
	recorder *MockInventoryClientMockRecorder
}

// MockInventoryClientMockRecorder is the mock recorder for MockInventoryClient
type MockInventoryClientMockRecorder struct {
	mock *MockInventoryClient
}

// NewMockInventoryClient creates a new mock instance
func NewMockInventoryClient(ctrl *gomock.Controller) *MockInventoryClient {
	mock := &MockInventoryClient{ctrl: ctrl}
	mock.recorder = &MockInventoryClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInventoryClient) EXPECT() *MockInventoryClientMockRecorder {
	return m.recorder
}

// EnsurePoolHealthy mocks base method
func (m *MockInventoryClient) EnsurePoolHealthy(ctx context.Context, in *EnsurePoolHealthyRequest, opts ...grpc.CallOption) (*EnsurePoolHealthyResponse, error) {
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "EnsurePoolHealthy", varargs...)
	ret0, _ := ret[0].(*EnsurePoolHealthyResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsurePoolHealthy indicates an expected call of EnsurePoolHealthy
func (mr *MockInventoryClientMockRecorder) EnsurePoolHealthy(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsurePoolHealthy", reflect.TypeOf((*MockInventoryClient)(nil).EnsurePoolHealthy), varargs...)
}

// ResizePool mocks base method
func (m *MockInventoryClient) ResizePool(ctx context.Context, in *ResizePoolRequest, opts ...grpc.CallOption) (*ResizePoolResponse, error) {
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ResizePool", varargs...)
	ret0, _ := ret[0].(*ResizePoolResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ResizePool indicates an expected call of ResizePool
func (mr *MockInventoryClientMockRecorder) ResizePool(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResizePool", reflect.TypeOf((*MockInventoryClient)(nil).ResizePool), varargs...)
}

// DeactivateDut mocks base method
func (m *MockInventoryClient) DeactivateDut(ctx context.Context, in *DeactivateDutRequest, opts ...grpc.CallOption) (*DeactivateDutResponse, error) {
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeactivateDut", varargs...)
	ret0, _ := ret[0].(*DeactivateDutResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeactivateDut indicates an expected call of DeactivateDut
func (mr *MockInventoryClientMockRecorder) DeactivateDut(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeactivateDut", reflect.TypeOf((*MockInventoryClient)(nil).DeactivateDut), varargs...)
}

// ActivateDut mocks base method
func (m *MockInventoryClient) ActivateDut(ctx context.Context, in *ActivateDutRequest, opts ...grpc.CallOption) (*ActivateDutResponse, error) {
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ActivateDut", varargs...)
	ret0, _ := ret[0].(*ActivateDutResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ActivateDut indicates an expected call of ActivateDut
func (mr *MockInventoryClientMockRecorder) ActivateDut(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ActivateDut", reflect.TypeOf((*MockInventoryClient)(nil).ActivateDut), varargs...)
}

// MockInventoryServer is a mock of InventoryServer interface
type MockInventoryServer struct {
	ctrl     *gomock.Controller
	recorder *MockInventoryServerMockRecorder
}

// MockInventoryServerMockRecorder is the mock recorder for MockInventoryServer
type MockInventoryServerMockRecorder struct {
	mock *MockInventoryServer
}

// NewMockInventoryServer creates a new mock instance
func NewMockInventoryServer(ctrl *gomock.Controller) *MockInventoryServer {
	mock := &MockInventoryServer{ctrl: ctrl}
	mock.recorder = &MockInventoryServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInventoryServer) EXPECT() *MockInventoryServerMockRecorder {
	return m.recorder
}

// EnsurePoolHealthy mocks base method
func (m *MockInventoryServer) EnsurePoolHealthy(arg0 context.Context, arg1 *EnsurePoolHealthyRequest) (*EnsurePoolHealthyResponse, error) {
	ret := m.ctrl.Call(m, "EnsurePoolHealthy", arg0, arg1)
	ret0, _ := ret[0].(*EnsurePoolHealthyResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsurePoolHealthy indicates an expected call of EnsurePoolHealthy
func (mr *MockInventoryServerMockRecorder) EnsurePoolHealthy(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsurePoolHealthy", reflect.TypeOf((*MockInventoryServer)(nil).EnsurePoolHealthy), arg0, arg1)
}

// ResizePool mocks base method
func (m *MockInventoryServer) ResizePool(arg0 context.Context, arg1 *ResizePoolRequest) (*ResizePoolResponse, error) {
	ret := m.ctrl.Call(m, "ResizePool", arg0, arg1)
	ret0, _ := ret[0].(*ResizePoolResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ResizePool indicates an expected call of ResizePool
func (mr *MockInventoryServerMockRecorder) ResizePool(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResizePool", reflect.TypeOf((*MockInventoryServer)(nil).ResizePool), arg0, arg1)
}

// DeactivateDut mocks base method
func (m *MockInventoryServer) DeactivateDut(arg0 context.Context, arg1 *DeactivateDutRequest) (*DeactivateDutResponse, error) {
	ret := m.ctrl.Call(m, "DeactivateDut", arg0, arg1)
	ret0, _ := ret[0].(*DeactivateDutResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeactivateDut indicates an expected call of DeactivateDut
func (mr *MockInventoryServerMockRecorder) DeactivateDut(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeactivateDut", reflect.TypeOf((*MockInventoryServer)(nil).DeactivateDut), arg0, arg1)
}

// ActivateDut mocks base method
func (m *MockInventoryServer) ActivateDut(arg0 context.Context, arg1 *ActivateDutRequest) (*ActivateDutResponse, error) {
	ret := m.ctrl.Call(m, "ActivateDut", arg0, arg1)
	ret0, _ := ret[0].(*ActivateDutResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ActivateDut indicates an expected call of ActivateDut
func (mr *MockInventoryServerMockRecorder) ActivateDut(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ActivateDut", reflect.TypeOf((*MockInventoryServer)(nil).ActivateDut), arg0, arg1)
}
