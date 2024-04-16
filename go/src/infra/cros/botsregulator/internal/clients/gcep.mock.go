// Code generated by MockGen. DO NOT EDIT.
// Source: gcep.go

package clients

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	config "go.chromium.org/luci/gce/api/config/v1"
	grpc "google.golang.org/grpc"
)

// MockGCEPClient is a mock of GCEPClient interface.
type MockGCEPClient struct {
	ctrl     *gomock.Controller
	recorder *MockGCEPClientMockRecorder
}

// MockGCEPClientMockRecorder is the mock recorder for MockGCEPClient.
type MockGCEPClientMockRecorder struct {
	mock *MockGCEPClient
}

// NewMockGCEPClient creates a new mock instance.
func NewMockGCEPClient(ctrl *gomock.Controller) *MockGCEPClient {
	mock := &MockGCEPClient{ctrl: ctrl}
	mock.recorder = &MockGCEPClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGCEPClient) EXPECT() *MockGCEPClientMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockGCEPClient) Get(ctx context.Context, in *config.GetRequest, opts ...grpc.CallOption) (*config.Config, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Get", varargs...)
	ret0, _ := ret[0].(*config.Config)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockGCEPClientMockRecorder) Get(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockGCEPClient)(nil).Get), varargs...)
}

// Update mocks base method.
func (m *MockGCEPClient) Update(ctx context.Context, in *config.UpdateRequest, opts ...grpc.CallOption) (*config.Config, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(*config.Config)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockGCEPClientMockRecorder) Update(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockGCEPClient)(nil).Update), varargs...)
}