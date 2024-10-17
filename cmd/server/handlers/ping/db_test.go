// Code generated by MockGen. DO NOT EDIT.
// Source: cmd/server/handlers/ping/handler.go

// Package ping is a generated GoMock package.
package ping

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockIDB is a mock of IDB interface.
type MockIDB struct {
	ctrl     *gomock.Controller
	recorder *MockIDBMockRecorder
}

// MockIDBMockRecorder is the mock recorder for MockIDB.
type MockIDBMockRecorder struct {
	mock *MockIDB
}

// NewMockIDB creates a new mock instance.
func NewMockIDB(ctrl *gomock.Controller) *MockIDB {
	mock := &MockIDB{ctrl: ctrl}
	mock.recorder = &MockIDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIDB) EXPECT() *MockIDBMockRecorder {
	return m.recorder
}

// PingContext mocks base method.
func (m *MockIDB) PingContext(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingContext", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// PingContext indicates an expected call of PingContext.
func (mr *MockIDBMockRecorder) PingContext(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingContext", reflect.TypeOf((*MockIDB)(nil).PingContext), ctx)
}