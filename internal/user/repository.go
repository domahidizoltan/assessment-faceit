package user

import (
	"context"
	"errors"
	"faceit/internal/common"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new DB connection
func NewRepository() (*gormRepository, error) {
	pgHost := common.GetEnv("PG_HOST", "localhost")
	pgPort := common.GetEnv("PG_PORT", "5432")
	pgUser := common.GetEnv("PG_USER", "admin")
	pgPass := common.GetEnv("PG_PASSWORD", "pass")
	pgDb := common.GetEnv("PG_DATABASE", "users")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", pgHost, pgPort, pgUser, pgPass, pgDb)
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, err
	}

	return &gormRepository{db: db}, nil
}

// GetDB returns the creates DB connection
func (r gormRepository) GetDB() *gorm.DB {
	return r.db
}

func (r gormRepository) findByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u *User
	if err := r.db.Take(&u, id).Error; err != nil {
		return nil, handleNotFoundError(err)
	}

	return u, nil
}

func (r gormRepository) list(ctx context.Context, pagination common.Pagination, filter *User) ([]User, error) {
	query := r.db

	if filter != nil {
		if filter.FirstName != "" {
			query = query.Where("first_name ILIKE ?", filter.FirstName+"%")
		}
		if filter.LastName != "" {
			query = query.Where("last_name ILIKE ?", filter.LastName+"%")
		}
		if filter.Nickname != "" {
			query = query.Where("nickname ILIKE ?", filter.Nickname+"%")
		}
		if filter.Email != "" {
			query = query.Where("email ILIKE ?", filter.Email+"%")
		}
		if filter.Country != "" {
			query = query.Where("country", strings.ToUpper(filter.Country))
		}
	}

	query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())

	var users []User
	if err := query.Order("created_at desc").Order("email asc").Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (r gormRepository) create(ctx context.Context, user User, password string) (*User, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return nil
		}
		return updatePassword(tx, ctx, user.ID, password)
	})

	return &user, err
}

func updatePassword(tx *gorm.DB, ctx context.Context, id uuid.UUID, password string) error {
	err := tx.Model(User{}).
		Where("id = ?", id).
		Update("password", password).
		Error
	return handleNotFoundError(err)
}

func (r gormRepository) updatePassword(ctx context.Context, id uuid.UUID, password string) error {
	return updatePassword(r.db, ctx, id, password)
}

func (r gormRepository) update(ctx context.Context, id uuid.UUID, user User) (*User, error) {
	var updatedUser User
	err := r.db.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("id = ?", id).
		Updates(user).
		Error
	return &updatedUser, handleNotFoundError(err)
}

func (r gormRepository) deleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.Delete(&User{}, id).Error
}

func handleNotFoundError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}
	return err
}
