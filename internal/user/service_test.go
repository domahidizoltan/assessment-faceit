package user

import (
	"errors"
	"faceit/internal/common"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	findByID               = "findByID"
	create                 = "create"
	list                   = "list"
	deleteByID             = "deleteByID"
	update                 = "update"
	updatePass             = "updatePassword"
	publishDeleted         = "publishDeleted"
	publishCreated         = "publishCreated"
	publishUpdated         = "publishUpdated"
	publishPasswordChanged = "publishPasswordChanged"

	testpwd     = "testpwd"
	testpwdHash = "a85b6a20813c31a8b1b3f3618da796271c9aa293b3f809873053b21aec501087"
)

var validUser = User{
	FirstName: "john",
	LastName:  "doe",
	Nickname:  "johndoe",
	Email:     "johndoe@email.com",
	Country:   "US",
}

type (
	serviceTestSuite struct {
		repoMock      *mockRepository
		publisherMock *mockEventPublisher
		service       Service
		suite.Suite
	}
)

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(serviceTestSuite))
}

func (s *serviceTestSuite) SetupSuite() {
	s.repoMock = newMockRepository(s.T())
	s.publisherMock = newMockEventPublisher(s.T())
	s.service = Service{
		repository:     s.repoMock,
		eventPublisher: s.publisherMock,
	}
}

func (s serviceTestSuite) TestGet() {
	id := uuid.New()
	u := User{
		ID:        id,
		FirstName: "test",
	}
	s.repoMock.
		On(findByID, mock.Anything, id).
		Return(&u, nil).
		Once()

	actualUser, err := s.service.Get(nil, id)
	s.NoError(err)
	s.Equal(id, actualUser.ID)
	s.Equal("test", actualUser.FirstName)
}

func (s serviceTestSuite) TestGet_ReturnsError() {
	id := uuid.New()
	s.repoMock.
		On(findByID, mock.Anything, id).
		Return(nil, ErrUserNotFound).
		Once()

	actualUser, err := s.service.Get(nil, id)
	s.Nil(actualUser)
	s.ErrorIs(err, ErrUserNotFound)
}

func (s serviceTestSuite) TestGet_ReturnsErrorOnNilUUID() {
	_, err := s.service.Get(nil, uuid.Nil)
	s.ErrorIs(err, ErrNilUUIDNotAllowed)
	s.repoMock.AssertNotCalled(s.T(), findByID)
}

func (s serviceTestSuite) TestList() {
	pag := common.Pagination{Page: 1, PageSize: 2}
	filter := User{FirstName: "test"}
	expectedUsers := []User{
		{FirstName: "Test", LastName: "LastName"},
	}

	s.repoMock.
		On(list, mock.Anything, pag, &filter).
		Return(expectedUsers, nil).
		Once()

	results, err := s.service.List(nil, pag, &filter)
	s.NoError(err)
	s.Len(results, 1)
	s.Equal("Test", results[0].FirstName)
	s.Equal("LastName", results[0].LastName)
}

func (s serviceTestSuite) TestList_ReturnsError() {
	validPagination := common.Pagination{Page: 1, PageSize: 2}
	changeValidPagination := func(change func(*common.Pagination)) common.Pagination {
		p := validPagination
		change(&p)
		return p
	}
	validFilter := User{
		FirstName: "fn",
		LastName:  "ln",
		Nickname:  "nn",
		Email:     "em",
		Country:   "uk",
	}
	changeValidFilter := func(change func(*User)) User {
		f := validFilter
		change(&f)
		return f
	}

	for _, test := range []struct {
		name          string
		p             common.Pagination
		filter        User
		expectedError error
	}{
		{
			name:          "negative page",
			p:             changeValidPagination(func(p *common.Pagination) { p.Page = -1 }),
			filter:        validFilter,
			expectedError: ErrInvalidPagination,
		},
		{
			name:          "negative pageSize",
			p:             changeValidPagination(func(p *common.Pagination) { p.PageSize = -1 }),
			filter:        validFilter,
			expectedError: ErrInvalidPagination,
		},
		{
			name:          "short first name",
			p:             validPagination,
			filter:        changeValidFilter(func(u *User) { u.FirstName = "x" }),
			expectedError: ErrInvalidFilter,
		},
		{
			name:          "short last name",
			p:             validPagination,
			filter:        changeValidFilter(func(u *User) { u.LastName = "x" }),
			expectedError: ErrInvalidFilter,
		},
		{
			name:          "short nickname",
			p:             validPagination,
			filter:        changeValidFilter(func(u *User) { u.Nickname = "x" }),
			expectedError: ErrInvalidFilter,
		},
		{
			name:          "short email",
			p:             validPagination,
			filter:        changeValidFilter(func(u *User) { u.Email = "x" }),
			expectedError: ErrInvalidFilter,
		},
		{
			name:          "short country",
			p:             validPagination,
			filter:        changeValidFilter(func(u *User) { u.Country = "x" }),
			expectedError: ErrInvalidFilter,
		},
	} {
		s.Run(test.name, func() {
			_, err := s.service.List(nil, test.p, &test.filter)
			s.ErrorIs(err, test.expectedError)
			s.repoMock.AssertNotCalled(s.T(), list)
		})
	}
}

