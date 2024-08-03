// Code generated by MockGen. DO NOT EDIT.
// Source: internal/metrics/storage.go

// Package mock is a generated GoMock package.
package mock

import (
	metrics "gmetrics/internal/metrics"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// AddCounter mocks base method.
func (m *MockStorage) AddCounter(name string, value metrics.Counter) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddCounter", name, value)
}

// AddCounter indicates an expected call of AddCounter.
func (mr *MockStorageMockRecorder) AddCounter(name, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCounter", reflect.TypeOf((*MockStorage)(nil).AddCounter), name, value)
}

// GetCounter mocks base method.
func (m *MockStorage) GetCounter(name string) (metrics.Counter, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCounter", name)
	ret0, _ := ret[0].(metrics.Counter)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetCounter indicates an expected call of GetCounter.
func (mr *MockStorageMockRecorder) GetCounter(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCounter", reflect.TypeOf((*MockStorage)(nil).GetCounter), name)
}

// GetCounters mocks base method.
func (m *MockStorage) GetCounters() map[string]metrics.Counter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCounters")
	ret0, _ := ret[0].(map[string]metrics.Counter)
	return ret0
}

// GetCounters indicates an expected call of GetCounters.
func (mr *MockStorageMockRecorder) GetCounters() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCounters", reflect.TypeOf((*MockStorage)(nil).GetCounters))
}

// GetGauge mocks base method.
func (m *MockStorage) GetGauge(name string) (metrics.Gauge, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGauge", name)
	ret0, _ := ret[0].(metrics.Gauge)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetGauge indicates an expected call of GetGauge.
func (mr *MockStorageMockRecorder) GetGauge(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGauge", reflect.TypeOf((*MockStorage)(nil).GetGauge), name)
}

// GetGauges mocks base method.
func (m *MockStorage) GetGauges() map[string]metrics.Gauge {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGauges")
	ret0, _ := ret[0].(map[string]metrics.Gauge)
	return ret0
}

// GetGauges indicates an expected call of GetGauges.
func (mr *MockStorageMockRecorder) GetGauges() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGauges", reflect.TypeOf((*MockStorage)(nil).GetGauges))
}

// SetGauge mocks base method.
func (m *MockStorage) SetGauge(name string, value metrics.Gauge) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetGauge", name, value)
}

// SetGauge indicates an expected call of SetGauge.
func (mr *MockStorageMockRecorder) SetGauge(name, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetGauge", reflect.TypeOf((*MockStorage)(nil).SetGauge), name, value)
}
