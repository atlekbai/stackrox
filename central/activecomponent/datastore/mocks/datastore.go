// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stackrox/rox/central/activecomponent/datastore (interfaces: DataStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	converter "github.com/stackrox/rox/central/activecomponent/converter"
	v1 "github.com/stackrox/rox/generated/api/v1"
	storage "github.com/stackrox/rox/generated/storage"
	search "github.com/stackrox/rox/pkg/search"
	reflect "reflect"
)

// MockDataStore is a mock of DataStore interface
type MockDataStore struct {
	ctrl     *gomock.Controller
	recorder *MockDataStoreMockRecorder
}

// MockDataStoreMockRecorder is the mock recorder for MockDataStore
type MockDataStoreMockRecorder struct {
	mock *MockDataStore
}

// NewMockDataStore creates a new mock instance
func NewMockDataStore(ctrl *gomock.Controller) *MockDataStore {
	mock := &MockDataStore{ctrl: ctrl}
	mock.recorder = &MockDataStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDataStore) EXPECT() *MockDataStoreMockRecorder {
	return m.recorder
}

// DeleteBatch mocks base method
func (m *MockDataStore) DeleteBatch(arg0 context.Context, arg1 ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteBatch", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteBatch indicates an expected call of DeleteBatch
func (mr *MockDataStoreMockRecorder) DeleteBatch(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBatch", reflect.TypeOf((*MockDataStore)(nil).DeleteBatch), varargs...)
}

// Exists mocks base method
func (m *MockDataStore) Exists(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exists indicates an expected call of Exists
func (mr *MockDataStoreMockRecorder) Exists(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockDataStore)(nil).Exists), arg0, arg1)
}

// Get mocks base method
func (m *MockDataStore) Get(arg0 context.Context, arg1 string) (*storage.ActiveComponent, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(*storage.ActiveComponent)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Get indicates an expected call of Get
func (mr *MockDataStoreMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDataStore)(nil).Get), arg0, arg1)
}

// GetBatch mocks base method
func (m *MockDataStore) GetBatch(arg0 context.Context, arg1 []string) ([]*storage.ActiveComponent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBatch", arg0, arg1)
	ret0, _ := ret[0].([]*storage.ActiveComponent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBatch indicates an expected call of GetBatch
func (mr *MockDataStoreMockRecorder) GetBatch(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBatch", reflect.TypeOf((*MockDataStore)(nil).GetBatch), arg0, arg1)
}

// Search mocks base method
func (m *MockDataStore) Search(arg0 context.Context, arg1 *v1.Query) ([]search.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Search", arg0, arg1)
	ret0, _ := ret[0].([]search.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Search indicates an expected call of Search
func (mr *MockDataStoreMockRecorder) Search(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Search", reflect.TypeOf((*MockDataStore)(nil).Search), arg0, arg1)
}

// SearchRawActiveComponents mocks base method
func (m *MockDataStore) SearchRawActiveComponents(arg0 context.Context, arg1 *v1.Query) ([]*storage.ActiveComponent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchRawActiveComponents", arg0, arg1)
	ret0, _ := ret[0].([]*storage.ActiveComponent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchRawActiveComponents indicates an expected call of SearchRawActiveComponents
func (mr *MockDataStoreMockRecorder) SearchRawActiveComponents(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchRawActiveComponents", reflect.TypeOf((*MockDataStore)(nil).SearchRawActiveComponents), arg0, arg1)
}

// UpsertBatch mocks base method
func (m *MockDataStore) UpsertBatch(arg0 context.Context, arg1 []*converter.CompleteActiveComponent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertBatch", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertBatch indicates an expected call of UpsertBatch
func (mr *MockDataStoreMockRecorder) UpsertBatch(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertBatch", reflect.TypeOf((*MockDataStore)(nil).UpsertBatch), arg0, arg1)
}