func (s serviceTestSuite) TestDelete() {
	id := uuid.New()
	s.repoMock.
		On(deleteByID, mock.Anything, id).
		Return(nil).
		Once()
	s.publisherMock.
		On(publishDeleted, mock.Anything, id).
		Return(nil).
		Once()

	s.NoError(s.service.Delete(nil, id))
}

func (s serviceTestSuite) TestDelete_ReturnsError() {
	id := uuid.New()
	s.repoMock.
		On(deleteByID, mock.Anything, id).
		Return(errors.New("any error")).
		Once()

	s.Error(s.service.Delete(nil, id))
	s.publisherMock.AssertNotCalled(s.T(), publishDeleted)
}

func (s serviceTestSuite) TestDelete_ReturnsErrorOnNilUUID() {
	err := s.service.Delete(nil, uuid.Nil)
	s.ErrorIs(err, ErrNilUUIDNotAllowed)
	s.repoMock.AssertNotCalled(s.T(), deleteByID)
}

func (s serviceTestSuite) TestCreate() {
	userIn := validUser
	userIn.Country = "us"
	createdUser := &User{ID: uuid.New(), Nickname: "johndoe", Country: "US"}
	s.repoMock.
		On(create, mock.Anything, validUser, testpwdHash).
		Return(createdUser, nil).
		Once()
	s.publisherMock.
		On(publishCreated, mock.Anything, createdUser.ID, createdUser).
		Return(nil).
		Once()

	newUser, err := s.service.Create(nil, userIn, testpwd)
	s.NoError(err)
	s.Equal(createdUser.ID, newUser.ID)
}

func (s serviceTestSuite) TestCreate_ReturnsError() {
	s.repoMock.
		On(create, mock.Anything, validUser, mock.Anything).
		Return(nil, errors.New("any error")).
		Once()

	_, err := s.service.Create(nil, validUser, "")
	s.Error(err)
	s.publisherMock.AssertNotCalled(s.T(), publishCreated)
}

func (s serviceTestSuite) TestCreate_ReturnsErrorOnUserValidation() {
	makeInvalidUser := func(change func(*User)) User {
		changeUser := validUser
		change(&changeUser)
		return changeUser
	}

	for i, change := range []func(*User){
		func(u *User) { u.ID = uuid.New() },
		func(u *User) { u.FirstName = "x" },
		func(u *User) { u.LastName = "x" },
		func(u *User) { u.Nickname = "x" },
		func(u *User) { u.Country = "x" },
		func(u *User) { u.Email = "x" },
	} {
		s.Run(fmt.Sprintf("scenario %d", i), func() {
			expectedError := ErrInvalidUserInputData
			if i == 0 {
				expectedError = ErrNewUserWithID
			}
			userIn := makeInvalidUser(change)
			_, err := s.service.Create(nil, userIn, "")
			s.ErrorIs(err, expectedError)
			s.repoMock.AssertNotCalled(s.T(), publishCreated)
		})
	}
}

