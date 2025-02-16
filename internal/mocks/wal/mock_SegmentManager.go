// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	wal "github.com/neekrasov/kvdb/internal/database/storage/wal"
	mock "github.com/stretchr/testify/mock"
)

// SegmentManager is an autogenerated mock type for the SegmentManager type
type SegmentManager struct {
	mock.Mock
}

type SegmentManager_Expecter struct {
	mock *mock.Mock
}

func (_m *SegmentManager) EXPECT() *SegmentManager_Expecter {
	return &SegmentManager_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *SegmentManager) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SegmentManager_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type SegmentManager_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *SegmentManager_Expecter) Close() *SegmentManager_Close_Call {
	return &SegmentManager_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *SegmentManager_Close_Call) Run(run func()) *SegmentManager_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SegmentManager_Close_Call) Return(_a0 error) *SegmentManager_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SegmentManager_Close_Call) RunAndReturn(run func() error) *SegmentManager_Close_Call {
	_c.Call.Return(run)
	return _c
}

// ForEach provides a mock function with given fields: action
func (_m *SegmentManager) ForEach(action func([]byte) error) error {
	ret := _m.Called(action)

	if len(ret) == 0 {
		panic("no return value specified for ForEach")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(func([]byte) error) error); ok {
		r0 = rf(action)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SegmentManager_ForEach_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ForEach'
type SegmentManager_ForEach_Call struct {
	*mock.Call
}

// ForEach is a helper method to define mock.On call
//   - action func([]byte) error
func (_e *SegmentManager_Expecter) ForEach(action interface{}) *SegmentManager_ForEach_Call {
	return &SegmentManager_ForEach_Call{Call: _e.mock.On("ForEach", action)}
}

func (_c *SegmentManager_ForEach_Call) Run(run func(action func([]byte) error)) *SegmentManager_ForEach_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(func([]byte) error))
	})
	return _c
}

func (_c *SegmentManager_ForEach_Call) Return(_a0 error) *SegmentManager_ForEach_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SegmentManager_ForEach_Call) RunAndReturn(run func(func([]byte) error) error) *SegmentManager_ForEach_Call {
	_c.Call.Return(run)
	return _c
}

// Write provides a mock function with given fields: data
func (_m *SegmentManager) Write(data []wal.WriteEntry) error {
	ret := _m.Called(data)

	if len(ret) == 0 {
		panic("no return value specified for Write")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]wal.WriteEntry) error); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SegmentManager_Write_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Write'
type SegmentManager_Write_Call struct {
	*mock.Call
}

// Write is a helper method to define mock.On call
//   - data []wal.WriteEntry
func (_e *SegmentManager_Expecter) Write(data interface{}) *SegmentManager_Write_Call {
	return &SegmentManager_Write_Call{Call: _e.mock.On("Write", data)}
}

func (_c *SegmentManager_Write_Call) Run(run func(data []wal.WriteEntry)) *SegmentManager_Write_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]wal.WriteEntry))
	})
	return _c
}

func (_c *SegmentManager_Write_Call) Return(_a0 error) *SegmentManager_Write_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SegmentManager_Write_Call) RunAndReturn(run func([]wal.WriteEntry) error) *SegmentManager_Write_Call {
	_c.Call.Return(run)
	return _c
}

// NewSegmentManager creates a new instance of SegmentManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSegmentManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *SegmentManager {
	mock := &SegmentManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
