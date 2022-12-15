package user

import (
	"context"
	"encoding/json"
	"faceit/internal/common"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const jsonType = "application/json"

type RmqEventPublisher struct {
	exchange string
	conn     amqp.Connection
	channel  amqp.Channel
}

// NewEventPublisher creates a new RabbitMQ connection to publish user related events.
func NewEventPublisher() (*RmqEventPublisher, error) {
	mqHost := common.GetEnv("RMQ_HOST", "localhost")
	mqPort := common.GetEnv("RMQ_PORT", "5672")
	mqUser := common.GetEnv("RMQ_USER", "guest")
	mqPass := common.GetEnv("RMQ_PASSWORD", "guest")

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", mqUser, mqPass, mqHost, mqPort))
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	exchange := common.GetEnv("USER_EVENT_EXCHANGE", "events.user")
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}

	return &RmqEventPublisher{
		exchange: exchange,
		conn:     *conn,
		channel:  *ch,
	}, nil
}

func (e RmqEventPublisher) close() error {
	if err := e.channel.Close(); err != nil {
		log.Err(err).Msg("failed to close RMQ channel")
	}
	return e.conn.Close()
}

// Channel returns the created RabbitMQ channel
func (e RmqEventPublisher) Channel() amqp.Channel {
	return e.channel
}

// Channel returns the created RabbitMQ exchange name for user events
func (e RmqEventPublisher) ExchangeName() string {
	return e.exchange
}

func (e RmqEventPublisher) publishCreated(ctx context.Context, userID uuid.UUID, userChanges *User) error {
	return e.publishEvent(ctx, UserEventTypeCreated, userID, userChanges)
}

func (e RmqEventPublisher) publishDeleted(ctx context.Context, userID uuid.UUID) error {
	return e.publishEvent(ctx, UserEventTypeDeleted, userID, nil)
}

func (e RmqEventPublisher) publishUpdated(ctx context.Context, userID uuid.UUID, userChanges *User) error {
	return e.publishEvent(ctx, UserEventTypeUpdated, userID, userChanges)
}

func (e RmqEventPublisher) publishPasswordChanged(ctx context.Context, userID uuid.UUID) error {
	return e.publishEvent(ctx, UserEventTypePasswordChanged, userID, nil)
}

func (e RmqEventPublisher) publishEvent(ctx context.Context, eventType UserEventType, userID uuid.UUID, userChanges *User) error {
	event := UserEvent{
		Type:        eventType,
		UserID:      userID,
		UserChanges: userChanges,
		Time:        time.Now(),
	}

	body, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		CorrelationId: common.GetCorrelationID(ctx),
		ContentType:   jsonType,
		Body:          body,
	}

	return e.channel.PublishWithContext(ctx, e.exchange, "#", false, false, msg)
}