func (s serviceTestSuite) TestUpdate_OnlyUser() {
	id := uuid.New()
	s.repoMock.
		On(update, mock.Anything, id, validUser).
		Return(&validUser, nil).
		Once()
	s.publisherMock.
		On(publishUpdated, mock.Anything, id, &validUser).
		Return(nil).
		Once()

	updatedUser, err := s.service.Update(nil, id, validUser, "")
	s.NoError(err)
	s.Equal(validUser.Email, updatedUser.Email)
	s.repoMock.AssertNotCalled(s.T(), updatePass)
	s.publisherMock.AssertNotCalled(s.T(), publishPasswordChanged)
}

func (s serviceTestSuite) TestUpdate_OnlyPassword() {
	id := uuid.New()
	s.repoMock.
		On(updatePass, mock.Anything, id, testpwdHash).
		Return(nil).
		Once()
	s.publisherMock.
		On(publishPasswordChanged, mock.Anything, id).
		Return(nil).
		Once()

	updatedUser, err := s.service.Update(nil, id, User{}, testpwd)
	s.NoError(err)
	s.Nil(updatedUser)
	s.repoMock.AssertNotCalled(s.T(), update)
	s.publisherMock.AssertNotCalled(s.T(), publishUpdated)
}

func (s serviceTestSuite) TestUpdate_UserAndPassword() {
	id := uuid.New()
	s.repoMock.
		On(update, mock.Anything, id, validUser).
		Return(&validUser, nil).
		Once()
	s.publisherMock.
		On(publishUpdated, mock.Anything, id, &validUser).
		Return(nil).
		Once()
	s.repoMock.
		On(updatePass, mock.Anything, id, testpwdHash).
		Return(nil).
		Once()
	s.publisherMock.
		On(publishPasswordChanged, mock.Anything, id).
		Return(nil).
		Once()

	newUser, err := s.service.Update(nil, id, validUser, testpwd)
	s.NoError(err)
	s.Equal(validUser.Email, newUser.Email)
}

func (s serviceTestSuite) TestUpdate_RetrurnError_WhenChangesUser() {
	id := uuid.New()
	s.repoMock.
		On(update, mock.Anything, id, validUser).
		Return(nil, errors.New("user change error")).
		Once()

	_, err := s.service.Update(nil, id, validUser, "")
	s.ErrorContains(err, "user change error")
	s.repoMock.AssertNotCalled(s.T(), updatePass)
}

func (s serviceTestSuite) TestUpdate_RetrurnError_WhenChangesPassword() {
	id := uuid.New()
	s.repoMock.
		On(updatePass, mock.Anything, id, testpwdHash).
		Return(errors.New("password change error")).
		Once()

	_, err := s.service.Update(nil, id, User{}, testpwd)
	s.ErrorContains(err, "password change error")
	s.repoMock.AssertNotCalled(s.T(), update)
}

func (s serviceTestSuite) TestUpdate_RetrurnErrorOnUserValidation() {
	makeInvalidUser := func(change func(*User)) User {
		changeUser := validUser
		change(&changeUser)
		return changeUser
	}

	for i, change := range []func(*User){
		func(u *User) { u.FirstName = "x" },
		func(u *User) { u.LastName = "x" },
		func(u *User) { u.Nickname = "x" },
		func(u *User) { u.Country = "x" },
		func(u *User) { u.Email = "x" },
	} {
		s.Run(fmt.Sprintf("scenario %d", i), func() {
			userIn := makeInvalidUser(change)
			_, err := s.service.Update(nil, uuid.New(), userIn, "")
			s.ErrorIs(err, ErrInvalidUserInputData)
			s.repoMock.AssertNotCalled(s.T(), update)
			s.repoMock.AssertNotCalled(s.T(), updatePass)
		})
	}
}

func (s serviceTestSuite) TestUpdate_RetrurnErrorOnNilUUID() {
	_, err := s.service.Update(nil, uuid.Nil, User{}, "")
	s.ErrorIs(err, ErrNilUUIDNotAllowed)
	s.repoMock.AssertNotCalled(s.T(), update)
	s.repoMock.AssertNotCalled(s.T(), updatePass)
}
