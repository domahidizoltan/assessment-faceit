package user

import (
	"context"
	"faceit/internal/common"
	"faceit/internal/user"
	"faceit/internal/user/api"
	"net/http"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v4"
)

type (
	userService interface {
		Create(ctx context.Context, user user.User, password string) (*user.User, error)
		Get(ctx context.Context, id uuid.UUID) (*user.User, error)
		Update(ctx context.Context, id uuid.UUID, user user.User, password string) (*user.User, error)
		Delete(ctx context.Context, id uuid.UUID) error
		List(ctx context.Context, pagination common.Pagination, filters *user.User) ([]user.User, error)
	}

	Handler struct {
		timeout time.Duration
		userSvc userService
	}
)

func NewHandler() (*Handler, error) {
	svc, err := user.NewService()
	if err != nil {
		return nil, err
	}

	t := common.GetEnv("REQUEST_TIMEOUT", "5s")
	timeout, err := time.ParseDuration(t)
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse request timeout")
		timeout = 5 * time.Second
	}

	return &Handler{
		timeout: timeout,
		userSvc: svc,
	}, nil
}

func (h Handler) List(ctx echo.Context, params api.ListParams) error {
	c, cancel := h.contextWithTimeout(ctx)
	defer cancel()

	pagination := getListPagination(params)
	filter := getListFilter(params)

	results, err := h.userSvc.List(c, pagination, &filter)
	if err != nil {
		log.Err(err).
			Str("operation", "List").
			Str("params", ctx.QueryString()).
			Str(common.CorrelationID, common.GetCorrelationID(c)).
			Send()

		switch err {
		case user.ErrInvalidPagination, user.ErrInvalidFilter:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	users := []api.UserResponse{}
	for _, r := range results {
		users = append(users, toUserResponse(&r))
	}
	return ctx.JSON(http.StatusOK, users)
}

func (h Handler) Create(ctx echo.Context) error {
	c, cancel := h.contextWithTimeout(ctx)
	defer cancel()

	var up *api.UserWithPassword
	if err := ctx.Bind(&up); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userIn := user.User{
		FirstName: up.FirstName,
		LastName:  up.LastName,
		Nickname:  up.Nickname,
		Email:     string(up.Email),
		Country:   up.Country,
	}

	pass := ""
	if up.Password != nil {
		pass = string(*up.Password)
	}

	u, err := h.userSvc.Create(c, userIn, pass)
	if err != nil {
		log.Err(err).
			Str("operation", "Create").
			Str(common.CorrelationID, common.GetCorrelationID(c)).
			Send()

		switch err {
		case user.ErrNewUserWithID, user.ErrInvalidUserInputData:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	return ctx.JSON(http.StatusCreated, toUserResponse(u))
}

func (h Handler) DeleteByID(ctx echo.Context, id uuid.UUID) error {
	c, cancel := h.contextWithTimeout(ctx)
	defer cancel()

	if err := h.userSvc.Delete(c, id); err != nil {
		log.Err(err).
			Str("operation", "DeleteByID").
			Str(common.CorrelationID, common.GetCorrelationID(c)).
			Stringer("ID", id).
			Send()

		switch err {
		case user.ErrNilUUIDNotAllowed:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (h Handler) GetByID(ctx echo.Context, id uuid.UUID) error {
	c, cancel := h.contextWithTimeout(ctx)
	defer cancel()

	u, err := h.userSvc.Get(c, id)
	if err != nil {
		log.Err(err).
			Str("operation", "GetByID").
			Str(common.CorrelationID, common.GetCorrelationID(c)).
			Stringer("ID", id).
			Send()

		switch err {
		case user.ErrUserNotFound:
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case user.ErrNilUUIDNotAllowed:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	return ctx.JSON(http.StatusOK, toUserResponse(u))
}

func (h Handler) UpdateByID(ctx echo.Context, id uuid.UUID) error {
	c, cancel := h.contextWithTimeout(ctx)
	defer cancel()

	var up *api.UpdateUserWithPassword
	if err := ctx.Bind(&up); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userIn := user.User{}
	if up.FirstName != nil {
		userIn.FirstName = *up.FirstName
	}
	if up.LastName != nil {
		userIn.LastName = *up.LastName
	}
	if up.Nickname != nil {
		userIn.Nickname = *up.Nickname
	}
	if up.Email != nil {
		userIn.Email = string(*up.Email)
	}
	if up.Country != nil {
		userIn.Country = *up.Country
	}
	password := ""
	if up.Password != nil {
		password = *up.Password
	}

	u, err := h.userSvc.Update(c, id, userIn, password)
	if err != nil {
		log.Err(err).
			Str("operation", "UpdateByID").
			Str(common.CorrelationID, common.GetCorrelationID(c)).
			Stringer("ID", id).
			Send()

		switch err {
		case user.ErrNilUUIDNotAllowed, user.ErrInvalidUserInputData:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	return ctx.JSON(http.StatusOK, toUserResponse(u))
}

func toUserResponse(u *user.User) api.UserResponse {
	return api.UserResponse{
		Id:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		Email:     types.Email(u.Email),
		Country:   u.Country,
		CreatedAt: &u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (h Handler) contextWithTimeout(ctx echo.Context) (context.Context, context.CancelFunc) {
	ec := ctx.Request().Context()
	c := context.WithValue(ec, common.CorrelationID, common.GetEchoCorrelationID(ctx))
	return context.WithTimeout(c, h.timeout)
}

func getListPagination(p api.ListParams) common.Pagination {
	pagination := common.Pagination{}
	if p.Page != nil {
		pagination.Page = *p.Page
	}

	if p.Pagesize != nil {
		pagination.PageSize = *p.Pagesize
	}

	return pagination
}

func getListFilter(p api.ListParams) user.User {
	filter := user.User{}
	if p.FirstName != nil {
		filter.FirstName = *p.FirstName
	}

	if p.LastName != nil {
		filter.LastName = *p.LastName
	}

	if p.Nickname != nil {
		filter.Nickname = *p.Nickname
	}

	if p.Email != nil {
		filter.Email = *p.Email
	}

	if p.Country != nil {
		filter.Country = *p.Country
	}

	return filter
}
