package model

import (
	"github.com/google/uuid"
)

// ActionType represents the type of action performed on a resource
type ActionType string

const (
	// ActionCreate represents a create operation
	ActionCreate ActionType = "CREATE"
	// ActionUpdate represents an update operation
	ActionUpdate ActionType = "UPDATE"
	// ActionDelete represents a delete operation
	ActionDelete ActionType = "DELETE"
	// ActionLogin represents a user login
	ActionLogin ActionType = "LOGIN"
	// ActionLogout represents a user logout
	ActionLogout ActionType = "LOGOUT"
)

// ResourceType represents the type of resource being acted upon
type ResourceType string

const (
	// ResourceUser represents a user resource
	ResourceUser ResourceType = "USER"
	// ResourceOrganization represents an organization resource
	ResourceOrganization ResourceType = "ORGANIZATION"
	// ResourceAPIKey represents an API key resource
	ResourceAPIKey ResourceType = "API_KEY"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	Base
	TenantID     uuid.UUID    `json:"tenant_id" gorm:"type:uuid;not null;index"`
	UserID       uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index"`
	Action       ActionType   `json:"action" gorm:"type:varchar(20);not null"`
	ResourceType ResourceType `json:"resource_type" gorm:"type:varchar(50);not null"`
	ResourceID   string       `json:"resource_id" gorm:"type:varchar(36);not null"`
	RequestID    string       `json:"request_id" gorm:"type:varchar(100);index"`
	IPAddress    string       `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent    string       `json:"user_agent" gorm:"type:text"`
	Metadata     JSONB        `json:"metadata,omitempty" gorm:"type:jsonb"`
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(
	tenantID uuid.UUID,
	userID uuid.UUID,
	action ActionType,
	resourceType ResourceType,
	resourceID string,
	requestID string,
	ipAddress string,
	userAgent string,
	metadata map[string]interface{},
) *AuditLog {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &AuditLog{
		TenantID:     tenantID,
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		RequestID:    requestID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Metadata:     metadata,
	}
}

// WithMetadata adds metadata to the audit log entry
func (a *AuditLog) WithMetadata(key string, value interface{}) *AuditLog {
	if a.Metadata == nil {
		a.Metadata = make(JSONB)
	}
	a.Metadata[key] = value
	return a
}

// TableName specifies the table name for the AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}
