package communication

import (
	"context"
	"time"
)

// Message represents a communication message to be sent
type Message struct {
	ID          string                 `json:"id"`
	To          []string               `json:"to"`
	Subject     string                 `json:"subject,omitempty"`
	Body        string                 `json:"body"`
	TemplateID  string                 `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	Channel     string                 `json:"channel"` // "email", "sms", "push", etc.
	Status      string                 `json:"status"`  // "queued", "sent", "delivered", "failed"
	CreatedAt   time.Time              `json:"created_at"`
	SentAt      *time.Time             `json:"sent_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// MessageOptions contains optional parameters for sending a message
type MessageOptions struct {
	From        string                 `json:"from,omitempty"`
	CC          []string               `json:"cc,omitempty"`
	BCC         []string               `json:"bcc,omitempty"`
	Attachments []Attachment           `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Attachment represents a file attachment for a message
type Attachment struct {
	Content     []byte `json:"content"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

// Client defines the interface for interacting with an external communication service
type Client interface {
	// Send a message
	SendMessage(ctx context.Context, message *Message, opts *MessageOptions) (string, error)
	
	// Get message status
	GetMessageStatus(ctx context.Context, messageID string) (*Message, error)
	
	// Template management
	CreateTemplate(ctx context.Context, name, content string, variables []string) (string, error)
	UpdateTemplate(ctx context.Context, templateID, content string, variables []string) error
	DeleteTemplate(ctx context.Context, templateID string) error
	
	// Webhook handling
	RegisterWebhook(ctx context.Context, url string, events []string) (string, error)
	UnregisterWebhook(ctx context.Context, webhookID string) error
}

// MockClient is a mock implementation of the Client interface for testing
type MockClient struct {
	Messages      map[string]*Message
	Templates     map[string]string
	TemplateVars  map[string][]string
	Webhooks      map[string][]string
}

// NewMockClient creates a new mock communication client
func NewMockClient() *MockClient {
	return &MockClient{
		Messages:     make(map[string]*Message),
		Templates:    make(map[string]string),
		TemplateVars: make(map[string][]string),
		Webhooks:     make(map[string][]string),
	}
}

// SendMessage sends a message
func (m *MockClient) SendMessage(ctx context.Context, msg *Message, opts *MessageOptions) (string, error) {
	msg.ID = "msg_" + time.Now().Format("20060102150405") + "_" + randomString(8)
	msg.Status = "sent"
	now := time.Now()
	msg.CreatedAt = now
	sentAt := now.Add(100 * time.Millisecond)
	msg.SentAt = &sentAt
	
	m.Messages[msg.ID] = msg
	return msg.ID, nil
}

// GetMessageStatus gets the status of a message
func (m *MockClient) GetMessageStatus(ctx context.Context, messageID string) (*Message, error) {
	msg, exists := m.Messages[messageID]
	if !exists {
		return nil, nil // Not found
	}
	return msg, nil
}

// CreateTemplate creates a new message template
func (m *MockClient) CreateTemplate(ctx context.Context, name, content string, variables []string) (string, error) {
	templateID := "tpl_" + randomString(8)
	m.Templates[templateID] = content
	m.TemplateVars[templateID] = variables
	return templateID, nil
}

// UpdateTemplate updates an existing template
func (m *MockClient) UpdateTemplate(ctx context.Context, templateID, content string, variables []string) error {
	if _, exists := m.Templates[templateID]; !exists {
		return nil // Not found
	}
	m.Templates[templateID] = content
	m.TemplateVars[templateID] = variables
	return nil
}

// DeleteTemplate deletes a template
func (m *MockClient) DeleteTemplate(ctx context.Context, templateID string) error {
	delete(m.Templates, templateID)
	delete(m.TemplateVars, templateID)
	return nil
}

// RegisterWebhook registers a webhook for event notifications
func (m *MockClient) RegisterWebhook(ctx context.Context, url string, events []string) (string, error) {
	webhookID := "wh_" + randomString(12)
	m.Webhooks[webhookID] = events
	return webhookID, nil
}

// UnregisterWebhook removes a webhook
func (m *MockClient) UnregisterWebhook(ctx context.Context, webhookID string) error {
	delete(m.Webhooks, webhookID)
	return nil
}

// Helper function to generate random strings
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
