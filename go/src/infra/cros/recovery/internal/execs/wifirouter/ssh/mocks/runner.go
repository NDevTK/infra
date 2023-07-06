// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by MockGen. DO NOT EDIT.
// Source: internal/execs/wifirouter/ssh/runner.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	ssh "infra/cros/recovery/internal/execs/wifirouter/ssh"
	tlw "infra/cros/recovery/tlw"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockAccess is a mock of Access interface.
type MockAccess struct {
	ctrl     *gomock.Controller
	recorder *MockAccessMockRecorder
}

// MockAccessMockRecorder is the mock recorder for MockAccess.
type MockAccessMockRecorder struct {
	mock *MockAccess
}

// NewMockAccess creates a new mock instance.
func NewMockAccess(ctrl *gomock.Controller) *MockAccess {
	mock := &MockAccess{ctrl: ctrl}
	mock.recorder = &MockAccessMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccess) EXPECT() *MockAccessMockRecorder {
	return m.recorder
}

// Run mocks base method.
func (m *MockAccess) Run(ctx context.Context, request *tlw.RunRequest) *tlw.RunResult {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", ctx, request)
	ret0, _ := ret[0].(*tlw.RunResult)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockAccessMockRecorder) Run(ctx, request interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockAccess)(nil).Run), ctx, request)
}

// MockRunResult is a mock of RunResult interface.
type MockRunResult struct {
	ctrl     *gomock.Controller
	recorder *MockRunResultMockRecorder
}

// MockRunResultMockRecorder is the mock recorder for MockRunResult.
type MockRunResultMockRecorder struct {
	mock *MockRunResult
}

// NewMockRunResult creates a new mock instance.
func NewMockRunResult(ctrl *gomock.Controller) *MockRunResult {
	mock := &MockRunResult{ctrl: ctrl}
	mock.recorder = &MockRunResultMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRunResult) EXPECT() *MockRunResultMockRecorder {
	return m.recorder
}

// GetCommand mocks base method.
func (m *MockRunResult) GetCommand() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCommand")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetCommand indicates an expected call of GetCommand.
func (mr *MockRunResultMockRecorder) GetCommand() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCommand", reflect.TypeOf((*MockRunResult)(nil).GetCommand))
}

// GetExitCode mocks base method.
func (m *MockRunResult) GetExitCode() int32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExitCode")
	ret0, _ := ret[0].(int32)
	return ret0
}

// GetExitCode indicates an expected call of GetExitCode.
func (mr *MockRunResultMockRecorder) GetExitCode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExitCode", reflect.TypeOf((*MockRunResult)(nil).GetExitCode))
}

// GetStderr mocks base method.
func (m *MockRunResult) GetStderr() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStderr")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetStderr indicates an expected call of GetStderr.
func (mr *MockRunResultMockRecorder) GetStderr() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStderr", reflect.TypeOf((*MockRunResult)(nil).GetStderr))
}

// GetStdout mocks base method.
func (m *MockRunResult) GetStdout() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStdout")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetStdout indicates an expected call of GetStdout.
func (mr *MockRunResultMockRecorder) GetStdout() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStdout", reflect.TypeOf((*MockRunResult)(nil).GetStdout))
}

// MockRunner is a mock of Runner interface.
type MockRunner struct {
	ctrl     *gomock.Controller
	recorder *MockRunnerMockRecorder
}

// MockRunnerMockRecorder is the mock recorder for MockRunner.
type MockRunnerMockRecorder struct {
	mock *MockRunner
}

// NewMockRunner creates a new mock instance.
func NewMockRunner(ctrl *gomock.Controller) *MockRunner {
	mock := &MockRunner{ctrl: ctrl}
	mock.recorder = &MockRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRunner) EXPECT() *MockRunnerMockRecorder {
	return m.recorder
}

// Run mocks base method.
func (m *MockRunner) Run(ctx context.Context, timeout time.Duration, cmd string, args ...string) (string, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, timeout, cmd}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Run", varargs...)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Run indicates an expected call of Run.
func (mr *MockRunnerMockRecorder) Run(ctx, timeout, cmd interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, timeout, cmd}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockRunner)(nil).Run), varargs...)
}

// RunForResult mocks base method.
func (m *MockRunner) RunForResult(ctx context.Context, timeout time.Duration, inBackground bool, cmd string, args ...string) ssh.RunResult {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, timeout, inBackground, cmd}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RunForResult", varargs...)
	ret0, _ := ret[0].(ssh.RunResult)
	return ret0
}

// RunForResult indicates an expected call of RunForResult.
func (mr *MockRunnerMockRecorder) RunForResult(ctx, timeout, inBackground, cmd interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, timeout, inBackground, cmd}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunForResult", reflect.TypeOf((*MockRunner)(nil).RunForResult), varargs...)
}

// RunInBackground mocks base method.
func (m *MockRunner) RunInBackground(ctx context.Context, timeout time.Duration, cmd string, args ...string) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, timeout, cmd}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "RunInBackground", varargs...)
}

// RunInBackground indicates an expected call of RunInBackground.
func (mr *MockRunnerMockRecorder) RunInBackground(ctx, timeout, cmd interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, timeout, cmd}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunInBackground", reflect.TypeOf((*MockRunner)(nil).RunInBackground), varargs...)
}
