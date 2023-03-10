// Code generated by mockery v2.15.0. DO NOT EDIT.

package api

import (
	echo "github.com/labstack/echo/v4"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockServerInterface is an autogenerated mock type for the ServerInterface type
type MockServerInterface struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx
func (_m *MockServerInterface) Create(ctx echo.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(echo.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteByID provides a mock function with given fields: ctx, id
func (_m *MockServerInterface) DeleteByID(ctx echo.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(echo.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *MockServerInterface) GetByID(ctx echo.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(echo.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// List provides a mock function with given fields: ctx, params
func (_m *MockServerInterface) List(ctx echo.Context, params ListParams) error {
	ret := _m.Called(ctx, params)

	var r0 error
	if rf, ok := ret.Get(0).(func(echo.Context, ListParams) error); ok {
		r0 = rf(ctx, params)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateByID provides a mock function with given fields: ctx, id
func (_m *MockServerInterface) UpdateByID(ctx echo.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(echo.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockServerInterface interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockServerInterface creates a new instance of MockServerInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockServerInterface(t mockConstructorTestingTNewMockServerInterface) *MockServerInterface {
	mock := &MockServerInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
