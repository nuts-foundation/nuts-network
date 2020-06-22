// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/nodelist/interface.go

// Package nodelist is a generated GoMock package.
package nodelist

import (
	gomock "github.com/golang/mock/gomock"
	model "github.com/nuts-foundation/nuts-network/pkg/model"
	reflect "reflect"
)

// MockNodeList is a mock of NodeList interface
type MockNodeList struct {
	ctrl     *gomock.Controller
	recorder *MockNodeListMockRecorder
}

// MockNodeListMockRecorder is the mock recorder for MockNodeList
type MockNodeListMockRecorder struct {
	mock *MockNodeList
}

// NewMockNodeList creates a new mock instance
func NewMockNodeList(ctrl *gomock.Controller) *MockNodeList {
	mock := &MockNodeList{ctrl: ctrl}
	mock.recorder = &MockNodeListMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNodeList) EXPECT() *MockNodeListMockRecorder {
	return m.recorder
}

// Start mocks base method
func (m *MockNodeList) Start(nodeID model.NodeID, address string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start", nodeID, address)
}

// Start indicates an expected call of Start
func (mr *MockNodeListMockRecorder) Start(nodeID, address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockNodeList)(nil).Start), nodeID, address)
}

// Stop mocks base method
func (m *MockNodeList) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockNodeListMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockNodeList)(nil).Stop))
}
