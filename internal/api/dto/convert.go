package dto

import (
	"github.com/rayda/rayda-service/internal/model"
)

// ToModelAuditLogFilter converts a dto.AuditLogFilter to a model.AuditLogFilter
func ToModelAuditLogFilter(dtoFilter AuditLogFilter) *model.AuditLogFilter {
	filter := &model.AuditLogFilter{
		Action:       model.ActionType(dtoFilter.Action),
		ResourceType: model.ResourceType(dtoFilter.ResourceType),
		ResourceID:   dtoFilter.ResourceID,
		UserID:       dtoFilter.UserID,
		StartTime:    dtoFilter.StartTime,
		EndTime:      dtoFilter.EndTime,
	}

	// Convert UserUUID from string to *string if set
	if dtoFilter.UserUUID != nil {
		userUUIDStr := dtoFilter.UserUUID.String()
		filter.UserUUID = &userUUIDStr
	}

	return filter
}
