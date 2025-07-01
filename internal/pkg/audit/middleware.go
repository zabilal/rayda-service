package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/model"
	"github.com/rayda/rayda-service/internal/repository"
)

// MiddlewareConfig holds configuration for the audit middleware
type MiddlewareConfig struct {
	SensitiveHeaders []string
	SensitiveFields  []string
}

// DefaultConfig returns the default configuration for the audit middleware
func DefaultConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		SensitiveHeaders: []string{"authorization", "cookie", "set-cookie"},
		SensitiveFields:  []string{"password", "token", "api_key"},
	}
}

// Middleware creates a new audit middleware
func Middleware(auditRepo repository.AuditLogRepository, config *MiddlewareConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultConfig()
	}

	return func(c *gin.Context) {
		// Skip logging for certain paths (e.g., health checks)
		if shouldSkipLogging(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Read the request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore the request body for further processing
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create a custom response writer to capture the response
		blw := &bodyLogWriter{body: bytes.NewBuffer([]byte{}), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process the request
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Skip logging for non-modifying methods
		if !isModifyingMethod(c.Request.Method) {
			return
		}

		// Get the current user and tenant from the context
		userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000") // Default to system user
		tenantID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000") // Default to system tenant

		// TODO: Uncomment and implement when auth is in place
		// if claims, exists := c.Get("user"); exists {
		// 	if userClaims, ok := claims.(*auth.Claims); ok {
		// 		userID = userClaims.UserID
		// 		tenantID = userClaims.TenantID
		// 	}
		// }

		// Determine the action type based on the HTTP method
		action := getActionType(c.Request.Method)

		// Get the resource type and ID from the URL
		resourceType, resourceID := extractResourceInfo(c)

		// Create the audit log entry
		auditLog := &model.AuditLog{
			TenantID:     tenantID,
			UserID:       userID,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			RequestID:    c.GetHeader("X-Request-ID"),
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Metadata:     make(map[string]interface{}),
		}

		// Add request and response data to metadata
		auditLog.Metadata["method"] = c.Request.Method
		auditLog.Metadata["path"] = c.Request.URL.Path
		auditLog.Metadata["query"] = c.Request.URL.RawQuery
		auditLog.Metadata["status_code"] = c.Writer.Status()
		auditLog.Metadata["duration_ms"] = duration.Milliseconds()

		// Add request headers (excluding sensitive ones)
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if !contains(config.SensitiveHeaders, strings.ToLower(k)) {
				headers[k] = strings.Join(v, ", ")
			}
		}
		auditLog.Metadata["request_headers"] = headers

		// Add request body if present
		if len(requestBody) > 0 {
			var bodyMap map[string]interface{}
			if err := json.Unmarshal(requestBody, &bodyMap); err == nil {
				// Remove sensitive fields from the request body
				for _, field := range config.SensitiveFields {
					delete(bodyMap, field)
				}
				auditLog.Metadata["request_body"] = bodyMap
			}
		}

		// Add response data if needed
		if c.Writer.Status() >= 400 {
			var response interface{}
			if err := json.Unmarshal(blw.body.Bytes(), &response); err == nil {
				auditLog.Metadata["error"] = response
			}
		}

		// Save the audit log entry in a goroutine to avoid blocking the response
		go func(log *model.AuditLog) {
			_ = auditRepo.Create(context.Background(), log)
		}(auditLog)
	}
}

// bodyLogWriter is a custom response writer that captures the response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// shouldSkipLogging checks if the request path should be skipped from logging
func shouldSkipLogging(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
	}

	for _, p := range skipPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// isModifyingMethod checks if the HTTP method modifies data
func isModifyingMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || 
	       method == http.MethodPatch || method == http.MethodDelete
}

// getActionType maps HTTP methods to audit log action types
func getActionType(method string) model.ActionType {
	switch method {
	case http.MethodPost:
		return model.ActionCreate
	case http.MethodPut, http.MethodPatch:
		return model.ActionUpdate
	case http.MethodDelete:
		return model.ActionDelete
	default:
		return model.ActionType(strings.ToUpper(method))
	}
}

// extractResourceInfo extracts the resource type and ID from the request path
func extractResourceInfo(c *gin.Context) (model.ResourceType, string) {
	path := c.Request.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) >= 2 {
		switch parts[0] {
		case "users":
			if len(parts) >= 2 && parts[1] != "" {
				return model.ResourceUser, parts[1]
			}
		case "organizations":
			if len(parts) >= 2 && parts[1] != "" {
				return model.ResourceOrganization, parts[1]
			}
		}
	}

	return model.ResourceType("UNKNOWN"), ""
}

// contains checks if a string is present in a slice
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
