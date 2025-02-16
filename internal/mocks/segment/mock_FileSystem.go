// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	io "io"
	fs "io/fs"

	segment "github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
	mock "github.com/stretchr/testify/mock"
)

// FileSystem is an autogenerated mock type for the FileSystem type
type FileSystem struct {
	mock.Mock
}

type FileSystem_Expecter struct {
	mock *mock.Mock
}

func (_m *FileSystem) EXPECT() *FileSystem_Expecter {
	return &FileSystem_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: name
func (_m *FileSystem) Create(name string) (io.ReadWriteCloser, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 io.ReadWriteCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (io.ReadWriteCloser, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) io.ReadWriteCloser); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadWriteCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FileSystem_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type FileSystem_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - name string
func (_e *FileSystem_Expecter) Create(name interface{}) *FileSystem_Create_Call {
	return &FileSystem_Create_Call{Call: _e.mock.On("Create", name)}
}

func (_c *FileSystem_Create_Call) Run(run func(name string)) *FileSystem_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *FileSystem_Create_Call) Return(_a0 io.ReadWriteCloser, _a1 error) *FileSystem_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *FileSystem_Create_Call) RunAndReturn(run func(string) (io.ReadWriteCloser, error)) *FileSystem_Create_Call {
	_c.Call.Return(run)
	return _c
}

// MkdirAll provides a mock function with given fields: path, perm
func (_m *FileSystem) MkdirAll(path string, perm fs.FileMode) error {
	ret := _m.Called(path, perm)

	if len(ret) == 0 {
		panic("no return value specified for MkdirAll")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, fs.FileMode) error); ok {
		r0 = rf(path, perm)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FileSystem_MkdirAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MkdirAll'
type FileSystem_MkdirAll_Call struct {
	*mock.Call
}

// MkdirAll is a helper method to define mock.On call
//   - path string
//   - perm fs.FileMode
func (_e *FileSystem_Expecter) MkdirAll(path interface{}, perm interface{}) *FileSystem_MkdirAll_Call {
	return &FileSystem_MkdirAll_Call{Call: _e.mock.On("MkdirAll", path, perm)}
}

func (_c *FileSystem_MkdirAll_Call) Run(run func(path string, perm fs.FileMode)) *FileSystem_MkdirAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(fs.FileMode))
	})
	return _c
}

func (_c *FileSystem_MkdirAll_Call) Return(_a0 error) *FileSystem_MkdirAll_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FileSystem_MkdirAll_Call) RunAndReturn(run func(string, fs.FileMode) error) *FileSystem_MkdirAll_Call {
	_c.Call.Return(run)
	return _c
}

// Open provides a mock function with given fields: name
func (_m *FileSystem) Open(name string) (io.ReadWriteCloser, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Open")
	}

	var r0 io.ReadWriteCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (io.ReadWriteCloser, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) io.ReadWriteCloser); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadWriteCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FileSystem_Open_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Open'
type FileSystem_Open_Call struct {
	*mock.Call
}

// Open is a helper method to define mock.On call
//   - name string
func (_e *FileSystem_Expecter) Open(name interface{}) *FileSystem_Open_Call {
	return &FileSystem_Open_Call{Call: _e.mock.On("Open", name)}
}

func (_c *FileSystem_Open_Call) Run(run func(name string)) *FileSystem_Open_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *FileSystem_Open_Call) Return(_a0 io.ReadWriteCloser, _a1 error) *FileSystem_Open_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *FileSystem_Open_Call) RunAndReturn(run func(string) (io.ReadWriteCloser, error)) *FileSystem_Open_Call {
	_c.Call.Return(run)
	return _c
}

// ReadDir provides a mock function with given fields: dirname
func (_m *FileSystem) ReadDir(dirname string) ([]fs.DirEntry, error) {
	ret := _m.Called(dirname)

	if len(ret) == 0 {
		panic("no return value specified for ReadDir")
	}

	var r0 []fs.DirEntry
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]fs.DirEntry, error)); ok {
		return rf(dirname)
	}
	if rf, ok := ret.Get(0).(func(string) []fs.DirEntry); ok {
		r0 = rf(dirname)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]fs.DirEntry)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(dirname)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FileSystem_ReadDir_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReadDir'
type FileSystem_ReadDir_Call struct {
	*mock.Call
}

// ReadDir is a helper method to define mock.On call
//   - dirname string
func (_e *FileSystem_Expecter) ReadDir(dirname interface{}) *FileSystem_ReadDir_Call {
	return &FileSystem_ReadDir_Call{Call: _e.mock.On("ReadDir", dirname)}
}

func (_c *FileSystem_ReadDir_Call) Run(run func(dirname string)) *FileSystem_ReadDir_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *FileSystem_ReadDir_Call) Return(_a0 []fs.DirEntry, _a1 error) *FileSystem_ReadDir_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *FileSystem_ReadDir_Call) RunAndReturn(run func(string) ([]fs.DirEntry, error)) *FileSystem_ReadDir_Call {
	_c.Call.Return(run)
	return _c
}

// Remove provides a mock function with given fields: name
func (_m *FileSystem) Remove(name string) error {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Remove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FileSystem_Remove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Remove'
type FileSystem_Remove_Call struct {
	*mock.Call
}

// Remove is a helper method to define mock.On call
//   - name string
func (_e *FileSystem_Expecter) Remove(name interface{}) *FileSystem_Remove_Call {
	return &FileSystem_Remove_Call{Call: _e.mock.On("Remove", name)}
}

func (_c *FileSystem_Remove_Call) Run(run func(name string)) *FileSystem_Remove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *FileSystem_Remove_Call) Return(_a0 error) *FileSystem_Remove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FileSystem_Remove_Call) RunAndReturn(run func(string) error) *FileSystem_Remove_Call {
	_c.Call.Return(run)
	return _c
}

// Stat provides a mock function with given fields: name
func (_m *FileSystem) Stat(name string) (segment.Sizer, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Stat")
	}

	var r0 segment.Sizer
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (segment.Sizer, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) segment.Sizer); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(segment.Sizer)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FileSystem_Stat_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stat'
type FileSystem_Stat_Call struct {
	*mock.Call
}

// Stat is a helper method to define mock.On call
//   - name string
func (_e *FileSystem_Expecter) Stat(name interface{}) *FileSystem_Stat_Call {
	return &FileSystem_Stat_Call{Call: _e.mock.On("Stat", name)}
}

func (_c *FileSystem_Stat_Call) Run(run func(name string)) *FileSystem_Stat_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *FileSystem_Stat_Call) Return(_a0 segment.Sizer, _a1 error) *FileSystem_Stat_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *FileSystem_Stat_Call) RunAndReturn(run func(string) (segment.Sizer, error)) *FileSystem_Stat_Call {
	_c.Call.Return(run)
	return _c
}

// NewFileSystem creates a new instance of FileSystem. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFileSystem(t interface {
	mock.TestingT
	Cleanup(func())
}) *FileSystem {
	mock := &FileSystem{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
