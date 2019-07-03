package hcl

import (
	"github.com/armory/plank"
	"github.com/golang/mock/gomock"
	"reflect"
)

// MockPlankClient is a mock of PlankClient interface
type MockPlankClient struct {
	ctrl     *gomock.Controller
	recorder *MockPlankClientMockRecorder
}

// MockPlankClientMockRecorder is the mock recorder for MockPlankClient
type MockPlankClientMockRecorder struct {
	mock *MockPlankClient
}

// NewMockPlankClient creates a new mock instance
func NewMockPlankClient(ctrl *gomock.Controller) *MockPlankClient {
	mock := &MockPlankClient{ctrl: ctrl}
	mock.recorder = &MockPlankClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPlankClient) EXPECT() *MockPlankClientMockRecorder {
	return m.recorder
}

// GetApplication mocks base method
func (m *MockPlankClient) GetApplication(arg0 string) (*plank.Application, error) {
	ret := m.ctrl.Call(m, "GetApplication", arg0)
	ret0, _ := ret[0].(*plank.Application)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplication indicates an expected call of GetApplication
func (mr *MockPlankClientMockRecorder) GetApplication(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplication", reflect.TypeOf((*MockPlankClient)(nil).GetApplication), arg0)
}

// CreateApplication mocks base method
func (m *MockPlankClient) CreateApplication(arg0 *plank.Application) error {
	ret := m.ctrl.Call(m, "CreateApplication", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateApplication indicates an expected call of CreateApplication
func (mr *MockPlankClientMockRecorder) CreateApplication(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateApplication", reflect.TypeOf((*MockPlankClient)(nil).CreateApplication), arg0)
}

// GetPipelines mocks base method
func (m *MockPlankClient) GetPipelines(arg0 string) ([]plank.Pipeline, error) {
	ret := m.ctrl.Call(m, "GetPipelines", arg0)
	ret0, _ := ret[0].([]plank.Pipeline)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPipelines indicates an expected call of GetPipelines
func (mr *MockPlankClientMockRecorder) GetPipelines(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPipelines", reflect.TypeOf((*MockPlankClient)(nil).GetPipelines), arg0)
}

// DeletePipeline mocks base method
func (m *MockPlankClient) DeletePipeline(arg0 plank.Pipeline) error {
	ret := m.ctrl.Call(m, "DeletePipeline", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePipeline indicates an expected call of DeletePipeline
func (mr *MockPlankClientMockRecorder) DeletePipeline(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePipeline", reflect.TypeOf((*MockPlankClient)(nil).DeletePipeline), arg0)
}

// UpsertPipeline mocks base method
func (m *MockPlankClient) UpsertPipeline(arg0 plank.Pipeline, arg1 string) error {
	ret := m.ctrl.Call(m, "UpsertPipeline", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertPipeline indicates an expected call of UpsertPipeline
func (mr *MockPlankClientMockRecorder) UpsertPipeline(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertPipeline", reflect.TypeOf((*MockPlankClient)(nil).UpsertPipeline), arg0, arg1)
}
