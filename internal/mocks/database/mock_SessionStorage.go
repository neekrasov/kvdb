// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	models "github.com/neekrasov/kvdb/internal/database/identity/models"
	mock "github.com/stretchr/testify/mock"
)

// SessionStorage is an autogenerated mock type for the SessionStorage type
type SessionStorage struct {
	mock.Mock
}

type SessionStorage_Expecter struct {
	mock *mock.Mock
}

func (_m *SessionStorage) EXPECT() *SessionStorage_Expecter {
	return &SessionStorage_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: id, user
func (_m *SessionStorage) Create(id string, user *models.User) error {
	ret := _m.Called(id, user)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *models.User) error); ok {
		r0 = rf(id, user)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SessionStorage_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type SessionStorage_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - id string
//   - user *models.User
func (_e *SessionStorage_Expecter) Create(id interface{}, user interface{}) *SessionStorage_Create_Call {
	return &SessionStorage_Create_Call{Call: _e.mock.On("Create", id, user)}
}

func (_c *SessionStorage_Create_Call) Run(run func(id string, user *models.User)) *SessionStorage_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(*models.User))
	})
	return _c
}

func (_c *SessionStorage_Create_Call) Return(_a0 error) *SessionStorage_Create_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SessionStorage_Create_Call) RunAndReturn(run func(string, *models.User) error) *SessionStorage_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: id
func (_m *SessionStorage) Delete(id string) {
	_m.Called(id)
}

// SessionStorage_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type SessionStorage_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - id string
func (_e *SessionStorage_Expecter) Delete(id interface{}) *SessionStorage_Delete_Call {
	return &SessionStorage_Delete_Call{Call: _e.mock.On("Delete", id)}
}

func (_c *SessionStorage_Delete_Call) Run(run func(id string)) *SessionStorage_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SessionStorage_Delete_Call) Return() *SessionStorage_Delete_Call {
	_c.Call.Return()
	return _c
}

func (_c *SessionStorage_Delete_Call) RunAndReturn(run func(string)) *SessionStorage_Delete_Call {
	_c.Run(run)
	return _c
}

// Get provides a mock function with given fields: id
func (_m *SessionStorage) Get(id string) (*models.Session, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *models.Session
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*models.Session, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) *models.Session); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Session)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SessionStorage_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type SessionStorage_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - id string
func (_e *SessionStorage_Expecter) Get(id interface{}) *SessionStorage_Get_Call {
	return &SessionStorage_Get_Call{Call: _e.mock.On("Get", id)}
}

func (_c *SessionStorage_Get_Call) Run(run func(id string)) *SessionStorage_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SessionStorage_Get_Call) Return(_a0 *models.Session, _a1 error) *SessionStorage_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SessionStorage_Get_Call) RunAndReturn(run func(string) (*models.Session, error)) *SessionStorage_Get_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with no fields
func (_m *SessionStorage) List() []models.Session {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []models.Session
	if rf, ok := ret.Get(0).(func() []models.Session); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Session)
		}
	}

	return r0
}

// SessionStorage_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type SessionStorage_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
func (_e *SessionStorage_Expecter) List() *SessionStorage_List_Call {
	return &SessionStorage_List_Call{Call: _e.mock.On("List")}
}

func (_c *SessionStorage_List_Call) Run(run func()) *SessionStorage_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SessionStorage_List_Call) Return(_a0 []models.Session) *SessionStorage_List_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SessionStorage_List_Call) RunAndReturn(run func() []models.Session) *SessionStorage_List_Call {
	_c.Call.Return(run)
	return _c
}

// NewSessionStorage creates a new instance of SessionStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSessionStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *SessionStorage {
	mock := &SessionStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
