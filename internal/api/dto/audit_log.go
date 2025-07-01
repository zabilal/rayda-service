package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/model"
)

// AuditLogResponse represents an audit log entry in the API response
// JSONMap is a type alias for JSON data stored as a map
// We use this instead of JSONB to avoid database-specific types in the DTO layer
type JSONMap map[string]interface{}

type AuditLogResponse struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	UserID       string    `json:"user_id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	RequestID    string    `json:"request_id,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	Metadata     JSONMap   `json:"metadata,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ToAuditLogResponse converts a model.AuditLog to an AuditLogResponse
func ToAuditLogResponse(log *model.AuditLog) *AuditLogResponse {
	if log == nil {
		return nil
	}

	// Convert model.JSONB to our JSONMap type
	metadata := make(JSONMap)
	for k, v := range log.Metadata {
		metadata[k] = v
	}

	return &AuditLogResponse{
		ID:           log.ID.String(),
		TenantID:     log.TenantID.String(),
		UserID:       log.UserID.String(),
		Action:       string(log.Action),
		ResourceType: string(log.ResourceType),
		ResourceID:   log.ResourceID,
		RequestID:    log.RequestID,
		IPAddress:    log.IPAddress,
		UserAgent:    log.UserAgent,
		Metadata:     metadata,
		CreatedAt:    log.CreatedAt,
	}
}

// ToAuditLogResponseList converts a slice of model.AuditLog to a slice of AuditLogResponse
func ToAuditLogResponseList(logs []*model.AuditLog) []*AuditLogResponse {
	result := make([]*AuditLogResponse, len(logs))
	for i, log := range logs {
		result[i] = ToAuditLogResponse(log)
	}
	return result
}

// AuditLogFilter represents the query parameters for filtering audit logs
type AuditLogFilter struct {
	Action       string     `form:"action"`
	ResourceType string     `form:"resource_type"`
	ResourceID   string     `form:"resource_id"`
	UserID       string     `form:"user_id"`
	UserUUID     *uuid.UUID `json:"-"` // Internal use only, not exposed in API
	StartTime    time.Time  `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime      time.Time  `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
	Page         int        `form:"page,default=1"`
	PageSize     int        `form:"page_size,default=20"`
}

// Validate validates the audit log filter
func (f *AuditLogFilter) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}

	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}

	return nil
}
