package tenant

import (
	"context"
	"github.com/google/uuid"
)

type contextKey string

const (
	// TenantIDKey is the key used to store the tenant ID in the context
	TenantIDKey contextKey = "tenant_id"
)

// NewContext creates a new context with the tenant ID
func NewContext(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// FromContext retrieves the tenant ID from the context
func FromContext(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	return tenantID, ok
}

// MustFromContext retrieves the tenant ID from the context or panics if not found
func MustFromContext(ctx context.Context) uuid.UUID {
	tenantID, ok := FromContext(ctx)
	if !ok {
		panic("tenant ID not found in context")
	}
	return tenantID
}
