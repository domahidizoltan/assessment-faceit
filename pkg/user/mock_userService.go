// Code generated by mockery v2.15.0. DO NOT EDIT.

package user

import (
	context "context"
	common "faceit/internal/common"

	internaluser "faceit/internal/user"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// mockUserService is an autogenerated mock type for the userService type
type mockUserService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, user, password
func (_m *mockUserService) Create(ctx context.Context, user internaluser.User, password string) (*internaluser.User, error) {
	ret := _m.Called(ctx, user, password)

	var r0 *internaluser.User
	if rf, ok := ret.Get(0).(func(context.Context, internaluser.User, string) *internaluser.User); ok {
		r0 = rf(ctx, user, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internaluser.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, internaluser.User, string) error); ok {
		r1 = rf(ctx, user, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *mockUserService) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *mockUserService) Get(ctx context.Context, id uuid.UUID) (*internaluser.User, error) {
	ret := _m.Called(ctx, id)

	var r0 *internaluser.User
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *internaluser.User); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internaluser.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, pagination, filters
func (_m *mockUserService) List(ctx context.Context, pagination common.Pagination, filters *internaluser.User) ([]internaluser.User, error) {
	ret := _m.Called(ctx, pagination, filters)

	var r0 []internaluser.User
	if rf, ok := ret.Get(0).(func(context.Context, common.Pagination, *internaluser.User) []internaluser.User); ok {
		r0 = rf(ctx, pagination, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]internaluser.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, common.Pagination, *internaluser.User) error); ok {
		r1 = rf(ctx, pagination, filters)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, user, password
func (_m *mockUserService) Update(ctx context.Context, id uuid.UUID, user internaluser.User, password string) (*internaluser.User, error) {
	ret := _m.Called(ctx, id, user, password)

	var r0 *internaluser.User
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, internaluser.User, string) *internaluser.User); ok {
		r0 = rf(ctx, id, user, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internaluser.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, internaluser.User, string) error); ok {
		r1 = rf(ctx, id, user, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTnewMockUserService interface {
	mock.TestingT
	Cleanup(func())
}

// newMockUserService creates a new instance of mockUserService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newMockUserService(t mockConstructorTestingTnewMockUserService) *mockUserService {
	mock := &mockUserService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
