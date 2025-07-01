package models

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionStatus represents the status of a webhook subscription
type SubscriptionStatus string

const (
	// SubscriptionStatusActive indicates the subscription is active
	SubscriptionStatusActive SubscriptionStatus = "active"
	// SubscriptionStatusPaused indicates the subscription is paused
	SubscriptionStatusPaused SubscriptionStatus = "paused"
	// SubscriptionStatusInactive indicates the subscription is inactive
	SubscriptionStatusInactive SubscriptionStatus = "inactive"
)

// WebhookSubscription represents a subscription to webhook events
type WebhookSubscription struct {
	ID            string            `json:"id"`
	Name          string            `json:
ame"`
	Description   string            `json:"description,omitempty"`
	CallbackURL   string            `json:"callback_url"`
	Events        []EventType       `json:"events"`
	Secret        string            `json:"-"` // Used for HMAC signature
	Status        SubscriptionStatus `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	LastDelivered *time.Time        `json:"last_delivered,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// NewWebhookSubscription creates a new webhook subscription
func NewWebhookSubscription(name, callbackURL string, events []EventType) *WebhookSubscription {
	now := time.Now().UTC()
	return &WebhookSubscription{
		ID:          uuid.New().String(),
		Name:        name,
		CallbackURL: callbackURL,
		Events:      events,
		Status:      SubscriptionStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsActive checks if the subscription is active
func (s *WebhookSubscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive
}

// MatchesEvent checks if the subscription is interested in the given event type
func (s *WebhookSubscription) MatchesEvent(eventType EventType) bool {
	for _, et := range s.Events {
		if et == eventType || et == "*" {
			return true
		}
	}
	return false
}
