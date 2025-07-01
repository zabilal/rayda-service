package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/model"
	"github.com/rayda/rayda-service/internal/pkg/tenant"
)

const (
	// AuthorizationHeader is the header key for the authorization token
	AuthorizationHeader = "Authorization"
	// BearerTokenPrefix is the prefix for the authorization header
	BearerTokenPrefix = "Bearer "
)

// Middleware handles JWT authentication
func Middleware(authSvc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		// Extract the token from the header
		if !strings.HasPrefix(authHeader, BearerTokenPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerTokenPrefix)

		// Parse and validate the token
		claims, err := authSvc.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Set the user information in the context
		ctx := context.WithValue(c.Request.Context(), userKey, claims)
		ctx = tenant.NewContext(ctx, claims.TenantID)

		// Update the request with the new context
		c.Request = c.Request.WithContext(ctx)

		// Continue to the next handler
		c.Next()
	}
}

// RoleRequired creates a middleware that requires a specific role
func RoleRequired(authSvc *Service, requiredRole model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the claims from the context
		claims, err := GetClaimsFromContext(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Check if the user has the required role
		if !authSvc.HasPermission(requiredRole, claims.Role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}

// contextKey is a type for context keys
type contextKey string

const (
	// userKey is the key for storing user claims in the context
	userKey contextKey = "user_claims"
)

// GetClaimsFromContext retrieves the claims from the context
func GetClaimsFromContext(ctx context.Context) (*Claims, error) {
	claims, ok := ctx.Value(userKey).(*Claims)
	if !ok {
		return nil, errors.New("no claims found in context")
	}
	return claims, nil
}

// GetCurrentUserID returns the current user ID from the context
func GetCurrentUserID(ctx context.Context) (uuid.UUID, error) {
	claims, err := GetClaimsFromContext(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

// GetCurrentTenantID returns the current tenant ID from the context
func GetCurrentTenantID(ctx context.Context) (uuid.UUID, error) {
	claims, err := GetClaimsFromContext(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.TenantID, nil
}
