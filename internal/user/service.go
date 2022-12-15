package user

import (
	"context"
	"crypto/sha256"
	"faceit/internal/common"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type (
	repository interface {
		findByID(ctx context.Context, id uuid.UUID) (*User, error)
		list(ctx context.Context, pagination common.Pagination, filter *User) ([]User, error)
		create(ctx context.Context, user User, password string) (*User, error)
		update(ctx context.Context, id uuid.UUID, user User) (*User, error)
		updatePassword(ctx context.Context, id uuid.UUID, password string) error
		deleteByID(ctx context.Context, id uuid.UUID) error
	}

	eventPublisher interface {
		publishCreated(ctx context.Context, userID uuid.UUID, userChanges *User) error
		publishDeleted(ctx context.Context, userID uuid.UUID) error
		publishUpdated(ctx context.Context, userID uuid.UUID, userChanges *User) error
		publishPasswordChanged(ctx context.Context, userID uuid.UUID) error
	}

	// Service manages the users
	Service struct {
		repository     repository
		eventPublisher eventPublisher
	}
)

// NewService creates a new Service with it's all required dependencies
func NewService() (*Service, error) {
	r, err := NewRepository()
	if err != nil {
		return nil, err
	}

	p, err := NewEventPublisher()
	if err != nil {
		return nil, err
	}

	return &Service{
		repository:     r,
		eventPublisher: p,
	}, nil
}

// Create validates and saves a new user.
// Password is handled separately and it will be encrypted.
// A UserEventTypeCreated event is published when it's done successfully.
func (s Service) Create(ctx context.Context, user User, password string) (*User, error) {
	if user.ID != uuid.Nil {
		return nil, ErrNewUserWithID
	}

	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidUserInputData, err.Error())
	}

	user.Country = strings.ToUpper(user.Country)
	newUser, err := s.repository.create(ctx, user, encryptPass(password))
	if err == nil {
		if err := s.eventPublisher.publishCreated(ctx, newUser.ID, newUser); err != nil {
			log.Err(err).
				Str(common.CorrelationID, common.GetCorrelationID(ctx)).
				Stringer("ID", newUser.ID).
				Msg("failed to publish create event")
		}
	}
	return newUser, err
}

// Get retrieves a single user.
func (s Service) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	if id == uuid.Nil {
		return nil, ErrNilUUIDNotAllowed
	}
	return s.repository.findByID(ctx, id)
}

// Update validates and saves changes on an existing user.
// Password is encrypted and saved separately when it is not empty.
// A UserEventTypeUpdated and UserEventTypePasswordChanged events are published when the updates are done successfully.
func (s Service) Update(ctx context.Context, id uuid.UUID, user User, password string) (*User, error) {
	if id == uuid.Nil {
		return nil, ErrNilUUIDNotAllowed
	}

	if err := user.ValidateIfNotEmpty(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidUserInputData, err.Error())
	}

	if user.Country != "" {
		user.Country = strings.ToUpper(user.Country)
	}

	var updatedUser *User = nil
	var err error
	emptyUser := User{}
	if user != emptyUser {
		updatedUser, err = s.repository.update(ctx, id, user)
		if err == nil {
			if err := s.eventPublisher.publishUpdated(ctx, id, updatedUser); err != nil {
				log.Err(err).
					Str(common.CorrelationID, common.GetCorrelationID(ctx)).
					Stringer("ID", id).
					Msg("failed to publish update event")
			}
		}

	}

	if password == "" {
		return updatedUser, err
	}

	pass := encryptPass(password)
	err = s.repository.updatePassword(ctx, id, pass)
	if err == nil {
		if err := s.eventPublisher.publishPasswordChanged(ctx, id); err != nil {
			log.Err(err).
				Str(common.CorrelationID, common.GetCorrelationID(ctx)).
				Stringer("ID", id).
				Msg("failed to publish update password event")
		}
	}

	return updatedUser, err
}

// Delete removes an existing user.
// A UserEventTypeDeleted event is published when it's done successfully.
func (s Service) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrNilUUIDNotAllowed
	}

	err := s.repository.deleteByID(ctx, id)
	if err == nil {
		if err := s.eventPublisher.publishDeleted(ctx, id); err != nil {
			log.Err(err).
				Str(common.CorrelationID, common.GetCorrelationID(ctx)).
				Stringer("ID", id).
				Msg("failed to publish delete event")
		}
	}
	return err
}

// List returns an ordered slice of users.
// The result is paged what could be parameterized and filtered.
func (s Service) List(ctx context.Context, pagination common.Pagination, filter *User) ([]User, error) {
	if err := pagination.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPagination, err.Error())
	}

	if err := filter.ValidateIfNotEmpty(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidFilter, err.Error())
	}

	return s.repository.list(ctx, pagination, filter)
}

func encryptPass(pass string) string {
	encrypted := sha256.Sum256([]byte(pass))
	return fmt.Sprintf("%x", encrypted)
}
