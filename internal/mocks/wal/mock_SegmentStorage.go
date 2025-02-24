// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	wal "github.com/neekrasov/kvdb/internal/database/storage/wal"
	mock "github.com/stretchr/testify/mock"
)

// SegmentStorage is an autogenerated mock type for the SegmentStorage type
type SegmentStorage struct {
	mock.Mock
}

type SegmentStorage_Expecter struct {
	mock *mock.Mock
}

func (_m *SegmentStorage) EXPECT() *SegmentStorage_Expecter {
	return &SegmentStorage_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: id, compressed
func (_m *SegmentStorage) Create(id int, compressed bool) (wal.Segment, error) {
	ret := _m.Called(id, compressed)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 wal.Segment
	var r1 error
	if rf, ok := ret.Get(0).(func(int, bool) (wal.Segment, error)); ok {
		return rf(id, compressed)
	}
	if rf, ok := ret.Get(0).(func(int, bool) wal.Segment); ok {
		r0 = rf(id, compressed)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(wal.Segment)
		}
	}

	if rf, ok := ret.Get(1).(func(int, bool) error); ok {
		r1 = rf(id, compressed)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SegmentStorage_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type SegmentStorage_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - id int
//   - compressed bool
func (_e *SegmentStorage_Expecter) Create(id interface{}, compressed interface{}) *SegmentStorage_Create_Call {
	return &SegmentStorage_Create_Call{Call: _e.mock.On("Create", id, compressed)}
}

func (_c *SegmentStorage_Create_Call) Run(run func(id int, compressed bool)) *SegmentStorage_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int), args[1].(bool))
	})
	return _c
}

func (_c *SegmentStorage_Create_Call) Return(_a0 wal.Segment, _a1 error) *SegmentStorage_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SegmentStorage_Create_Call) RunAndReturn(run func(int, bool) (wal.Segment, error)) *SegmentStorage_Create_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with no fields
func (_m *SegmentStorage) List() ([]int, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []int
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]int, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []int); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SegmentStorage_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type SegmentStorage_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
func (_e *SegmentStorage_Expecter) List() *SegmentStorage_List_Call {
	return &SegmentStorage_List_Call{Call: _e.mock.On("List")}
}

func (_c *SegmentStorage_List_Call) Run(run func()) *SegmentStorage_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SegmentStorage_List_Call) Return(_a0 []int, _a1 error) *SegmentStorage_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SegmentStorage_List_Call) RunAndReturn(run func() ([]int, error)) *SegmentStorage_List_Call {
	_c.Call.Return(run)
	return _c
}

// Open provides a mock function with given fields: id
func (_m *SegmentStorage) Open(id int) (wal.Segment, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for Open")
	}

	var r0 wal.Segment
	var r1 error
	if rf, ok := ret.Get(0).(func(int) (wal.Segment, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(int) wal.Segment); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(wal.Segment)
		}
	}

	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SegmentStorage_Open_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Open'
type SegmentStorage_Open_Call struct {
	*mock.Call
}

// Open is a helper method to define mock.On call
//   - id int
func (_e *SegmentStorage_Expecter) Open(id interface{}) *SegmentStorage_Open_Call {
	return &SegmentStorage_Open_Call{Call: _e.mock.On("Open", id)}
}

func (_c *SegmentStorage_Open_Call) Run(run func(id int)) *SegmentStorage_Open_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *SegmentStorage_Open_Call) Return(_a0 wal.Segment, _a1 error) *SegmentStorage_Open_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SegmentStorage_Open_Call) RunAndReturn(run func(int) (wal.Segment, error)) *SegmentStorage_Open_Call {
	_c.Call.Return(run)
	return _c
}

// Remove provides a mock function with given fields: id
func (_m *SegmentStorage) Remove(id int) error {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for Remove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SegmentStorage_Remove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Remove'
type SegmentStorage_Remove_Call struct {
	*mock.Call
}

// Remove is a helper method to define mock.On call
//   - id int
func (_e *SegmentStorage_Expecter) Remove(id interface{}) *SegmentStorage_Remove_Call {
	return &SegmentStorage_Remove_Call{Call: _e.mock.On("Remove", id)}
}

func (_c *SegmentStorage_Remove_Call) Run(run func(id int)) *SegmentStorage_Remove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *SegmentStorage_Remove_Call) Return(_a0 error) *SegmentStorage_Remove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SegmentStorage_Remove_Call) RunAndReturn(run func(int) error) *SegmentStorage_Remove_Call {
	_c.Call.Return(run)
	return _c
}

// NewSegmentStorage creates a new instance of SegmentStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSegmentStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *SegmentStorage {
	mock := &SegmentStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
