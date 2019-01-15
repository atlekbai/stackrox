// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stackrox/rox/central/enrichment (interfaces: Enricher)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	storage "github.com/stackrox/rox/generated/storage"
	reflect "reflect"
)

// MockEnricher is a mock of Enricher interface
type MockEnricher struct {
	ctrl     *gomock.Controller
	recorder *MockEnricherMockRecorder
}

// MockEnricherMockRecorder is the mock recorder for MockEnricher
type MockEnricherMockRecorder struct {
	mock *MockEnricher
}

// NewMockEnricher creates a new mock instance
func NewMockEnricher(ctrl *gomock.Controller) *MockEnricher {
	mock := &MockEnricher{ctrl: ctrl}
	mock.recorder = &MockEnricherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEnricher) EXPECT() *MockEnricherMockRecorder {
	return m.recorder
}

// EnrichDeployment mocks base method
func (m *MockEnricher) EnrichDeployment(arg0 *storage.Deployment) ([]*storage.Image, bool, error) {
	ret := m.ctrl.Call(m, "EnrichDeployment", arg0)
	ret0, _ := ret[0].([]*storage.Image)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// EnrichDeployment indicates an expected call of EnrichDeployment
func (mr *MockEnricherMockRecorder) EnrichDeployment(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnrichDeployment", reflect.TypeOf((*MockEnricher)(nil).EnrichDeployment), arg0)
}

// EnrichDeploymentWithEmptyImages mocks base method
func (m *MockEnricher) EnrichDeploymentWithEmptyImages(arg0 *storage.Deployment) ([]*storage.Image, bool, error) {
	ret := m.ctrl.Call(m, "EnrichDeploymentWithEmptyImages", arg0)
	ret0, _ := ret[0].([]*storage.Image)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// EnrichDeploymentWithEmptyImages indicates an expected call of EnrichDeploymentWithEmptyImages
func (mr *MockEnricherMockRecorder) EnrichDeploymentWithEmptyImages(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnrichDeploymentWithEmptyImages", reflect.TypeOf((*MockEnricher)(nil).EnrichDeploymentWithEmptyImages), arg0)
}
