package server

import (
	"context"
	"faceit/internal/user"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type HealthResponse struct {
	Status   string `json:"status"`
	RabbitMQ string `json:"rabbitmq"`
	Postgres string `json:"postgres"`
}

type Health struct {
	publisher *user.RmqEventPublisher
	db        *gorm.DB
}

func NewHealth() Health {
	p, err := user.NewEventPublisher()
	if err != nil {
		log.Err(err).Msg("failed to create RabbitMQ connection")
	}

	r, err := user.NewRepository()
	if err != nil {
		log.Err(err).Msg("failed to create Postgres connection")
	}

	return Health{
		publisher: p,
		db:        r.GetDB(),
	}
}

func (h Health) Check(echoCtx echo.Context) error {
	_, cancel := context.WithTimeout(echoCtx.Request().Context(), time.Second)
	defer cancel()

	rmqConn := true
	c := h.publisher.Channel()
	if err := c.ExchangeDeclarePassive(h.publisher.ExchangeName(), "topic", true, false, false, false, nil); err != nil {
		rmqConn = false
		log.Err(err).Msg("RabbitMQ connection is down")
	}

	dbConn := true
	if err := h.db.Exec("select 1").Error; err != nil {
		dbConn = false
		log.Err(err).Msg("Postgres connection is down")
	}

	status := http.StatusOK
	if !rmqConn || !dbConn {
		status = http.StatusServiceUnavailable
	}

	resp := HealthResponse{
		Status:   getStatus(status == http.StatusOK),
		RabbitMQ: getStatus(rmqConn),
		Postgres: getStatus(dbConn),
	}

	echoCtx.JSON(status, resp)
	return nil
}

func getStatus(b bool) string {
	s := "UP"
	if !b {
		s = "DOWN"
	}
	return s
}
