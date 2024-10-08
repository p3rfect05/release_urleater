// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	echo "github.com/labstack/echo/v4"

	http "net/http"

	mock "github.com/stretchr/testify/mock"

	sessions "github.com/gorilla/sessions"
)

// SessionStore is an autogenerated mock type for the SessionStore type
type SessionStore struct {
	mock.Mock
}

// Get provides a mock function with given fields: r, key
func (_m *SessionStore) Get(r *http.Request, key string) (*sessions.Session, error) {
	ret := _m.Called(r, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *sessions.Session
	var r1 error
	if rf, ok := ret.Get(0).(func(*http.Request, string) (*sessions.Session, error)); ok {
		return rf(r, key)
	}
	if rf, ok := ret.Get(0).(func(*http.Request, string) *sessions.Session); ok {
		r0 = rf(r, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sessions.Session)
		}
	}

	if rf, ok := ret.Get(1).(func(*http.Request, string) error); ok {
		r1 = rf(r, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveEmailFromSession provides a mock function with given fields: c
func (_m *SessionStore) RetrieveEmailFromSession(c echo.Context) (string, error) {
	ret := _m.Called(c)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveEmailFromSession")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(echo.Context) (string, error)); ok {
		return rf(c)
	}
	if rf, ok := ret.Get(0).(func(echo.Context) string); ok {
		r0 = rf(c)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(echo.Context) error); ok {
		r1 = rf(c)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: c, email, session
func (_m *SessionStore) Save(c echo.Context, email string, session *sessions.Session) error {
	ret := _m.Called(c, email, session)

	if len(ret) == 0 {
		panic("no return value specified for Save")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(echo.Context, string, *sessions.Session) error); ok {
		r0 = rf(c, email, session)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSessionStore creates a new instance of SessionStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSessionStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *SessionStore {
	mock := &SessionStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
