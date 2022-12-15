//go:build integration
// +build integration

package user

import (
	"context"
	"encoding/json"
	"faceit/internal/common"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
)

type (
	eventPublisherTestSuite struct {
		publisher    RmqEventPublisher
		consumedMsgs <-chan amqp091.Delivery
		suite.Suite
	}
)

func TestEventPublisherTestSuite(t *testing.T) {
	suite.Run(t, new(eventPublisherTestSuite))
}

func (s *eventPublisherTestSuite) SetupSuite() {
	p, err := NewEventPublisher()
	s.Require().NoError(err)
	s.publisher = *p

	q, err := p.channel.QueueDeclare("itest", true, false, false, false, nil)
	if err != nil {
		s.T().Log("queue already exists")
		return
	}

	if err := p.channel.QueueBind(q.Name, "#", "events.user", false, nil); err != nil {
		s.Require().NoError(err)
	}

	s.consumedMsgs, err = p.channel.Consume(q.Name, "test-consumer", true, true, false, false, nil)
	if err != nil {
		s.Require().NoError(err)
	}
}

func (s *eventPublisherTestSuite) SetupTest() {
	total, err := s.publisher.channel.QueuePurge("itest", false)
	s.Require().NoError(err)
	s.T().Logf("purged %d messages from test queue", total)
}

func (s eventPublisherTestSuite) TestPublishCreated() {
	correlationID := uuid.New()
	ctx := context.WithValue(context.TODO(), common.CorrelationID, correlationID)
	user := User{
		ID:       uuid.New(),
		Nickname: "johndoe",
	}

	err := s.publisher.publishCreated(ctx, user.ID, &user)
	s.NoError(err)

	consumedEvent, consumedCorrelationID, err := s.getUserEvent()
	s.NoError(err)
	s.Equal(correlationID, consumedCorrelationID)
	s.Equal(UserEventTypeCreated, consumedEvent.Type)
	s.Equal(user.ID, consumedEvent.UserID)
	s.Equal("johndoe", consumedEvent.UserChanges.Nickname)
}

func (s eventPublisherTestSuite) TestPublishDeleted() {
	correlationID := uuid.New()
	ctx := context.WithValue(context.TODO(), common.CorrelationID, correlationID)
	userID := uuid.New()

	err := s.publisher.publishDeleted(ctx, userID)
	s.NoError(err)

	consumedEvent, consumedCorrelationID, err := s.getUserEvent()
	s.NoError(err)
	s.Equal(correlationID, consumedCorrelationID)
	s.Equal(UserEventTypeDeleted, consumedEvent.Type)
	s.Equal(userID, consumedEvent.UserID)
}

func (s eventPublisherTestSuite) TestPublishUpdated() {
	correlationID := uuid.New()
	ctx := context.WithValue(context.TODO(), common.CorrelationID, correlationID)
	user := User{
		ID:    uuid.New(),
		Email: "new@email.com",
	}

	err := s.publisher.publishUpdated(ctx, user.ID, &user)
	s.NoError(err)

	consumedEvent, consumedCorrelationID, err := s.getUserEvent()
	s.NoError(err)
	s.Equal(correlationID, consumedCorrelationID)
	s.Equal(UserEventTypeUpdated, consumedEvent.Type)
	s.Equal(user.ID, consumedEvent.UserID)
	s.Equal("new@email.com", consumedEvent.UserChanges.Email)
}

func (s eventPublisherTestSuite) TestPasswordChanged() {
	correlationID := uuid.New()
	ctx := context.WithValue(context.TODO(), common.CorrelationID, correlationID)
	userID := uuid.New()

	err := s.publisher.publishPasswordChanged(ctx, userID)
	s.NoError(err)

	consumedEvent, consumedCorrelationID, err := s.getUserEvent()
	s.NoError(err)
	s.Equal(correlationID, consumedCorrelationID)
	s.Equal(UserEventTypePasswordChanged, consumedEvent.Type)
	s.Equal(userID, consumedEvent.UserID)
}

func (s eventPublisherTestSuite) getUserEvent() (*UserEvent, uuid.UUID, error) {
	select {
	case <-time.After(3 * time.Second):
		return nil, uuid.Nil, context.DeadlineExceeded
	default:
		for msg := range s.consumedMsgs {
			cID, err := uuid.Parse(msg.CorrelationId)
			if err != nil {
				s.T().Logf("failed to parse correlation id %s", msg.CorrelationId)
				return nil, uuid.Nil, err
			}

			var event *UserEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				return nil, uuid.Nil, err
			}

			return event, cID, nil
		}
	}

	return nil, uuid.Nil, nil
}
