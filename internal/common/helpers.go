package common

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func Ptr[T any](in T) *T {
	return &in
}

func GetEnv(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func GetEchoCorrelationID(ctx echo.Context) string {
	correlationID := ""
	if ctx == nil {
		return correlationID
	}

	cID := ctx.Get(echo.HeaderXCorrelationID)
	switch c := cID.(type) {
	case string:
		correlationID = c
	case uuid.UUID:
		correlationID = c.String()
	}

	return correlationID
}

func GetCorrelationID(ctx context.Context) string {
	correlationID := ""
	if ctx == nil {
		return correlationID
	}

	cID := ctx.Value(CorrelationID)
	switch c := cID.(type) {
	case string:
		correlationID = c
	case uuid.UUID:
		correlationID = c.String()
	}
	return correlationID
}
