package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/api/dto"
	"github.com/rayda/rayda-service/internal/model"
	"github.com/rayda/rayda-service/internal/repository"
)

// OrganizationHandler handles organization-related requests
type OrganizationHandler struct {
	orgRepo repository.OrganizationRepository
}

// NewOrganizationHandler creates a new OrganizationHandler
func NewOrganizationHandler(orgRepo repository.OrganizationRepository) *OrganizationHandler {
	return &OrganizationHandler{
		orgRepo: orgRepo,
	}
}

// RegisterRoutes registers the organization routes with the provided router group
func (h *OrganizationHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	orgsGroup := router.Group("/organizations")
	orgsGroup.Use(authMiddleware)
	{
		orgsGroup.GET("", h.ListOrganizations)
		orgsGroup.POST("", h.CreateOrganization)
		orgsGroup.GET("/:id", h.GetOrganization)
		orgsGroup.PUT("/:id", h.UpdateOrganization)
		orgsGroup.DELETE("/:id", h.DeleteOrganization)
	}
}

// ListOrganizations returns a list of organizations for the current tenant
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	// In a real implementation, you would filter by the current user's tenant ID
	// For now, we'll return all organizations
	orgs, err := h.orgRepo.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organizations"})
		return
	}

	// Convert []model.Organization to []*model.Organization
	orgPtrs := make([]*model.Organization, len(orgs))
	for i := range orgs {
		orgPtrs[i] = &orgs[i]
	}

	c.JSON(http.StatusOK, dto.ToOrganizationResponseList(orgPtrs))
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	// Get current user's tenant ID (in a real app, you might have different permissions)
	// For now, we'll just create the organization

	var req dto.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if organization with this domain already exists
	existingOrg, _ := h.orgRepo.FindByDomain(c.Request.Context(), req.Domain)
	if existingOrg != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "organization with this domain already exists"})
		return
	}

	// Create new organization
	org := model.NewOrganization(
		req.Name,
		req.DisplayName,
		req.Domain,
	)

	// Save organization to database
	if err := h.orgRepo.Create(c.Request.Context(), org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
		return
	}

	c.JSON(http.StatusCreated, dto.ToOrganizationResponse(org))
}

// GetOrganization returns an organization by ID
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	// Parse organization ID from URL
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	// Get organization by ID
	org, err := h.orgRepo.FindByID(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ToOrganizationResponse(org))
}

// UpdateOrganization updates an organization
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	// Parse organization ID from URL
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	// Get organization by ID
	org, err := h.orgRepo.FindByID(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	var req dto.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update organization fields if provided
	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.DisplayName != nil {
		org.DisplayName = *req.DisplayName
	}
	if req.Domain != nil && *req.Domain != org.Domain {
		// Check if the new domain is already taken
		existingOrg, _ := h.orgRepo.FindByDomain(c.Request.Context(), *req.Domain)
		if existingOrg != nil && existingOrg.ID != org.ID {
			c.JSON(http.StatusConflict, gin.H{"error": "organization with this domain already exists"})
			return
		}
		org.Domain = *req.Domain
	}
	if req.Status != nil {
		org.Status = *req.Status
	}

	// Save updated organization
	if err := h.orgRepo.Update(c.Request.Context(), org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization"})
		return
	}

	c.JSON(http.StatusOK, dto.ToOrganizationResponse(org))
}

// DeleteOrganization deletes an organization
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	// Parse organization ID from URL
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	// Check if organization exists
	_, err = h.orgRepo.FindByID(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	// Delete organization
	if err := h.orgRepo.Delete(c.Request.Context(), orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization"})
		return
	}

	c.Status(http.StatusNoContent)
}
