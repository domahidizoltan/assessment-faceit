//go:build integration
// +build integration

package user

import (
	"context"
	"encoding/json"
	"faceit/internal/user"
	"faceit/internal/user/api"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type (
	usersAPITestSuite struct {
		handler      Handler
		wrapper      api.ServerInterfaceWrapper
		e            *echo.Echo
		consumedMsgs <-chan amqp.Delivery
		db           *gorm.DB
		suite.Suite
	}
)

func TestUsersAPITestSuite(t *testing.T) {
	suite.Run(t, new(usersAPITestSuite))
}

func (s *usersAPITestSuite) SetupSuite() {
	s.e = echo.New()
	svc, err := user.NewService()
	s.Require().NoError(err)
	s.handler = Handler{
		userSvc: svc,
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

	repo, err := user.NewRepository()
	if err != nil {
		s.Require().NoError(err)
	}
	s.db = repo.GetDB()

	// init RMQ consumer
	p, err := user.NewEventPublisher()
	s.Require().NoError(err)

	channel := p.Channel()
	q, err := channel.QueueDeclare("itest", true, false, false, false, nil)
	if err != nil {
		s.T().Log("queue already exists")
		return
	}

	if err := channel.QueueBind(q.Name, "#", "events.user", false, nil); err != nil {
		s.Require().NoError(err)
	}

	s.consumedMsgs, err = channel.Consume(q.Name, "apitest-consumer", true, true, false, false, nil)
	if err != nil {
		s.Require().NoError(err)
	}
}

func (s usersAPITestSuite) TestUserCreation() {
	// delete test user from DB if exists
	if err := s.db.Where(user.User{Email: "api@email.com"}).Delete(user.User{}).Error; err != nil {
		s.Require().NoError(err)
	}

	// prepare request to create a new user
	createUser := api.UserWithPassword{
		FirstName: "fnAPI",
		LastName:  "lnAPI",
		Nickname:  "nnAPI",
		Email:     "api@email.com",
		Country:   "HU",
	}

	b, err := json.Marshal(createUser)
	if err != nil {
		s.Require().NoError(err)
	}
	body := strings.NewReader(string(b))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := s.e.NewContext(req, rec)

	// create user
	s.wrapper.Create(ctx)
	s.Equal(http.StatusCreated, ctx.Response().Status)

	var newUser *user.User
	if err := json.Unmarshal(rec.Body.Bytes(), &newUser); err != nil {
		s.Require().NoError(err)
	}

	// check if it was created in the db
	var dbUser *user.User
	if err := s.db.Take(&dbUser, newUser.ID).Error; err != nil {
		s.Require().NoError(err)
	}
	s.NotNil(dbUser)

	// get published user event
	userEvent, err := s.getUserEvent()
	s.NoError(err)
	s.NotNil(userEvent)
	s.Equal(user.UserEventTypeCreated, userEvent.Type)
	s.Equal(newUser.ID, userEvent.UserID)
	s.True(time.Now().Sub(userEvent.Time) < 5*time.Second)

	u := userEvent.UserChanges
	s.NotNil(u)
	s.Equal("fnAPI", u.FirstName)
	s.Equal("lnAPI", u.LastName)
	s.Equal("nnAPI", u.Nickname)
	s.Equal("api@email.com", u.Email)
	s.Equal("HU", u.Country)

	// clean up new user from db
	if err := s.db.Where(user.User{Email: "api@email.com"}).Delete(user.User{}).Error; err != nil {
		s.Require().NoError(err)
	}
}

func (s usersAPITestSuite) getUserEvent() (*user.UserEvent, error) {
	select {
	case <-time.After(3 * time.Second):
		return nil, context.DeadlineExceeded
	default:
		for msg := range s.consumedMsgs {
			var event *user.UserEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				return nil, err
			}

			return event, nil
		}
	}

	return nil, nil
}
