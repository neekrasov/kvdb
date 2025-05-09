// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	storage "github.com/neekrasov/kvdb/internal/database/storage"

	sync "github.com/neekrasov/kvdb/pkg/sync"
)

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

type Storage_Expecter struct {
	mock *mock.Mock
}

func (_m *Storage) EXPECT() *Storage_Expecter {
	return &Storage_Expecter{mock: &_m.Mock}
}

// Del provides a mock function with given fields: ctx, key
func (_m *Storage) Del(ctx context.Context, key string) error {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Del")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_Del_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Del'
type Storage_Del_Call struct {
	*mock.Call
}

// Del is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *Storage_Expecter) Del(ctx interface{}, key interface{}) *Storage_Del_Call {
	return &Storage_Del_Call{Call: _e.mock.On("Del", ctx, key)}
}

func (_c *Storage_Del_Call) Run(run func(ctx context.Context, key string)) *Storage_Del_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Storage_Del_Call) Return(_a0 error) *Storage_Del_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_Del_Call) RunAndReturn(run func(context.Context, string) error) *Storage_Del_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, key
func (_m *Storage) Get(ctx context.Context, key string) (string, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Storage_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *Storage_Expecter) Get(ctx interface{}, key interface{}) *Storage_Get_Call {
	return &Storage_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *Storage_Get_Call) Run(run func(ctx context.Context, key string)) *Storage_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Storage_Get_Call) Return(_a0 string, _a1 error) *Storage_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_Get_Call) RunAndReturn(run func(context.Context, string) (string, error)) *Storage_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Set provides a mock function with given fields: ctx, key, value
func (_m *Storage) Set(ctx context.Context, key string, value string) error {
	ret := _m.Called(ctx, key, value)

	if len(ret) == 0 {
		panic("no return value specified for Set")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type Storage_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - value string
func (_e *Storage_Expecter) Set(ctx interface{}, key interface{}, value interface{}) *Storage_Set_Call {
	return &Storage_Set_Call{Call: _e.mock.On("Set", ctx, key, value)}
}

func (_c *Storage_Set_Call) Run(run func(ctx context.Context, key string, value string)) *Storage_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *Storage_Set_Call) Return(_a0 error) *Storage_Set_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_Set_Call) RunAndReturn(run func(context.Context, string, string) error) *Storage_Set_Call {
	_c.Call.Return(run)
	return _c
}

// Stats provides a mock function with no fields
func (_m *Storage) Stats() (*storage.Stats, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Stats")
	}

	var r0 *storage.Stats
	var r1 error
	if rf, ok := ret.Get(0).(func() (*storage.Stats, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *storage.Stats); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.Stats)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_Stats_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stats'
type Storage_Stats_Call struct {
	*mock.Call
}

// Stats is a helper method to define mock.On call
func (_e *Storage_Expecter) Stats() *Storage_Stats_Call {
	return &Storage_Stats_Call{Call: _e.mock.On("Stats")}
}

func (_c *Storage_Stats_Call) Run(run func()) *Storage_Stats_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Storage_Stats_Call) Return(_a0 *storage.Stats, _a1 error) *Storage_Stats_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_Stats_Call) RunAndReturn(run func() (*storage.Stats, error)) *Storage_Stats_Call {
	_c.Call.Return(run)
	return _c
}

// Watch provides a mock function with given fields: ctx, key
func (_m *Storage) Watch(ctx context.Context, key string) sync.Future[string] {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Watch")
	}

	var r0 sync.Future[string]
	if rf, ok := ret.Get(0).(func(context.Context, string) sync.Future[string]); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(sync.Future[string])
	}

	return r0
}

// Storage_Watch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Watch'
type Storage_Watch_Call struct {
	*mock.Call
}

// Watch is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *Storage_Expecter) Watch(ctx interface{}, key interface{}) *Storage_Watch_Call {
	return &Storage_Watch_Call{Call: _e.mock.On("Watch", ctx, key)}
}

func (_c *Storage_Watch_Call) Run(run func(ctx context.Context, key string)) *Storage_Watch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Storage_Watch_Call) Return(_a0 sync.Future[string]) *Storage_Watch_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_Watch_Call) RunAndReturn(run func(context.Context, string) sync.Future[string]) *Storage_Watch_Call {
	_c.Call.Return(run)
	return _c
}

// NewStorage creates a new instance of Storage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *Storage {
	mock := &Storage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
