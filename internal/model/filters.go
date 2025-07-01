package model

import "time"

// AuditLogFilter represents the filter criteria for querying audit logs
type AuditLogFilter struct {
	// Action filters logs by the type of action performed
	Action ActionType `json:"action,omitempty"`
	// ResourceType filters logs by the type of resource
	ResourceType ResourceType `json:"resource_type,omitempty"`
	// ResourceID filters logs by the specific resource ID
	ResourceID string `json:"resource_id,omitempty"`
	// UserID filters logs by the user who performed the action (as a string)
	UserID string `json:"user_id,omitempty"`
	// UserUUID filters logs by the user who performed the action (as a UUID)
	UserUUID *string `json:"-"`
	// StartTime filters logs that occurred after this time
	StartTime time.Time `json:"start_time,omitempty"`
	// EndTime filters logs that occurred before this time
	EndTime time.Time `json:"end_time,omitempty"`
}

// HasTimeRange returns true if either StartTime or EndTime is set
func (f *AuditLogFilter) HasTimeRange() bool {
	return !f.StartTime.IsZero() || !f.EndTime.IsZero()
}

// HasFilters returns true if any filter criteria are set
func (f *AuditLogFilter) HasFilters() bool {
	return f.Action != "" ||
		f.ResourceType != "" ||
		f.ResourceID != "" ||
		f.UserID != "" ||
		f.UserUUID != nil ||
		f.HasTimeRange()
}
