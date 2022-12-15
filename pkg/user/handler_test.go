package user

import (
	"encoding/json"
	"errors"
	"faceit/internal/common"
	c "faceit/internal/common"
	"faceit/internal/user"
	"faceit/internal/user/api"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	baseUrl  = "/api/v1"
	usersUrl = baseUrl + "/users"
	Get      = "Get"
	List     = "List"
	Create   = "Create"
	Delete   = "Delete"
	Update   = "Update"
)

var (
	userID        = uuid.New()
	missingUserID = uuid.New()
	errorUserID   = uuid.New()
	invalidUserID = uuid.New()
)

type (
	handlerTestSuite struct {
		userSvcMock *mockUserService
		handler     Handler
		wrapper     api.ServerInterfaceWrapper
		e           *echo.Echo
		suite.Suite
	}

	scenario struct {
		name           string
		id             *string
		expectedStatus int
		prepareMock    func()
		assertMock     func()
	}
)

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}

func (s *handlerTestSuite) SetupSuite() {
	s.e = echo.New()
	s.userSvcMock = newMockUserService(s.T())
	s.handler = Handler{
		userSvc: s.userSvcMock,
	}

	// register wrapper to test OpenAPI validation as well
	s.wrapper = api.ServerInterfaceWrapper{
		Handler: s.handler,
	}

	s.e.GET(usersUrl, s.wrapper.List)
	s.e.POST(usersUrl, s.wrapper.Create)
	s.e.DELETE(usersUrl+"/:id", s.wrapper.DeleteByID)
	s.e.GET(usersUrl+"/:id", s.wrapper.GetByID)
	s.e.PATCH(usersUrl+"/:id", s.wrapper.UpdateByID)
}

func (s handlerTestSuite) TestGetByID() {
	u := user.User{
		ID:    userID,
		Email: "test@test.com",
	}
	s.userSvcMock.
		On(Get, mock.Anything, userID).
		Return(&u, nil).
		Once()

	ctx, rec := s.call(http.MethodGet, c.Ptr(userID.String()), nil)

	s.NoError(s.wrapper.GetByID(ctx))
	s.Equal(http.StatusOK, rec.Code)

	actualUser, err := asUserResponse(rec.Body.Bytes())
	s.NoError(err)
	s.Equal(userID, actualUser.Id)
	s.Equal(types.Email("test@test.com"), actualUser.Email)
}

func (s handlerTestSuite) TestGetByID_ReturnsError() {
	prepareMock := func(id uuid.UUID, returnErr error) {
		s.userSvcMock.
			On(Get, mock.Anything, id).
			Return(nil, returnErr).
			Once()
	}

	tests := []scenario{
		{
			name:           "invalid UUID format",
			id:             c.Ptr("invalid"),
			expectedStatus: http.StatusBadRequest,
			assertMock:     func() { s.userSvcMock.AssertNotCalled(s.T(), Get) },
		},
		{
			name:           "nil UUID",
			id:             c.Ptr(uuid.Nil.String()),
			expectedStatus: http.StatusBadRequest,
			prepareMock:    func() { prepareMock(uuid.Nil, user.ErrNilUUIDNotAllowed) },
		},
		{
			name:           "not found",
			id:             c.Ptr(missingUserID.String()),
			expectedStatus: http.StatusNotFound,
			prepareMock:    func() { prepareMock(missingUserID, user.ErrUserNotFound) },
		},
		{
			name:           "service error",
			id:             c.Ptr(errorUserID.String()),
			expectedStatus: http.StatusInternalServerError,
			prepareMock:    func() { prepareMock(errorUserID, errors.New("any error")) },
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			if test.prepareMock != nil {
				test.prepareMock()
			}
			ctx, _ := s.call(http.MethodGet, test.id, nil)

			err := s.wrapper.GetByID(ctx).(*echo.HTTPError)
			s.Equal(test.expectedStatus, err.Code)
			if test.assertMock != nil {
				test.assertMock()
			}
		})
	}
}

func (s handlerTestSuite) TestList() {
	pagination := common.Pagination{Page: 1, PageSize: 2}
	filter := user.User{
		FirstName: "fn",
		LastName:  "ln",
		Nickname:  "nn",
		Email:     "em",
		Country:   "uk",
	}
	expectedUsers := []user.User{
		{Email: "res1@email.com"},
		{Email: "res2@email.com"},
	}
	s.userSvcMock.
		On(List, mock.Anything, pagination, &filter).
		Return(expectedUsers, nil).
		Once()

	params := fmt.Sprintf("?page=%d&pagesize=%d&first_name=%s&last_name=%s&nickname=%s&email=%s&country=%s", 1, 2, "fn", "ln", "nn", "em", "uk")
	ctx, rec := s.call(http.MethodGet, c.Ptr(params), nil)

	s.NoError(s.wrapper.List(ctx))
	s.Equal(http.StatusOK, rec.Code)

	var res []user.User
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	s.NoError(err)
	s.Len(res, 2)
	s.Equal("res1@email.com", res[0].Email)
	s.Equal("res2@email.com", res[1].Email)
}

