// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/distlock/distlock.go

// Package distlock is a generated GoMock package.
package distlock

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockDistLock is a mock of DistLock interface.
type MockDistLock struct {
	ctrl     *gomock.Controller
	recorder *MockDistLockMockRecorder
}

// MockDistLockMockRecorder is the mock recorder for MockDistLock.
type MockDistLockMockRecorder struct {
	mock *MockDistLock
}

// NewMockDistLock creates a new mock instance.
func NewMockDistLock(ctrl *gomock.Controller) *MockDistLock {
	mock := &MockDistLock{ctrl: ctrl}
	mock.recorder = &MockDistLockMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDistLock) EXPECT() *MockDistLockMockRecorder {
	return m.recorder
}

// Acquire mocks base method.
func (m *MockDistLock) Acquire(ctx context.Context, key string, duration time.Duration, retryCount int) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Acquire", ctx, key, duration, retryCount)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Acquire indicates an expected call of Acquire.
func (mr *MockDistLockMockRecorder) Acquire(ctx, key, duration, retryCount interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Acquire", reflect.TypeOf((*MockDistLock)(nil).Acquire), ctx, key, duration, retryCount)
}

// Release mocks base method.
func (m *MockDistLock) Release(ctx context.Context, key string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Release", ctx, key)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Release indicates an expected call of Release.
func (mr *MockDistLockMockRecorder) Release(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Release", reflect.TypeOf((*MockDistLock)(nil).Release), ctx, key)
}