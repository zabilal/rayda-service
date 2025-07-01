// Package handlers contains HTTP request handlers for the API endpoints.
// This file implements the audit log API endpoints for listing and retrieving audit logs.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/api/dto"
	"github.com/rayda/rayda-service/internal/repository"
)

// AuditLogHandler handles audit log related HTTP requests.
// It provides endpoints for listing and retrieving audit logs with various filtering options.
type AuditLogHandler struct {
	auditLogRepo repository.AuditLogRepository // Repository for audit log data access
}

// NewAuditLogHandler creates a new AuditLogHandler
// NewAuditLogHandler creates a new AuditLogHandler with the given repository.
//
// Parameters:
//   - auditLogRepo: The repository used for audit log data access
//
// Returns:
//   - *AuditLogHandler: A new instance of AuditLogHandler
func NewAuditLogHandler(auditLogRepo repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{
		auditLogRepo: auditLogRepo,
	}
}

// RegisterRoutes registers the audit log routes with the provided router group.
// All routes are protected by the provided auth middleware.
//
// Routes:
//   - GET /audit-logs - List audit logs with optional filtering
//   - GET /audit-logs/:id - Get a specific audit log by ID
//
// Parameters:
//   - router: The router group to register routes with
//   - authMiddleware: The authentication middleware to protect the routes
func (h *AuditLogHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	auditGroup := router.Group("/audit-logs")
	auditGroup.Use(authMiddleware)
	{
		auditGroup.GET("", h.ListAuditLogs)
		auditGroup.GET("/:id", h.GetAuditLog)
	}
}

// ListAuditLogs handles GET /audit-logs
// It returns a paginated list of audit logs with optional filtering.
//
// The handler performs the following steps:
// 1. Extracts tenant ID from the request context
// 2. Parses and validates query parameters
// 3. Fetches logs from the repository
// 4. Applies additional filters in memory (temporary implementation)
// 5. Returns paginated results
//
// Note: The current implementation applies some filters in memory for simplicity.
// In a production environment with large datasets, these filters should be moved
// to the database level for better performance.
//
// Query Parameters:
//   - action: Filter by action type (e.g., "CREATE", "UPDATE", "DELETE")
//   - resource_type: Filter by resource type (e.g., "USER", "ORGANIZATION")
//   - resource_id: Filter by specific resource ID
//   - user_id: Filter by user ID
//   - start_time: Filter logs after this timestamp (ISO 8601)
//   - end_time: Filter logs before this timestamp (ISO 8601)
//   - page: Page number (default: 1)
//   - page_size: Number of items per page (default: 20, max: 100)
//
// Responses:
//   - 200: Successfully retrieved audit logs
//   - 400: Invalid request parameters
//   - 401: Unauthorized
//   - 500: Internal server error
//
// @Summary List audit logs
// @Description Get a paginated list of audit logs with optional filtering
// @Tags audit-logs
// @Accept json
// @Produce json
// @Security Bearer
// @Param action query string false "Action type (e.g., CREATE, UPDATE, DELETE)"
// @Param resource_type query string false "Resource type (e.g., USER, ORGANIZATION)"
// @Param resource_id query string false "Resource ID"
// @Param user_id query string false "User ID"
// @Param start_time query string false "Start time (ISO 8601)"
// @Param end_time query string false "End time (ISO 8601)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} dto.PaginatedResponse{data=[]dto.AuditLogResponse} "List of audit logs"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /audit-logs [get]
func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	// Get tenant ID from context
	tenantID, err := getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID in context"})
		return
	}

	// Parse query parameters
	var filter dto.AuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	// Convert DTO filter to model filter
	modelFilter := dto.ToModelAuditLogFilter(filter)

	// Use the repository's filtering capabilities
	logs, total, err := h.auditLogRepo.FindByTenantID(
		c.Request.Context(),
		tenantID,
		modelFilter,
		filter.Page,
		filter.PageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit logs: " + err.Error()})
		return
	}

	// No need for manual pagination as it's handled by the repository

	// Convert to DTOs
	response := make([]*dto.AuditLogResponse, len(logs))
	for i, log := range logs {
		response[i] = dto.ToAuditLogResponse(log)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       response,
		"pagination": getPagination(filter.Page, filter.PageSize, int(total)),
	})
}

// GetAuditLog returns a specific audit log by ID
// @Summary Get an audit log by ID
// @Description Get detailed information about a specific audit log entry
// @Tags audit-logs
// @Accept  json
// @Produce  json
// @Security Bearer
// @Param   id   path      string  true  "Audit Log ID"
// @Success 200 {object} dto.AuditLogResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /audit-logs/{id} [get]
func (h *AuditLogHandler) GetAuditLog(c *gin.Context) {
	// Get tenant ID from context
	tenantID, err := getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse log ID from URL
	logID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log ID"})
		return
	}

	// Get the audit log
	log, err := h.auditLogRepo.FindByID(c.Request.Context(), logID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audit log not found"})
		return
	}

	// Verify the log belongs to the current tenant
	if log.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, dto.ToAuditLogResponse(log))
}

// getTenantIDFromContext extracts the tenant ID from the request context
func getTenantIDFromContext(c *gin.Context) (uuid.UUID, error) {
	// TODO: Implement proper tenant ID extraction from JWT claims
	// For now, return a default tenant ID
	return uuid.Parse("00000000-0000-0000-0000-000000000001")
}

// getPagination generates pagination metadata
func getPagination(page, pageSize, total int) map[string]interface{} {
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return map[string]interface{}{
		"page":        page,
		"page_size":   pageSize,
		"total_items": total,
		"total_pages": totalPages,
		"has_prev":    page > 1,
		"has_next":    page < totalPages,
	}
}
