package server

import (
	"faceit/internal/user/api"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func HTTPErrorHandler(err error, ctx echo.Context) {
	msg := err.Error()
	code := ctx.Response().Status
	if e, ok := err.(*echo.HTTPError); ok {
		msg = e.Message.(string)
		code = e.Code
	}

	e := api.Error{
		CorrelationId: getCorrelationId(ctx),
		Message:       msg,
		Status:        code,
		Time:          time.Now(),
	}
	log.Err(err).Msgf("%+v", e)
	ctx.JSON(e.Status, e)
}

func getCorrelationId(ctx echo.Context) uuid.UUID {
	correlationID := ctx.Request().Header.Get(echo.HeaderXRequestID)
	if correlationID == "" {
		correlationID = uuid.NewString()
		log.Info().Msgf("creating missing correlationID %s", correlationID)
		ctx.Request().Header.Set(echo.HeaderXRequestID, correlationID)
	}

	id, err := uuid.Parse(correlationID)
	if err != nil {
		cID := uuid.New()
		ctx.Request().Header.Set(echo.HeaderXRequestID, cID.String())
		return cID
	}

	return id
}
