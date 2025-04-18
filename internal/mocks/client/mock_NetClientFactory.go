// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	client "github.com/neekrasov/kvdb/pkg/client"
	mock "github.com/stretchr/testify/mock"

	tcp "github.com/neekrasov/kvdb/internal/delivery/tcp"
)

// NetClientFactory is an autogenerated mock type for the NetClientFactory type
type NetClientFactory struct {
	mock.Mock
}

type NetClientFactory_Expecter struct {
	mock *mock.Mock
}

func (_m *NetClientFactory) EXPECT() *NetClientFactory_Expecter {
	return &NetClientFactory_Expecter{mock: &_m.Mock}
}

// Make provides a mock function with given fields: address, opts
func (_m *NetClientFactory) Make(address string, opts ...tcp.ClientOption) (client.NetClient, error) {
	var tmpRet mock.Arguments
	if len(opts) > 0 {
		tmpRet = _m.Called(address, opts)
	} else {
		tmpRet = _m.Called(address)
	}
	ret := tmpRet

	if len(ret) == 0 {
		panic("no return value specified for Make")
	}

	var r0 client.NetClient
	var r1 error
	if rf, ok := ret.Get(0).(func(string, ...tcp.ClientOption) (client.NetClient, error)); ok {
		return rf(address, opts...)
	}
	if rf, ok := ret.Get(0).(func(string, ...tcp.ClientOption) client.NetClient); ok {
		r0 = rf(address, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.NetClient)
		}
	}

	if rf, ok := ret.Get(1).(func(string, ...tcp.ClientOption) error); ok {
		r1 = rf(address, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NetClientFactory_Make_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Make'
type NetClientFactory_Make_Call struct {
	*mock.Call
}

// Make is a helper method to define mock.On call
//   - address string
//   - opts ...tcp.ClientOption
func (_e *NetClientFactory_Expecter) Make(address interface{}, opts ...interface{}) *NetClientFactory_Make_Call {
	return &NetClientFactory_Make_Call{Call: _e.mock.On("Make",
		append([]interface{}{address}, opts...)...)}
}

func (_c *NetClientFactory_Make_Call) Run(run func(address string, opts ...tcp.ClientOption)) *NetClientFactory_Make_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]tcp.ClientOption, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(tcp.ClientOption)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *NetClientFactory_Make_Call) Return(_a0 client.NetClient, _a1 error) *NetClientFactory_Make_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NetClientFactory_Make_Call) RunAndReturn(run func(string, ...tcp.ClientOption) (client.NetClient, error)) *NetClientFactory_Make_Call {
	_c.Call.Return(run)
	return _c
}

// NewNetClientFactory creates a new instance of NetClientFactory. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNetClientFactory(t interface {
	mock.TestingT
	Cleanup(func())
}) *NetClientFactory {
	mock := &NetClientFactory{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
