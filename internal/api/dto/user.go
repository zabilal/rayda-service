package dto

import (
	"time"

	"github.com/rayda/rayda-service/internal/model"
)

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password,omitempty" binding:"required,min=8"`
	FirstName string    `json:"first_name" binding:"required"`
	LastName  string    `json:"last_name" binding:"required"`
	Role      model.Role `json:"role" binding:"required,oneof=admin member"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Email     *string    `json:"email,omitempty" binding:"omitempty,email"`
	Password  *string    `json:"password,omitempty" binding:"omitempty,min=8"`
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	Role      *model.Role `json:"role,omitempty" binding:"omitempty,oneof=admin member"`
	IsActive  *bool       `json:"is_active,omitempty"`
}

// UserResponse represents the user response
type UserResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Role      model.Role `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// ToUserResponse converts a model.User to a UserResponse
func ToUserResponse(user *model.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}

// ToUserResponseList converts a slice of model.User to a slice of UserResponse
func ToUserResponseList(users []*model.User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i, user := range users {
		result[i] = ToUserResponse(user)
	}
	return result
}