func (s handlerTestSuite) TestList_ReturnsErrorOnInvalidParameters() {
	invalidPaginationQuery := "?page=-1&pagesize=-1"
	invalidPagination := common.Pagination{Page: -1, PageSize: -1}
	invalidFilterQuery := "?first_name=x&last_name=x&nickname=x&email=x&country=x"
	invalidFilter := user.User{
		FirstName: "x",
		LastName:  "x",
		Nickname:  "x",
		Email:     "x",
		Country:   "x",
	}
	prepareMock := func(p common.Pagination, f user.User, returnErr error) {
		s.userSvcMock.
			On(List, mock.Anything, p, &f).
			Return(nil, returnErr).
			Once()
	}

	tests := []scenario{
		{
			name:        "invalid pagination",
			id:          c.Ptr(invalidPaginationQuery),
			prepareMock: func() { prepareMock(invalidPagination, user.User{}, user.ErrInvalidPagination) },
		},
		{
			name:        "invalid pagination",
			id:          c.Ptr(invalidFilterQuery),
			prepareMock: func() { prepareMock(common.Pagination{}, invalidFilter, user.ErrInvalidFilter) },
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.prepareMock()
			ctx, _ := s.call(http.MethodGet, test.id, nil)

			err := s.wrapper.List(ctx).(*echo.HTTPError)
			s.Equal(http.StatusBadRequest, err.Code)
		})
	}

}

func (s handlerTestSuite) TestCreate() {
	expectedUser := user.User{
		FirstName: "john",
		LastName:  "doe",
		Nickname:  "johndoe",
		Email:     "test@test.com",
		Country:   "US",
	}
	savedUser := expectedUser
	savedUser.ID = uuid.New()

	s.userSvcMock.
		On(Create, mock.Anything, expectedUser, "testpwd").
		Return(&savedUser, nil).
		Once()

	jsonIn, err := toJsonBody(expectedUser, "testpwd")
	s.NoError(err)
	ctx, rec := s.call(http.MethodPost, nil, jsonIn)

	s.NoError(s.wrapper.Create(ctx))
	s.Equal(http.StatusCreated, rec.Code)

	actualUser, err := asUserResponse(rec.Body.Bytes())
	s.NoError(err)
	s.Equal(savedUser.ID, actualUser.Id)
	s.Equal("john", actualUser.FirstName)
	s.Equal("doe", actualUser.LastName)
	s.Equal("johndoe", actualUser.Nickname)
	s.Equal(types.Email("test@test.com"), actualUser.Email)
	s.Equal("US", actualUser.Country)
}

func (s handlerTestSuite) TestCreate_ReturnsError() {
	expectedUser := user.User{
		FirstName: "john",
		LastName:  "doe",
		Nickname:  "johndoe",
		Email:     "test@test.com",
		Country:   "US",
	}

	prepareMock := func(pwd string, returnErr error) {
		s.userSvcMock.
			On(Create, mock.Anything, expectedUser, pwd).
			Return(nil, returnErr).
			Once()
	}

	tests := []scenario{
		{
			name:           "invalid data",
			expectedStatus: http.StatusBadRequest,
			prepareMock:    func() { prepareMock("0", user.ErrInvalidUserInputData) },
		},
		{
			name:           "tempered data",
			expectedStatus: http.StatusBadRequest,
			prepareMock:    func() { prepareMock("1", user.ErrNewUserWithID) },
		},
		{
			name:           "service error",
			expectedStatus: http.StatusInternalServerError,
			prepareMock:    func() { prepareMock("2", errors.New("any error")) },
		},
	}

	for i, test := range tests {
		s.Run(test.name, func() {
			if test.prepareMock != nil {
				test.prepareMock()
			}
			jsonIn, e := toJsonBody(expectedUser, strconv.Itoa(i))
			s.NoError(e)
			ctx, _ := s.call(http.MethodPost, nil, jsonIn)

			err := s.wrapper.Create(ctx).(*echo.HTTPError)
			s.Equal(test.expectedStatus, err.Code)
			if test.assertMock != nil {
				test.assertMock()
			}
		})
	}
}

func (s handlerTestSuite) TestDeleteByID() {
	s.userSvcMock.
		On(Delete, mock.Anything, userID).
		Return(nil).
		Once()

	ctx, rec := s.call(http.MethodDelete, c.Ptr(userID.String()), nil)

	s.NoError(s.wrapper.DeleteByID(ctx))
	s.Equal(http.StatusNoContent, rec.Code)
}

