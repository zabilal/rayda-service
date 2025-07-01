package repository

import (
	"context"
	"errors"

	"github.com/rayda/rayda-service/internal/model"
	"gorm.io/gorm"
)

// OrganizationRepository defines the interface for organization data operations
type OrganizationRepository interface {
	TenantAwareRepository[model.Organization]
	FindByDomain(ctx context.Context, domain string) (*model.Organization, error)
	FindAll(ctx context.Context) ([]model.Organization, error)
}

type organizationRepository struct {
	TenantAwareRepository[model.Organization]
	db *gorm.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{
		TenantAwareRepository: NewTenantAwareRepository[model.Organization](db),
		db: db,
	}
}

// FindByDomain finds an organization by its domain
func (r *organizationRepository) FindByDomain(ctx context.Context, domain string) (*model.Organization, error) {
	var org model.Organization

	err := r.db.WithContext(ctx).
		Where("domain = ?", domain).
		First(&org).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &org, nil
}

// FindAll returns all organizations
func (r *organizationRepository) FindAll(ctx context.Context) ([]model.Organization, error) {
	var orgs []model.Organization

	err := r.db.WithContext(ctx).Find(&orgs).Error
	if err != nil {
		return nil, err
	}

	return orgs, nil
}
