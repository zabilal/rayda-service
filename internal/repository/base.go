package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a record is not found
var ErrNotFound = errors.New("record not found")

// BaseRepository defines the common operations for all repositories
type BaseRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*T, error)
}

// TenantAwareRepository extends BaseRepository with tenant-specific operations
type TenantAwareRepository[T any] interface {
	BaseRepository[T]
	FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]T, error)
	DeleteByTenantID(ctx context.Context, tenantID uuid.UUID) error
}

// baseRepository is the base implementation of BaseRepository
type baseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T any](db *gorm.DB) BaseRepository[T] {
	return &baseRepository[T]{db: db}
}

// Create creates a new entity
func (r *baseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// Update updates an existing entity
func (r *baseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes an entity by ID
func (r *baseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	result := r.db.WithContext(ctx).Delete(&entity, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// FindByID finds an entity by ID
func (r *baseRepository[T]) FindByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

// tenantAwareRepository is the base implementation of TenantAwareRepository
type tenantAwareRepository[T any] struct {
	baseRepository[T]
}

// NewTenantAwareRepository creates a new tenant-aware repository
func NewTenantAwareRepository[T any](db *gorm.DB) TenantAwareRepository[T] {
	return &tenantAwareRepository[T]{
		baseRepository: baseRepository[T]{db: db},
	}
}

// FindByTenantID finds all entities for a specific tenant
func (r *tenantAwareRepository[T]) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

// DeleteByTenantID deletes all entities for a specific tenant
func (r *tenantAwareRepository[T]) DeleteByTenantID(ctx context.Context, tenantID uuid.UUID) error {
	var entity T
	result := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&entity)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// WithTenantScope returns a scoped database query with tenant filter
func WithTenantScope(db *gorm.DB, tenantID uuid.UUID) *gorm.DB {
	return db.Where("tenant_id = ?", tenantID)
}
