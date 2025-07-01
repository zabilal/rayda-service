package dto

import (
	"time"

	"github.com/rayda/rayda-service/internal/model"
)

// CreateOrganizationRequest represents the request payload for creating an organization
type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name,omitempty"`
	Domain      string `json:"domain" binding:"required,fqdn"`
}

// UpdateOrganizationRequest represents the request payload for updating an organization
type UpdateOrganizationRequest struct {
	Name        *string             `json:"name,omitempty"`
	DisplayName *string             `json:"display_name,omitempty"`
	Domain      *string             `json:"domain,omitempty" binding:"omitempty,fqdn"`
	Status      *model.OrganizationStatus `json:"status,omitempty"`
}

// OrganizationResponse represents the organization response
type OrganizationResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	DisplayName string                `json:"display_name"`
	Domain      string                `json:"domain"`
	Status      model.OrganizationStatus `json:"status"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
}

// ToOrganizationResponse converts a model.Organization to an OrganizationResponse
func ToOrganizationResponse(org *model.Organization) *OrganizationResponse {
	if org == nil {
		return nil
	}

	resp := &OrganizationResponse{
		ID:          org.ID.String(),
		Name:        org.Name,
		DisplayName: org.DisplayName,
		Domain:      org.Domain,
		Status:      org.Status,
		CreatedAt:   org.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   org.UpdatedAt.Format(time.RFC3339),
	}

	return resp
}

// ToOrganizationResponseList converts a slice of model.Organization to a slice of OrganizationResponse
func ToOrganizationResponseList(orgs []*model.Organization) []*OrganizationResponse {
	result := make([]*OrganizationResponse, len(orgs))
	for i, org := range orgs {
		result[i] = ToOrganizationResponse(org)
	}
	return result
}
