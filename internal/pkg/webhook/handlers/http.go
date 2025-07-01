package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rayda/rayda-service/internal/pkg/webhook/models"
	"github.com/rayda/rayda-service/internal/pkg/webhook/services"
)

// WebhookHandler handles HTTP requests for webhook operations
type WebhookHandler struct {
	service services.WebhookService
}

// NewWebhookHandler creates a new WebhookHandler
func NewWebhookHandler(service services.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		service: service,
	}
}

// CreateSubscriptionRequest represents the request body for creating a subscription
type CreateSubscriptionRequest struct {
	Name        string            `json:"name" binding:"required"`
	CallbackURL string            `json:"callback_url" binding:"required,url"`
	Events      []models.EventType `json:"events" binding:"required,min=1"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CreateSubscription creates a new webhook subscription
func (h *WebhookHandler) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := h.service.Subscribe(c.Request.Context(), req.Name, req.CallbackURL, req.Events, req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// GetSubscription retrieves a subscription by ID
func (h *WebhookHandler) GetSubscription(c *gin.Context) {
	subID := c.Param("id")
	if subID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription ID is required"})
		return
	}

	sub, err := h.service.GetSubscription(c.Request.Context(), subID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// ListSubscriptions lists all subscriptions, optionally filtered by event type
func (h *WebhookHandler) ListSubscriptions(c *gin.Context) {
	eventType := c.Query("event_type")
	
	subs, err := h.service.ListSubscriptions(c.Request.Context(), models.EventType(eventType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": subs})
}

// DeleteSubscription deletes a subscription
func (h *WebhookHandler) DeleteSubscription(c *gin.Context) {
	subID := c.Param("id")
	if subID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription ID is required"})
		return
	}

	if err := h.service.Unsubscribe(c.Request.Context(), subID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateSubscriptionRequest represents the request body for updating a subscription
type UpdateSubscriptionRequest struct {
	Name        *string               `json:"name,omitempty"`
	CallbackURL *string               `json:"callback_url,omitempty"`
	Status      *string               `json:"status,omitempty"`
	Events      *[]models.EventType   `json:"events,omitempty"`
	Metadata    *map[string]string    `json:"metadata,omitempty"`
}

// UpdateSubscription updates a subscription
func (h *WebhookHandler) UpdateSubscription(c *gin.Context) {
	subID := c.Param("id")
	if subID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription ID is required"})
		return
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request to updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.CallbackURL != nil {
		updates["callback_url"] = *req.CallbackURL
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Events != nil {
		updates["events"] = *req.Events
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	sub, err := h.service.UpdateSubscription(c.Request.Context(), subID, updates)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// RegisterRoutes registers the webhook routes
func (h *WebhookHandler) RegisterRoutes(router *gin.RouterGroup) {
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("", h.CreateSubscription)
		webhooks.GET("", h.ListSubscriptions)
		webhooks.GET("/:id", h.GetSubscription)
		webhooks.PUT("/:id", h.UpdateSubscription)
		webhooks.DELETE("/:id", h.DeleteSubscription)
	}
}
