// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	service "github.com/hmoragrega/paybile/service"
)

// LoginService is an autogenerated mock type for the LoginService type
type LoginService struct {
	mock.Mock
}

// Login provides a mock function with given fields: ctx, login, password
func (_m *LoginService) Login(ctx context.Context, login string, password string) (service.User, error) {
	ret := _m.Called(ctx, login, password)

	var r0 service.User
	if rf, ok := ret.Get(0).(func(context.Context, string, string) service.User); ok {
		r0 = rf(ctx, login, password)
	} else {
		r0 = ret.Get(0).(service.User)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, login, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