func (s handlerTestSuite) TestDeleteByID_ReturnsError() {
	prepareMock := func(id uuid.UUID, returnErr error) {
		s.userSvcMock.
			On(Delete, mock.Anything, id).
			Return(returnErr).
			Once()
	}

	tests := []scenario{
		{
			name:           "invalid UUID format",
			id:             c.Ptr("invalid"),
			expectedStatus: http.StatusBadRequest,
			assertMock:     func() { s.userSvcMock.AssertNotCalled(s.T(), Delete) },
		},
		{
			name:           "nil UUID",
			id:             c.Ptr(uuid.Nil.String()),
			expectedStatus: http.StatusBadRequest,
			prepareMock:    func() { prepareMock(uuid.Nil, user.ErrNilUUIDNotAllowed) },
		},
		{
			name:           "service error",
			id:             c.Ptr(errorUserID.String()),
			expectedStatus: http.StatusInternalServerError,
			prepareMock:    func() { prepareMock(errorUserID, errors.New("any error")) },
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			if test.prepareMock != nil {
				test.prepareMock()
			}
			ctx, _ := s.call(http.MethodDelete, test.id, nil)

			err := s.wrapper.DeleteByID(ctx).(*echo.HTTPError)
			s.Equal(test.expectedStatus, err.Code)
			if test.assertMock != nil {
				test.assertMock()
			}
		})
	}
}

func (s handlerTestSuite) TestUpdate() {
	id := uuid.New()
	expectedUser := user.User{
		FirstName: "john",
		LastName:  "doe",
		Nickname:  "johndoe",
		Email:     "test@test.com",
		Country:   "US",
	}
	savedUser := expectedUser
	savedUser.ID = id

	s.userSvcMock.
		On(Update, mock.Anything, id, expectedUser, "testpwd").
		Return(&savedUser, nil).
		Once()

	jsonIn, err := toJsonBody(expectedUser, "testpwd")
	s.NoError(err)
	ctx, rec := s.call(http.MethodPatch, common.Ptr(id.String()), jsonIn)

	s.NoError(s.wrapper.UpdateByID(ctx))
	s.Equal(http.StatusOK, rec.Code)

	actualUser, err := asUserResponse(rec.Body.Bytes())
	s.NoError(err)
	s.Equal(savedUser.ID, actualUser.Id)
	s.Equal("john", actualUser.FirstName)
	s.Equal("doe", actualUser.LastName)
	s.Equal("johndoe", actualUser.Nickname)
	s.Equal(types.Email("test@test.com"), actualUser.Email)
	s.Equal("US", actualUser.Country)
}

func (s handlerTestSuite) TestUpdate_ReturnsError() {
	expectedUser := user.User{
		FirstName: "john",
		LastName:  "doe",
		Nickname:  "johndoe",
		Email:     "test@test.com",
		Country:   "US",
	}

	prepareMock := func(pwd string, returnErr error) {
		s.userSvcMock.
			On(Update, mock.Anything, mock.Anything, expectedUser, pwd).
			Return(nil, returnErr).
			Once()
	}

	tests := []scenario{
		{
			name:           "invalid data",
			id:             common.Ptr(invalidUserID.String()),
			expectedStatus: http.StatusBadRequest,
			prepareMock:    func() { prepareMock("0", user.ErrInvalidUserInputData) },
		},
		{
			name:           "nil id",
			id:             common.Ptr(uuid.Nil.String()),
			expectedStatus: http.StatusBadRequest,
			prepareMock:    func() { prepareMock("1", user.ErrNilUUIDNotAllowed) },
		},
		{
			name:           "service error",
			id:             common.Ptr(errorUserID.String()),
			expectedStatus: http.StatusInternalServerError,
			prepareMock:    func() { prepareMock("2", errors.New("any error")) },
		},
	}

	for i, test := range tests {
		s.Run(test.name, func() {
			if test.prepareMock != nil {
				test.prepareMock()
			}
			jsonIn, e := toJsonBody(expectedUser, strconv.Itoa(i))
			s.NoError(e)
			ctx, _ := s.call(http.MethodPatch, test.id, jsonIn)

			err := s.wrapper.UpdateByID(ctx).(*echo.HTTPError)
			s.Equal(test.expectedStatus, err.Code)
		})
	}
}

func (s handlerTestSuite) call(method string, id *string, body io.Reader) (echo.Context, *httptest.ResponseRecorder) {
	url := usersUrl
	if id != nil {
		url += *id
	}
	req := httptest.NewRequest(method, url, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := s.e.NewContext(req, rec)
	if id != nil {
		ctx.SetParamNames("id")
		ctx.SetParamValues(*id)
	}
	return ctx, rec
}

func asUserResponse(data []byte) (*api.UserResponse, error) {
	var u *user.User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, err
	}
	return c.Ptr(toUserResponse(u)), nil
}

func toJsonBody(u user.User, password string) (io.Reader, error) {
	up := api.UserWithPassword{
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		Email:     types.Email(u.Email),
		Country:   u.Country,
		Password:  &password,
	}
	out, err := json.Marshal(&up)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(string(out)), nil
}
