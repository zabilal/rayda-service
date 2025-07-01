package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/rayda/rayda-service/internal/model"
)

// AuditLogRepository defines the interface for audit log data access
type AuditLogRepository interface {
	// Create saves a new audit log entry
	Create(ctx context.Context, log *model.AuditLog) error
	// FindByID retrieves an audit log by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error)
	// FindByTenantID retrieves audit logs for a specific tenant with filtering and pagination
	FindByTenantID(
		ctx context.Context,
		tenantID uuid.UUID,
		filter *model.AuditLogFilter,
		page, pageSize int,
	) ([]*model.AuditLog, int64, error)
	// FindByUserID retrieves audit logs for a specific user with pagination
	FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*model.AuditLog, int64, error)
	// FindByResource retrieves audit logs for a specific resource with pagination
	FindByResource(
		ctx context.Context,
		resourceType model.ResourceType,
		resourceID string,
		page, pageSize int,
	) ([]*model.AuditLog, int64, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{
		db: db,
	}
}

// Create creates a new audit log entry
func (r *auditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByID finds an audit log by its ID
func (r *auditLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	var log model.AuditLog
	err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, err
	}

	return &log, nil
}

// FindByTenantID returns audit logs for a specific tenant with filtering and pagination
func (r *auditLogRepository) FindByTenantID(
	ctx context.Context,
	tenantID uuid.UUID,
	filter *model.AuditLogFilter,
	page, pageSize int,
) ([]*model.AuditLog, int64, error) {
	var logs []*model.AuditLog
	var total int64

	offset := (page - 1) * pageSize

	// Start building the query
	db := r.db.WithContext(ctx).Model(&model.AuditLog{}).
		Where("tenant_id = ?", tenantID)

	// Apply filters if provided
	if filter != nil {
		if filter.Action != "" {
			db = db.Where("action = ?", filter.Action)
		}
		if filter.ResourceType != "" {
			db = db.Where("resource_type = ?", filter.ResourceType)
		}
		if filter.ResourceID != "" {
			db = db.Where("resource_id = ?", filter.ResourceID)
		}
		if filter.UserID != "" {
			userID, err := uuid.Parse(filter.UserID)
			if err == nil {
				db = db.Where("user_id = ?", userID)
			}
		}
		if !filter.StartTime.IsZero() {
			db = db.Where("created_at >= ?", filter.StartTime)
		}
		if !filter.EndTime.IsZero() {
			db = db.Where("created_at <= ?", filter.EndTime)
		}
	}

	// Get total count with filters applied
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get paginated results with filters applied
	if err := db.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find audit logs: %w", err)
	}

	return logs, total, nil
}

// FindByUserID finds audit logs by user ID with pagination
func (r *auditLogRepository) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*model.AuditLog, int64, error) {
	var logs []*model.AuditLog
	var count int64

	offset := (page - 1) * pageSize

	// Get total count
	if err := r.db.WithContext(ctx).Model(&model.AuditLog{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}

// FindByResource finds audit logs by resource type and ID with pagination
func (r *auditLogRepository) FindByResource(ctx context.Context, resourceType model.ResourceType, resourceID string, page, pageSize int) ([]*model.AuditLog, int64, error) {
	var logs []*model.AuditLog
	var count int64

	offset := (page - 1) * pageSize

	// Get total count
	if err := r.db.WithContext(ctx).Model(&model.AuditLog{}).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}
