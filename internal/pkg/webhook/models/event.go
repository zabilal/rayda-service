package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of webhook event
type EventType string

// Common webhook event types
const (
	EventTypeUserCreated       EventType = "user.created"
	EventTypeUserUpdated       EventType = "user.updated"
	EventTypeUserDeleted       EventType = "user.deleted"
	EventTypePaymentSucceeded EventType = "payment.succeeded"
	EventTypePaymentFailed    EventType = "payment.failed"
	EventTypePaymentRefunded  EventType = "payment.refunded"
)

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID            string          `json:"id"`
	EventType     EventType       `json:"event_type"`
	Source        string          `json:"source"` // e.g., "user_service", "payment_service"
	Data          json.RawMessage `json:"data"`
	Timestamp     time.Time       `json:"timestamp"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// NewWebhookEvent creates a new webhook event
func NewWebhookEvent(eventType EventType, source string, data interface{}) (*WebhookEvent, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &WebhookEvent{
		ID:        uuid.New().String(),
		EventType: eventType,
		Source:    source,
		Data:      payload,
		Timestamp: time.Now().UTC(),
	}, nil
}

// ParseData parses the event data into the provided interface
func (e *WebhookEvent) ParseData(target interface{}) error {
	return json.Unmarshal(e.Data, target)
}
