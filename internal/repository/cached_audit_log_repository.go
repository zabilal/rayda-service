package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/model"
	"github.com/rayda/rayda-service/internal/pkg/cache"
)

// CachedAuditLogRepository is a wrapper around AuditLogRepository that adds caching
// to improve performance for frequently accessed audit logs.
type CachedAuditLogRepository struct {
	repo  AuditLogRepository
	cache cache.Cache
	ttl   time.Duration
}

// NewCachedAuditLogRepository creates a new cached audit log repository
func NewCachedAuditLogRepository(repo AuditLogRepository, cache cache.Cache, ttl time.Duration) *CachedAuditLogRepository {
	return &CachedAuditLogRepository{
		repo:  repo,
		cache: cache,
		ttl:   ttl,
	}
}

// Create saves a new audit log entry and invalidates relevant cache entries
func (r *CachedAuditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	if err := r.repo.Create(ctx, log); err != nil {
		return err
	}

	// Invalidate relevant cache entries
	// We don't cache individual creates, but we need to invalidate any list caches
	// that might be affected by this new log entry
	tenantKey := r.tenantListCacheKey(log.TenantID, nil)
	_ = r.cache.Delete(ctx, tenantKey)

	return nil
}

// FindByID retrieves an audit log by its ID with caching
func (r *CachedAuditLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	cacheKey := r.logCacheKey(id)

	// Try to get from cache first
	var cachedLog model.AuditLog
	found, err := r.cache.Get(ctx, cacheKey, &cachedLog)
	if err != nil {
		// Log the error but continue with the database lookup
		// In a production environment, you might want to use a proper logger here
		_ = err
	}

	if found {
		return &cachedLog, nil
	}

	// Not found in cache, get from database
	log, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	_ = r.cache.Set(ctx, cacheKey, log, r.ttl)

	return log, nil
}

// FindByTenantID retrieves audit logs for a specific tenant with filtering and pagination
func (r *CachedAuditLogRepository) FindByTenantID(
	ctx context.Context,
	tenantID uuid.UUID,
	filter *model.AuditLogFilter,
	page, pageSize int,
) ([]*model.AuditLog, int64, error) {
	// Generate a cache key based on the query parameters
	cacheKey := r.tenantListCacheKey(tenantID, filter)

	// For paginated results, we'll use a different caching strategy
	// We'll cache the first few pages with a shorter TTL
	if page == 1 && pageSize <= 100 {
		// Try to get from cache first
		var cachedResult struct {
			Logs  []*model.AuditLog
			Total int64
		}

		found, err := r.cache.Get(ctx, cacheKey, &cachedResult)
		if err == nil && found {
			// Return cached result
			return cachedResult.Logs, cachedResult.Total, nil
		}
	}

	// Not found in cache or not cacheable, get from database
	logs, total, err := r.repo.FindByTenantID(ctx, tenantID, filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// Cache the first page of results with a shorter TTL
	if page == 1 && pageSize <= 100 {
		result := struct {
			Logs  []*model.AuditLog
			Total int64
		}{
			Logs:  logs,
			Total: total,
		}

		// Use a shorter TTL for list results
		ttl := r.ttl / 2
		if ttl < time.Minute {
			ttl = time.Minute
		}

		_ = r.cache.Set(ctx, cacheKey, result, ttl)
	}

	return logs, total, nil
}

// FindByUserID retrieves audit logs for a specific user with pagination
func (r *CachedAuditLogRepository) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
	page, pageSize int,
) ([]*model.AuditLog, int64, error) {
	// For simplicity, we'll bypass caching for user-specific queries
	// as they're less common and more likely to be unique
	return r.repo.FindByUserID(ctx, userID, page, pageSize)
}

// FindByResource retrieves audit logs for a specific resource with pagination
func (r *CachedAuditLogRepository) FindByResource(
	ctx context.Context,
	resourceType model.ResourceType,
	resourceID string,
	page, pageSize int,
) ([]*model.AuditLog, int64, error) {
	// For simplicity, we'll bypass caching for resource-specific queries
	// as they're less common and more likely to be unique
	return r.repo.FindByResource(ctx, resourceType, resourceID, page, pageSize)
}

// logCacheKey generates a cache key for an individual audit log
func (r *CachedAuditLogRepository) logCacheKey(id uuid.UUID) string {
	return fmt.Sprintf("audit:log:%s", id.String())
}

// tenantListCacheKey generates a cache key for a tenant's audit log list
func (r *CachedAuditLogRepository) tenantListCacheKey(tenantID uuid.UUID, filter *model.AuditLogFilter) string {
	key := fmt.Sprintf("audit:tenant:%s:list", tenantID.String())

	if filter == nil {
		return key
	}

	// Include filter parameters in the cache key
	if filter.Action != "" {
		key = fmt.Sprintf("%s:action:%s", key, filter.Action)
	}
	if filter.ResourceType != "" {
		key = fmt.Sprintf("%s:resource_type:%s", key, filter.ResourceType)
	}
	if filter.ResourceID != "" {
		key = fmt.Sprintf("%s:resource_id:%s", key, filter.ResourceID)
	}
	if filter.UserID != "" {
		key = fmt.Sprintf("%s:user_id:%s", key, filter.UserID)
	}
	if !filter.StartTime.IsZero() {
		key = fmt.Sprintf("%s:start:%d", key, filter.StartTime.Unix())
	}
	if !filter.EndTime.IsZero() {
		key = fmt.Sprintf("%s:end:%d", key, filter.EndTime.Unix())
	}

	return key
}
