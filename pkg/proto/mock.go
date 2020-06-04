// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/proto/interface.go

// Package proto is a generated GoMock package.
package proto

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	model "github.com/nuts-foundation/nuts-network/pkg/model"
	p2p "github.com/nuts-foundation/nuts-network/pkg/p2p"
	stats "github.com/nuts-foundation/nuts-network/pkg/stats"
	io "io"
	reflect "reflect"
	time "time"
)

// MockProtocol is a mock of Protocol interface
type MockProtocol struct {
	ctrl     *gomock.Controller
	recorder *MockProtocolMockRecorder
}

// MockProtocolMockRecorder is the mock recorder for MockProtocol
type MockProtocolMockRecorder struct {
	mock *MockProtocol
}

// NewMockProtocol creates a new mock instance
func NewMockProtocol(ctrl *gomock.Controller) *MockProtocol {
	mock := &MockProtocol{ctrl: ctrl}
	mock.recorder = &MockProtocolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProtocol) EXPECT() *MockProtocolMockRecorder {
	return m.recorder
}

// Statistics mocks base method
func (m *MockProtocol) Statistics() []stats.Statistic {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Statistics")
	ret0, _ := ret[0].([]stats.Statistic)
	return ret0
}

// Statistics indicates an expected call of Statistics
func (mr *MockProtocolMockRecorder) Statistics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Statistics", reflect.TypeOf((*MockProtocol)(nil).Statistics))
}

// Start mocks base method
func (m *MockProtocol) Start(p2pNetwork p2p.P2PNetwork, source HashSource) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start", p2pNetwork, source)
}

// Start indicates an expected call of Start
func (mr *MockProtocolMockRecorder) Start(p2pNetwork, source interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockProtocol)(nil).Start), p2pNetwork, source)
}

// Stop mocks base method
func (m *MockProtocol) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockProtocolMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockProtocol)(nil).Stop))
}

// ReceivedConsistencyHashes mocks base method
func (m *MockProtocol) ReceivedConsistencyHashes() PeerHashQueue {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReceivedConsistencyHashes")
	ret0, _ := ret[0].(PeerHashQueue)
	return ret0
}

// ReceivedConsistencyHashes indicates an expected call of ReceivedConsistencyHashes
func (mr *MockProtocolMockRecorder) ReceivedConsistencyHashes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReceivedConsistencyHashes", reflect.TypeOf((*MockProtocol)(nil).ReceivedConsistencyHashes))
}

// ReceivedDocumentHashes mocks base method
func (m *MockProtocol) ReceivedDocumentHashes() PeerHashQueue {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReceivedDocumentHashes")
	ret0, _ := ret[0].(PeerHashQueue)
	return ret0
}

// ReceivedDocumentHashes indicates an expected call of ReceivedDocumentHashes
func (mr *MockProtocolMockRecorder) ReceivedDocumentHashes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReceivedDocumentHashes", reflect.TypeOf((*MockProtocol)(nil).ReceivedDocumentHashes))
}

// AdvertConsistencyHash mocks base method
func (m *MockProtocol) AdvertConsistencyHash(hash model.Hash) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AdvertConsistencyHash", hash)
}

// AdvertConsistencyHash indicates an expected call of AdvertConsistencyHash
func (mr *MockProtocolMockRecorder) AdvertConsistencyHash(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AdvertConsistencyHash", reflect.TypeOf((*MockProtocol)(nil).AdvertConsistencyHash), hash)
}

// QueryHashList mocks base method
func (m *MockProtocol) QueryHashList(peer model.PeerID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryHashList", peer)
	ret0, _ := ret[0].(error)
	return ret0
}

// QueryHashList indicates an expected call of QueryHashList
func (mr *MockProtocolMockRecorder) QueryHashList(peer interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryHashList", reflect.TypeOf((*MockProtocol)(nil).QueryHashList), peer)
}

// MockPeerHashQueue is a mock of PeerHashQueue interface
type MockPeerHashQueue struct {
	ctrl     *gomock.Controller
	recorder *MockPeerHashQueueMockRecorder
}

// MockPeerHashQueueMockRecorder is the mock recorder for MockPeerHashQueue
type MockPeerHashQueueMockRecorder struct {
	mock *MockPeerHashQueue
}

