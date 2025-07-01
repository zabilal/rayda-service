package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/model"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	TenantAwareRepository[model.User]
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*model.User, error)
}

type userRepository struct {
	TenantAwareRepository[model.User]
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		TenantAwareRepository: NewTenantAwareRepository[model.User](db),
		db: db,
	}
}

// FindByEmail finds a user by email (across all tenants)
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// FindByEmailAndTenant finds a user by email within a specific tenant
func (r *userRepository) FindByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).
		Where("email = ? AND tenant_id = ?", email, tenantID).
		First(&user).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
