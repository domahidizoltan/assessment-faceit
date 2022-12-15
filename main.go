package main

import (
	"faceit/internal/common"
	"faceit/internal/user/api"
	srv "faceit/pkg/server"
	"faceit/pkg/user"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	server := echo.New()
	server.Use(srv.RequestIDMiddleware, srv.LoggerMiddleware, srv.CorsMiddleware)
	server.HTTPErrorHandler = srv.HTTPErrorHandler

	usersHandler, err := user.NewHandler()
	if err != nil {
		log.Fatal().Msgf("failed to create users handler: %+v", err)
	}
	api.RegisterHandlersWithBaseURL(server, usersHandler, "api/v1")

	health := srv.NewHealth()
	server.GET("/health", health.Check)

	host := common.GetEnv("SERVER_HOST", "localhost")
	port := common.GetEnv("SERVER_PORT", "8000")

	if err := server.Start(fmt.Sprintf("%s:%s", host, port)); err != nil {
		log.Fatal().Msgf("failed to start server: %+v", err)
	}
}