// NewMockPeerHashQueue creates a new mock instance
func NewMockPeerHashQueue(ctrl *gomock.Controller) *MockPeerHashQueue {
	mock := &MockPeerHashQueue{ctrl: ctrl}
	mock.recorder = &MockPeerHashQueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPeerHashQueue) EXPECT() *MockPeerHashQueueMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockPeerHashQueue) Get(cxt context.Context) (PeerHash, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", cxt)
	ret0, _ := ret[0].(PeerHash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockPeerHashQueueMockRecorder) Get(cxt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockPeerHashQueue)(nil).Get), cxt)
}

// MockHashSource is a mock of HashSource interface
type MockHashSource struct {
	ctrl     *gomock.Controller
	recorder *MockHashSourceMockRecorder
}

// MockHashSourceMockRecorder is the mock recorder for MockHashSource
type MockHashSourceMockRecorder struct {
	mock *MockHashSource
}

// NewMockHashSource creates a new mock instance
func NewMockHashSource(ctrl *gomock.Controller) *MockHashSource {
	mock := &MockHashSource{ctrl: ctrl}
	mock.recorder = &MockHashSourceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHashSource) EXPECT() *MockHashSourceMockRecorder {
	return m.recorder
}

// Documents mocks base method
func (m *MockHashSource) Documents() []model.Document {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Documents")
	ret0, _ := ret[0].([]model.Document)
	return ret0
}

// Documents indicates an expected call of Documents
func (mr *MockHashSourceMockRecorder) Documents() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Documents", reflect.TypeOf((*MockHashSource)(nil).Documents))
}

// HasDocument mocks base method
func (m *MockHashSource) HasDocument(hash model.Hash) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasDocument", hash)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasDocument indicates an expected call of HasDocument
func (mr *MockHashSourceMockRecorder) HasDocument(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasDocument", reflect.TypeOf((*MockHashSource)(nil).HasDocument), hash)
}

// HasContentsForDocument mocks base method
func (m *MockHashSource) HasContentsForDocument(hash model.Hash) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasContentsForDocument", hash)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasContentsForDocument indicates an expected call of HasContentsForDocument
func (mr *MockHashSourceMockRecorder) HasContentsForDocument(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasContentsForDocument", reflect.TypeOf((*MockHashSource)(nil).HasContentsForDocument), hash)
}

// GetDocument mocks base method
func (m *MockHashSource) GetDocument(hash model.Hash) *model.Document {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDocument", hash)
	ret0, _ := ret[0].(*model.Document)
	return ret0
}

// GetDocument indicates an expected call of GetDocument
func (mr *MockHashSourceMockRecorder) GetDocument(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDocument", reflect.TypeOf((*MockHashSource)(nil).GetDocument), hash)
}

// GetDocumentContents mocks base method
func (m *MockHashSource) GetDocumentContents(hash model.Hash) (io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDocumentContents", hash)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDocumentContents indicates an expected call of GetDocumentContents
func (mr *MockHashSourceMockRecorder) GetDocumentContents(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDocumentContents", reflect.TypeOf((*MockHashSource)(nil).GetDocumentContents), hash)
}

// AddDocument mocks base method
func (m *MockHashSource) AddDocument(document model.Document) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddDocument", document)
}

// AddDocument indicates an expected call of AddDocument
func (mr *MockHashSourceMockRecorder) AddDocument(document interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddDocument", reflect.TypeOf((*MockHashSource)(nil).AddDocument), document)
}

// AddDocumentWithContents mocks base method
func (m *MockHashSource) AddDocumentWithContents(timestamp time.Time, documentType string, contents io.Reader) (model.Document, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddDocumentWithContents", timestamp, documentType, contents)
	ret0, _ := ret[0].(model.Document)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddDocumentWithContents indicates an expected call of AddDocumentWithContents
func (mr *MockHashSourceMockRecorder) AddDocumentWithContents(timestamp, documentType, contents interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddDocumentWithContents", reflect.TypeOf((*MockHashSource)(nil).AddDocumentWithContents), timestamp, documentType, contents)
}

// AddDocumentContents mocks base method
func (m *MockHashSource) AddDocumentContents(hash model.Hash, contents io.Reader) (model.Document, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddDocumentContents", hash, contents)
	ret0, _ := ret[0].(model.Document)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddDocumentContents indicates an expected call of AddDocumentContents
func (mr *MockHashSourceMockRecorder) AddDocumentContents(hash, contents interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddDocumentContents", reflect.TypeOf((*MockHashSource)(nil).AddDocumentContents), hash, contents)
}