// Code generated by MockGen. DO NOT EDIT.
// Source: cmd/agent/sendpool/pool.go

// Package sendpool is a generated GoMock package.
package sendpool

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockIClient is a mock of IClient interface.
type MockIClient struct {
	ctrl     *gomock.Controller
	recorder *MockIClientMockRecorder
}

// MockIClientMockRecorder is the mock recorder for MockIClient.
type MockIClientMockRecorder struct {
	mock *MockIClient
}

// NewMockIClient creates a new mock instance.
func NewMockIClient(ctrl *gomock.Controller) *MockIClient {
	mock := &MockIClient{ctrl: ctrl}
	mock.recorder = &MockIClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIClient) EXPECT() *MockIClientMockRecorder {
	return m.recorder
}

// EnableManualCompression mocks base method.
func (m *MockIClient) EnableManualCompression() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnableManualCompression")
	ret0, _ := ret[0].(bool)
	return ret0
}

// EnableManualCompression indicates an expected call of EnableManualCompression.
func (mr *MockIClientMockRecorder) EnableManualCompression() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnableManualCompression", reflect.TypeOf((*MockIClient)(nil).EnableManualCompression))
}

// Post mocks base method.
func (m *MockIClient) Post(url string, body []byte, headers ...Header) (MetricResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{url, body}
	for _, a := range headers {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Post", varargs...)
	ret0, _ := ret[0].(MetricResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Post indicates an expected call of Post.
func (mr *MockIClientMockRecorder) Post(url, body interface{}, headers ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{url, body}, headers...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Post", reflect.TypeOf((*MockIClient)(nil).Post), varargs...)
}

// MockMetricResponse is a mock of MetricResponse interface.
type MockMetricResponse struct {
	ctrl     *gomock.Controller
	recorder *MockMetricResponseMockRecorder
}

// MockMetricResponseMockRecorder is the mock recorder for MockMetricResponse.
type MockMetricResponseMockRecorder struct {
	mock *MockMetricResponse
}

// NewMockMetricResponse creates a new mock instance.
func NewMockMetricResponse(ctrl *gomock.Controller) *MockMetricResponse {
	mock := &MockMetricResponse{ctrl: ctrl}
	mock.recorder = &MockMetricResponseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricResponse) EXPECT() *MockMetricResponseMockRecorder {
	return m.recorder
}

// StatusCode mocks base method.
func (m *MockMetricResponse) StatusCode() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StatusCode")
	ret0, _ := ret[0].(int)
	return ret0
}

// StatusCode indicates an expected call of StatusCode.
func (mr *MockMetricResponseMockRecorder) StatusCode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StatusCode", reflect.TypeOf((*MockMetricResponse)(nil).StatusCode))
}
