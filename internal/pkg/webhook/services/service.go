package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/pkg/webhook/models"
)

// WebhookService defines the interface for webhook operations
type WebhookService interface {
	// Subscribe creates a new webhook subscription
	Subscribe(ctx context.Context, name, callbackURL string, events []models.EventType, metadata map[string]string) (*models.WebhookSubscription, error)
	
	// Unsubscribe removes a webhook subscription
	Unsubscribe(ctx context.Context, subscriptionID string) error
	
	// GetSubscription retrieves a subscription by ID
	GetSubscription(ctx context.Context, subscriptionID string) (*models.WebhookSubscription, error)
	
	// ListSubscriptions returns all subscriptions, optionally filtered by event type
	ListSubscriptions(ctx context.Context, eventType models.EventType) ([]*models.WebhookSubscription, error)
	
	// ProcessEvent processes a webhook event and delivers it to relevant subscribers
	ProcessEvent(ctx context.Context, event *models.WebhookEvent) error
	
	// UpdateSubscription updates a subscription's details
	UpdateSubscription(ctx context.Context, subscriptionID string, updates map[string]interface{}) (*models.WebhookSubscription, error)
}

// DeliveryAttempt represents an attempt to deliver a webhook event
type DeliveryAttempt struct {
	ID             string
	SubscriptionID string
	EventID        string
	URL            string
	Status         string
	StatusCode     int
	ResponseBody   string
	Error          string
	AttemptedAt    time.Time
	Duration       time.Duration
}

// HTTPClient defines the interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// WebhookServiceImpl is the implementation of WebhookService
type WebhookServiceImpl struct {
	subscriptions map[string]*models.WebhookSubscription
	mtx           sync.RWMutex
	httpClient    HTTPClient
	deliveryQueue chan *deliveryRequest
	workers       int
}

type deliveryRequest struct {
	subscription *models.WebhookSubscription
	event       *models.WebhookEvent
}

// NewWebhookService creates a new webhook service
func NewWebhookService(httpClient HTTPClient, workers int) *WebhookServiceImpl {
	svc := &WebhookServiceImpl{
		subscriptions: make(map[string]*models.WebhookSubscription),
		httpClient:    httpClient,
		deliveryQueue: make(chan *deliveryRequest, 1000), // Buffer up to 1000 events
		workers:       workers,
	}

	// Start worker pool
	for i := 0; i < workers; i++ {
		go svc.worker()
	}

	return svc
}

// Subscribe creates a new webhook subscription
func (s *WebhookServiceImpl) Subscribe(ctx context.Context, name, callbackURL string, events []models.EventType, metadata map[string]string) (*models.WebhookSubscription, error) {
	sub := models.NewWebhookSubscription(name, callbackURL, events)
	sub.Metadata = metadata

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.subscriptions[sub.ID] = sub
	return sub, nil
}

// Unsubscribe removes a webhook subscription
func (s *WebhookServiceImpl) Unsubscribe(ctx context.Context, subscriptionID string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, exists := s.subscriptions[subscriptionID]; !exists {
		return fmt.Errorf("subscription not found")
	}

	delete(s.subscriptions, subscriptionID)
	return nil
}

// GetSubscription retrieves a subscription by ID
func (s *WebhookServiceImpl) GetSubscription(ctx context.Context, subscriptionID string) (*models.WebhookSubscription, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	sub, exists := s.subscriptions[subscriptionID]
	if !exists {
		return nil, fmt.Errorf("subscription not found")
	}

	return sub, nil
}

// ListSubscriptions returns all subscriptions, optionally filtered by event type
func (s *WebhookServiceImpl) ListSubscriptions(ctx context.Context, eventType models.EventType) ([]*models.WebhookSubscription, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	var result []*models.WebhookSubscription
	for _, sub := range s.subscriptions {
		if eventType == "" || sub.MatchesEvent(eventType) {
			result = append(result, sub)
		}
	}

	return result, nil
}

// ProcessEvent processes a webhook event and delivers it to relevant subscribers
func (s *WebhookServiceImpl) ProcessEvent(ctx context.Context, event *models.WebhookEvent) error {
	subs, err := s.ListSubscriptions(ctx, event.EventType)
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}

	// Queue the event for delivery to each subscriber
	for _, sub := range subs {
		if sub.IsActive() {
			s.deliveryQueue <- &deliveryRequest{
				subscription: sub,
				event:       event,
			}
		}
	}

	return nil
}

// UpdateSubscription updates a subscription's details
func (s *WebhookServiceImpl) UpdateSubscription(ctx context.Context, subscriptionID string, updates map[string]interface{}) (*models.WebhookSubscription, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	sub, exists := s.subscriptions[subscriptionID]
	if !exists {
		return nil, fmt.Errorf("subscription not found")
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "name":
			sub.Name = value.(string)
		case "callback_url":
			sub.CallbackURL = value.(string)
		case "status":
			sub.Status = models.SubscriptionStatus(value.(string))
		case "events":
			sub.Events = value.([]models.EventType)
		}
	}

	sub.UpdatedAt = time.Now().UTC()
	return sub, nil
}

// worker processes delivery requests from the queue
func (s *WebhookServiceImpl) worker() {
	for req := range s.deliveryQueue {
		s.deliverWebhook(req.subscription, req.event)
	}
}

// deliverWebhook delivers a webhook event to a subscriber
func (s *WebhookServiceImpl) deliverWebhook(sub *models.WebhookSubscription, event *models.WebhookEvent) {
	// Create request body
	payload, err := json.Marshal(map[string]interface{}{
		"id":        event.ID,
		"event":     event.EventType,
		"timestamp": event.Timestamp,
		"data":      event.Data,
	})
	if err != nil {
		// Log error
		return
	}

	// Create request
	req, err := http.NewRequest("POST", sub.CallbackURL, bytes.NewReader(payload))
	if err != nil {
		// Log error
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Rayda-Webhook-Service/1.0")
	req.Header.Set("X-Webhook-Event", string(event.EventType))
	req.Header.Set("X-Webhook-Delivery", uuid.New().String())

	// Add HMAC signature if secret is set
	if sub.Secret != "" {
		signature := generateHMAC(payload, []byte(sub.Secret))
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Send request
	start := time.Now()
	resp, err := s.httpClient.Do(req)
	duration := time.Since(start)

	// Create delivery attempt
	attempt := &DeliveryAttempt{
		ID:             uuid.New().String(),
		SubscriptionID: sub.ID,
		EventID:        event.ID,
		URL:            sub.CallbackURL,
		AttemptedAt:    time.Now().UTC(),
		Duration:       duration,
	}

	// Handle response
	if err != nil {
		attempt.Status = "failed"
		attempt.Error = err.Error()
	} else {
		defer resp.Body.Close()
		attempt.Status = "delivered"
		attempt.StatusCode = resp.StatusCode

		// Read response body (up to 1MB)
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB
		attempt.ResponseBody = string(body)

		// Consider non-2xx status codes as failures
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			attempt.Status = "failed"
			attempt.Error = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		}
	}

	// Update subscription last delivered time
	s.mtx.Lock()
	if sub, exists := s.subscriptions[sub.ID]; exists {
		now := time.Now().UTC()
		sub.LastDelivered = &now
		sub.UpdatedAt = now
	}
	s.mtx.Unlock()

	// TODO: Store delivery attempt in database
}

// generateHMAC generates an HMAC signature for the payload
func generateHMAC(payload, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}
