// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/documentlog/store/interface.go

// Package store is a generated GoMock package.
package store

import (
	gomock "github.com/golang/mock/gomock"
	model "github.com/nuts-foundation/nuts-network/pkg/model"
	io "io"
	reflect "reflect"
)

// MockDocumentStore is a mock of DocumentStore interface
type MockDocumentStore struct {
	ctrl     *gomock.Controller
	recorder *MockDocumentStoreMockRecorder
}

// MockDocumentStoreMockRecorder is the mock recorder for MockDocumentStore
type MockDocumentStoreMockRecorder struct {
	mock *MockDocumentStore
}

// NewMockDocumentStore creates a new mock instance
func NewMockDocumentStore(ctrl *gomock.Controller) *MockDocumentStore {
	mock := &MockDocumentStore{ctrl: ctrl}
	mock.recorder = &MockDocumentStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDocumentStore) EXPECT() *MockDocumentStoreMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockDocumentStore) Get(hash model.Hash) (*model.DocumentDescriptor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", hash)
	ret0, _ := ret[0].(*model.DocumentDescriptor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockDocumentStoreMockRecorder) Get(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDocumentStore)(nil).Get), hash)
}

// GetByConsistencyHash mocks base method
func (m *MockDocumentStore) GetByConsistencyHash(hash model.Hash) (*model.DocumentDescriptor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByConsistencyHash", hash)
	ret0, _ := ret[0].(*model.DocumentDescriptor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByConsistencyHash indicates an expected call of GetByConsistencyHash
func (mr *MockDocumentStoreMockRecorder) GetByConsistencyHash(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByConsistencyHash", reflect.TypeOf((*MockDocumentStore)(nil).GetByConsistencyHash), hash)
}

// Add mocks base method
func (m *MockDocumentStore) Add(document model.Document) (model.Hash, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", document)
	ret0, _ := ret[0].(model.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Add indicates an expected call of Add
func (mr *MockDocumentStoreMockRecorder) Add(document interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockDocumentStore)(nil).Add), document)
}

// GetAll mocks base method
func (m *MockDocumentStore) GetAll() ([]model.DocumentDescriptor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll")
	ret0, _ := ret[0].([]model.DocumentDescriptor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll
func (mr *MockDocumentStoreMockRecorder) GetAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockDocumentStore)(nil).GetAll))
}

// WriteContents mocks base method
func (m *MockDocumentStore) WriteContents(hash model.Hash, contents io.Reader) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteContents", hash, contents)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteContents indicates an expected call of WriteContents
func (mr *MockDocumentStoreMockRecorder) WriteContents(hash, contents interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteContents", reflect.TypeOf((*MockDocumentStore)(nil).WriteContents), hash, contents)
}

// ReadContents mocks base method
func (m *MockDocumentStore) ReadContents(hash model.Hash) (io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadContents", hash)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadContents indicates an expected call of ReadContents
func (mr *MockDocumentStoreMockRecorder) ReadContents(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadContents", reflect.TypeOf((*MockDocumentStore)(nil).ReadContents), hash)
}

// LastConsistencyHash mocks base method
func (m *MockDocumentStore) LastConsistencyHash() model.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastConsistencyHash")
	ret0, _ := ret[0].(model.Hash)
	return ret0
}

// LastConsistencyHash indicates an expected call of LastConsistencyHash
func (mr *MockDocumentStoreMockRecorder) LastConsistencyHash() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastConsistencyHash", reflect.TypeOf((*MockDocumentStore)(nil).LastConsistencyHash))
}

// ContentsSize mocks base method
func (m *MockDocumentStore) ContentsSize() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContentsSize")
	ret0, _ := ret[0].(int)
	return ret0
}

// ContentsSize indicates an expected call of ContentsSize
func (mr *MockDocumentStoreMockRecorder) ContentsSize() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContentsSize", reflect.TypeOf((*MockDocumentStore)(nil).ContentsSize))
}

// Size mocks base method
func (m *MockDocumentStore) Size() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size")
	ret0, _ := ret[0].(int)
	return ret0
}

// Size indicates an expected call of Size
func (mr *MockDocumentStoreMockRecorder) Size() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockDocumentStore)(nil).Size))
}