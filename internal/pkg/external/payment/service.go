package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Payment represents a payment in the external payment system
type Payment struct {
	ID            string    `json:"id"`
	Amount        int64     `json:"amount"` // in smallest currency unit (e.g., cents)
	Currency      string    `json:"currency"`
	CustomerID    string    `json:"customer_id"`
	Description   string    `json:"description"`
	Status        string    `json:"status"` // e.g., "succeeded", "pending", "failed"
	PaymentMethod string    `json:"payment_method"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Metadata      string    `json:"metadata,omitempty"`
}

// Refund represents a refund of a payment
type Refund struct {
	ID        string    `json:"id"`
	PaymentID string    `json:"payment_id"`
	Amount    int64     `json:"amount"`
	Reason    string    `json:"reason,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Client defines the interface for interacting with an external payment service
type Client interface {
	// Payment operations
	CreatePayment(ctx context.Context, amount int64, currency, customerID, description string) (*Payment, error)
	GetPayment(ctx context.Context, paymentID string) (*Payment, error)
	ListPayments(ctx context.Context, filter map[string]interface{}) ([]*Payment, error)
	
	// Refund operations
	CreateRefund(ctx context.Context, paymentID string, amount int64, reason string) (*Refund, error)
	GetRefund(ctx context.Context, refundID string) (*Refund, error)
	ListRefunds(ctx context.Context, paymentID string) ([]*Refund, error)
	
	// Payment method operations
	CreatePaymentMethod(ctx context.Context, customerID, paymentMethodType string, details map[string]interface{}) (string, error)
	GetPaymentMethod(ctx context.Context, paymentMethodID string) (map[string]interface{}, error)
	DeletePaymentMethod(ctx context.Context, paymentMethodID string) error
}

// MockClient is a mock implementation of the Client interface for testing
type MockClient struct {
	Payments  map[string]*Payment
	Refunds   map[string]*Refund
	Customers map[string]map[string]interface{} // customerID -> payment methods
}

// NewMockClient creates a new mock payment client
func NewMockClient() *MockClient {
	return &MockClient{
		Payments:  make(map[string]*Payment),
		Refunds:   make(map[string]*Refund),
		Customers: make(map[string]map[string]interface{}),
	}
}

// CreatePayment creates a new payment
func (m *MockClient) CreatePayment(ctx context.Context, amount int64, currency, customerID, description string) (*Payment, error) {
	payment := &Payment{
		ID:            "pay_" + uuid.New().String(),
		Amount:        amount,
		Currency:      currency,
		CustomerID:    customerID,
		Description:   description,
		Status:        "succeeded",
		PaymentMethod: "card_" + uuid.New().String()[:8],
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	m.Payments[payment.ID] = payment
	return payment, nil
}

// GetPayment retrieves a payment by ID
func (m *MockClient) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	if payment, exists := m.Payments[paymentID]; exists {
		return payment, nil
	}
	return nil, nil // Simulate "not found"
}

// ListPayments lists payments based on filters
func (m *MockClient) ListPayments(ctx context.Context, filter map[string]interface{}) ([]*Payment, error) {
	var result []*Payment
	for _, payment := range m.Payments {
		result = append(result, payment)
	}
	return result, nil
}

// CreateRefund creates a refund for a payment
func (m *MockClient) CreateRefund(ctx context.Context, paymentID string, amount int64, reason string) (*Refund, error) {
	_, exists := m.Payments[paymentID]
	if !exists {
		return nil, nil // Payment not found
	}

	refund := &Refund{
		ID:        "re_" + uuid.New().String(),
		PaymentID: paymentID,
		Amount:    amount,
		Reason:    reason,
		Status:    "succeeded",
		CreatedAt: time.Now(),
	}

	m.Refunds[refund.ID] = refund
	return refund, nil
}

// GetRefund retrieves a refund by ID
func (m *MockClient) GetRefund(ctx context.Context, refundID string) (*Refund, error) {
	if refund, exists := m.Refunds[refundID]; exists {
		return refund, nil
	}
	return nil, nil // Simulate "not found"
}

// ListRefunds lists refunds for a payment
func (m *MockClient) ListRefunds(ctx context.Context, paymentID string) ([]*Refund, error) {
	var result []*Refund
	for _, refund := range m.Refunds {
		if refund.PaymentID == paymentID {
			result = append(result, refund)
		}
	}
	return result, nil
}

// CreatePaymentMethod adds a payment method for a customer
func (m *MockClient) CreatePaymentMethod(ctx context.Context, customerID, paymentMethodType string, details map[string]interface{}) (string, error) {
	if _, exists := m.Customers[customerID]; !exists {
		m.Customers[customerID] = make(map[string]interface{})
	}

	methodID := "pm_" + uuid.New().String()
	m.Customers[customerID][methodID] = map[string]interface{}{
		"id":      methodID,
		"type":    paymentMethodType,
		"details": details,
		"created": time.Now().Unix(),
	}

	return methodID, nil
}

// GetPaymentMethod retrieves a payment method
func (m *MockClient) GetPaymentMethod(ctx context.Context, paymentMethodID string) (map[string]interface{}, error) {
	for _, methods := range m.Customers {
		if method, exists := methods[paymentMethodID]; exists {
			return method.(map[string]interface{}), nil
		}
	}
	return nil, nil // Not found
}

// DeletePaymentMethod removes a payment method
func (m *MockClient) DeletePaymentMethod(ctx context.Context, paymentMethodID string) error {
	for _, methods := range m.Customers {
		if _, exists := methods[paymentMethodID]; exists {
			delete(methods, paymentMethodID)
			return nil
		}
	}
	return nil // No-op if not found
}
