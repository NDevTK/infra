// Code generated by MockGen. DO NOT EDIT.
// Source: ufs.go

package clients

import (
	context "context"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufspb0 "infra/unifiedfleet/api/v1/rpc"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockUFSClient is a mock of UFSClient interface.
type MockUFSClient struct {
	ctrl     *gomock.Controller
	recorder *MockUFSClientMockRecorder
}

// MockUFSClientMockRecorder is the mock recorder for MockUFSClient.
type MockUFSClientMockRecorder struct {
	mock *MockUFSClient
}

// NewMockUFSClient creates a new mock instance.
func NewMockUFSClient(ctrl *gomock.Controller) *MockUFSClient {
	mock := &MockUFSClient{ctrl: ctrl}
	mock.recorder = &MockUFSClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUFSClient) EXPECT() *MockUFSClientMockRecorder {
	return m.recorder
}

// ListMachineLSEs mocks base method.
func (m *MockUFSClient) ListMachineLSEs(ctx context.Context, in *ufspb0.ListMachineLSEsRequest, opts ...grpc.CallOption) (*ufspb0.ListMachineLSEsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListMachineLSEs", varargs...)
	ret0, _ := ret[0].(*ufspb0.ListMachineLSEsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMachineLSEs indicates an expected call of ListMachineLSEs.
func (mr *MockUFSClientMockRecorder) ListMachineLSEs(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMachineLSEs", reflect.TypeOf((*MockUFSClient)(nil).ListMachineLSEs), varargs...)
}

// ListMachines mocks base method.
func (m *MockUFSClient) ListMachines(ctx context.Context, in *ufspb0.ListMachinesRequest, opts ...grpc.CallOption) (*ufspb0.ListMachinesResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListMachines", varargs...)
	ret0, _ := ret[0].(*ufspb0.ListMachinesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMachines indicates an expected call of ListMachines.
func (mr *MockUFSClientMockRecorder) ListMachines(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMachines", reflect.TypeOf((*MockUFSClient)(nil).ListMachines), varargs...)
}

// ListSchedulingUnits mocks base method.
func (m *MockUFSClient) ListSchedulingUnits(ctx context.Context, in *ufspb0.ListSchedulingUnitsRequest, opts ...grpc.CallOption) (*ufspb0.ListSchedulingUnitsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListSchedulingUnits", varargs...)
	ret0, _ := ret[0].(*ufspb0.ListSchedulingUnitsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSchedulingUnits indicates an expected call of ListSchedulingUnits.
func (mr *MockUFSClientMockRecorder) ListSchedulingUnits(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSchedulingUnits", reflect.TypeOf((*MockUFSClient)(nil).ListSchedulingUnits), varargs...)
}

// UpdateMachineLSE mocks base method.
func (m *MockUFSClient) UpdateMachineLSE(ctx context.Context, in *ufspb0.UpdateMachineLSERequest, opts ...grpc.CallOption) (*ufspb.MachineLSE, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateMachineLSE", varargs...)
	ret0, _ := ret[0].(*ufspb.MachineLSE)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateMachineLSE indicates an expected call of UpdateMachineLSE.
func (mr *MockUFSClientMockRecorder) UpdateMachineLSE(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateMachineLSE", reflect.TypeOf((*MockUFSClient)(nil).UpdateMachineLSE), varargs...)
}
