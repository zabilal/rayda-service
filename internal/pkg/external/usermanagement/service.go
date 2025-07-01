package usermanagement

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the external user management system
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLoginAt  time.Time `json:"last_login_at,omitempty"`
	ProfileImage string    `json:"profile_image,omitempty"`
	Metadata     string    `json:"metadata,omitempty"`
}

// Client defines the interface for interacting with an external user management service
type Client interface {
	// User operations
	GetUser(ctx context.Context, userID string) (*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) (*User, error)
	DeleteUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, filter map[string]interface{}) ([]*User, error)

	// Authentication
	Authenticate(ctx context.Context, email, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (*User, error)
	InvalidateToken(ctx context.Context, token string) error

	// User groups/roles
	AddUserToGroup(ctx context.Context, userID, groupID string) error
	RemoveUserFromGroup(ctx context.Context, userID, groupID string) error
	ListUserGroups(ctx context.Context, userID string) ([]string, error)
}

// MockClient is a mock implementation of the Client interface for testing
type MockClient struct {
	Users map[string]*User
}

// NewMockClient creates a new mock user management client
func NewMockClient() *MockClient {
	return &MockClient{
		Users: make(map[string]*User),
	}
}

// GetUser retrieves a user by ID
func (m *MockClient) GetUser(ctx context.Context, userID string) (*User, error) {
	if user, exists := m.Users[userID]; exists {
		return user, nil
	}
	return nil, nil // Simulate "not found"
}

// CreateUser creates a new user
func (m *MockClient) CreateUser(ctx context.Context, user *User) (*User, error) {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt
	m.Users[user.ID] = user
	return user, nil
}

// UpdateUser updates a user's information
func (m *MockClient) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) (*User, error) {
	user, exists := m.Users[userID]
	if !exists {
		return nil, nil
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "email":
			user.Email = value.(string)
		case "first_name":
			user.FirstName = value.(string)
		case "last_name":
			user.LastName = value.(string)
		case "active":
			user.Active = value.(bool)
		}
	}

	user.UpdatedAt = time.Now()
	return user, nil
}

// DeleteUser removes a user
func (m *MockClient) DeleteUser(ctx context.Context, userID string) error {
	delete(m.Users, userID)
	return nil
}

// ListUsers lists users based on filters
func (m *MockClient) ListUsers(ctx context.Context, filter map[string]interface{}) ([]*User, error) {
	var result []*User
	for _, user := range m.Users {
		result = append(result, user)
	}
	return result, nil
}

// Authenticate authenticates a user and returns a token
func (m *MockClient) Authenticate(ctx context.Context, email, password string) (string, error) {
	for _, user := range m.Users {
		if user.Email == email {
			// In a real implementation, we'd verify the password
			return "mock_token_" + user.ID, nil
		}
	}
	return "", nil // Simulate "invalid credentials"
}

// ValidateToken validates a user token
func (m *MockClient) ValidateToken(ctx context.Context, token string) (*User, error) {
	// Simple token format: "mock_token_<user_id>"
	if len(token) > 11 && token[:11] == "mock_token_" {
		userID := token[11:]
		return m.GetUser(ctx, userID)
	}
	return nil, nil // Invalid token format
}

// InvalidateToken invalidates a user's authentication token
func (m *MockClient) InvalidateToken(ctx context.Context, token string) error {
	// In a real implementation, we'd add this to a blacklist
	return nil
}

// AddUserToGroup adds a user to a group
func (m *MockClient) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	// Implementation would maintain group memberships
	return nil
}

// RemoveUserFromGroup removes a user from a group
func (m *MockClient) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	// Implementation would update group memberships
	return nil
}

// ListUserGroups lists all groups a user belongs to
func (m *MockClient) ListUserGroups(ctx context.Context, userID string) ([]string, error) {
	// Return empty list for mock
	return []string{}, nil
}
