package server

import (
	"faceit/internal/common"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

var (
	RequestIDMiddleware = middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		RequestIDHandler: func(ctx echo.Context, s string) {
			cID := ctx.Request().Header.Get(echo.HeaderXRequestID)
			if cID != "" {
				log.Info().Msgf("keeping existing correlationID %s", cID)
				ctx.Set(common.CorrelationID, cID)
				return
			}
			cID = uuid.NewString()
			ctx.Request().Header.Set(echo.HeaderXRequestID, cID)
			ctx.Set(common.CorrelationID, cID)
		},
	})

	LoggerMiddleware = middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status} correlation_id=${header:X-Request-Id}\n",
	})

	CorsMiddleware = middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	})
)
