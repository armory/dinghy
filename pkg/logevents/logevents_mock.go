// Code generated by MockGen. DO NOT EDIT.
// Source: logevents.go

// Package logevents is a generated GoMock package.
package logevents

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockLogEventsClient is a mock of LogEventsClient interface
type MockLogEventsClient struct {
	ctrl     *gomock.Controller
	recorder *MockLogEventsClientMockRecorder
}

// MockLogEventsClientMockRecorder is the mock recorder for MockLogEventsClient
type MockLogEventsClientMockRecorder struct {
	mock *MockLogEventsClient
}

// NewMockLogEventsClient creates a new mock instance
func NewMockLogEventsClient(ctrl *gomock.Controller) *MockLogEventsClient {
	mock := &MockLogEventsClient{ctrl: ctrl}
	mock.recorder = &MockLogEventsClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLogEventsClient) EXPECT() *MockLogEventsClientMockRecorder {
	return m.recorder
}

// GetLogEvents mocks base method
func (m *MockLogEventsClient) GetLogEvents() ([]LogEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLogEvents")
	ret0, _ := ret[0].([]LogEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLogEvents indicates an expected call of GetLogEvents
func (mr *MockLogEventsClientMockRecorder) GetLogEvents() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLogEvents", reflect.TypeOf((*MockLogEventsClient)(nil).GetLogEvents))
}

// SaveLogEvent mocks base method
func (m *MockLogEventsClient) SaveLogEvent(logEvent LogEvent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveLogEvent", logEvent)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveLogEvent indicates an expected call of SaveLogEvent
func (mr *MockLogEventsClientMockRecorder) SaveLogEvent(logEvent interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveLogEvent", reflect.TypeOf((*MockLogEventsClient)(nil).SaveLogEvent), logEvent)
}
