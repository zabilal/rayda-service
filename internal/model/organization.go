package model

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationStatus string

const (
	OrganizationStatusActive   OrganizationStatus = "active"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
	OrganizationStatusTrial    OrganizationStatus = "trial"
)

type Organization struct {
	TenantModel
	Name        string             `json:"name" gorm:"not null"`
	DisplayName string             `json:"display_name"`
	Domain      string             `json:"domain" gorm:"uniqueIndex"`
	Status      OrganizationStatus `json:"status" gorm:"type:varchar(20);not null;default:'trial'"`
	Metadata    JSONB              `json:"metadata,omitempty" gorm:"type:jsonb"`
}

// JSONB is a custom type for handling JSONB data in PostgreSQL
type JSONB map[string]interface{}

// NewOrganization creates a new organization with the given name and domain
func NewOrganization(name, displayName, domain string) *Organization {
	if displayName == "" {
		displayName = name
	}

	return &Organization{
		TenantModel: TenantModel{
			Base: Base{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			// TenantID is set when the organization is created in the repository
		},
		Name:        name,
		DisplayName: displayName,
		Domain:      domain,
		Status:      OrganizationStatusTrial,
		Metadata:    JSONB{},
	}
}

// IsActive returns true if the organization is active
func (o *Organization) IsActive() bool {
	return o.Status == OrganizationStatusActive || o.Status == OrganizationStatusTrial
}

// SetStatus updates the organization status
func (o *Organization) SetStatus(status OrganizationStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}

// UpdateMetadata updates the organization metadata
func (o *Organization) UpdateMetadata(updates map[string]interface{}) {
	if o.Metadata == nil {
		o.Metadata = JSONB{}
	}

	for key, value := range updates {
		o.Metadata[key] = value
	}
	o.UpdatedAt = time.Now()
}
