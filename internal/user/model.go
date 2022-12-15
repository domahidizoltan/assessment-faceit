package user

import (
	"context"
	"errors"
	"net/mail"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrNilUUIDNotAllowed    = errors.New("nil UUID is not allowed")
	ErrNewUserWithID        = errors.New("new user can't have a predefined ID")
	ErrInvalidUserInputData = errors.New("input user data is invalid")
	ErrInvalidPagination    = errors.New("invalid pagination")
	ErrInvalidFilter        = errors.New("invalid filter")
)

type User struct {
	ID        uuid.UUID  `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Nickname  string     `json:"nickname"`
	Email     string     `json:"email"`
	Country   string     `json:"country"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func (u User) Validate() error {
	_, emailErr := mail.ParseAddress(u.Email)
	switch {
	case len(u.FirstName) < 2:
		return errors.New("first name is too short")
	case len(u.LastName) < 2:
		return errors.New("last name is too short")
	case len(u.Nickname) < 3:
		return errors.New("nickname is too short")
	case len(u.Country) != 2:
		return errors.New("country code must be 2 letters")
	case emailErr != nil:
		return errors.New("invalid email address")
	}
	return nil
}

func (u User) ValidateIfNotEmpty() error {
	if u.FirstName != "" && len(u.FirstName) < 2 {
		return errors.New("first_name must be at least 2 characters")
	}

	if u.LastName != "" && len(u.LastName) < 2 {
		return errors.New("last_name must be at least 2 characters")
	}

	if u.Nickname != "" && len(u.Nickname) < 2 {
		return errors.New("nickname must be at least 2 characters")
	}

	if u.Email != "" && len(u.Email) < 2 {
		return errors.New("email must be at least 2 characters")
	}

	if u.Country != "" && len(u.Country) != 2 {
		return errors.New("country must have exactly 2 characters")
	}

	return nil
}

type UserEventType string

var (
	UserEventTypeCreated         UserEventType = "USER_CREATED"
	UserEventTypeUpdated         UserEventType = "USER_UPDATED"
	UserEventTypePasswordChanged UserEventType = "USER_PASSWORD_CHANGED"
	UserEventTypeDeleted         UserEventType = "USER_DELETED"
)

type UserEvent struct {
	context     *context.Context `json:"-"`
	Type        UserEventType    `json:"type"`
	UserID      uuid.UUID        `json:"user_id"`
	UserChanges *User            `json:"user_changes,omitempty"`
	Time        time.Time        `json:"time"`
}
