package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/api/dto"
	"github.com/rayda/rayda-service/internal/model"
	"github.com/rayda/rayda-service/internal/pkg/auth"
	"github.com/rayda/rayda-service/internal/repository"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userRepo repository.UserRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// RegisterRoutes registers the user routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	usersGroup := router.Group("/users")
	usersGroup.Use(authMiddleware)
	{
		usersGroup.GET("", h.ListUsers)
		usersGroup.POST("", h.CreateUser)
		usersGroup.GET("/:id", h.GetUser)
		usersGroup.PUT("/:id", h.UpdateUser)
		usersGroup.DELETE("/:id", h.DeleteUser)
	}
}

// ListUsers returns a list of users for the current tenant
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Get tenant ID from context
	tenantID, err := auth.GetCurrentTenantID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get users for the current tenant
	users, err := h.userRepo.FindByTenantID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	// Convert []model.User to []*model.User
	userPtrs := make([]*model.User, len(users))
	for i := range users {
		userPtrs[i] = &users[i]
	}

	c.JSON(http.StatusOK, dto.ToUserResponseList(userPtrs))
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Get tenant ID from context
	tenantID, err := auth.GetCurrentTenantID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	existingUser, _ := h.userRepo.FindByEmail(c.Request.Context(), req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	// Create new user
	user, err := model.NewUser(
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
		req.Role,
		tenantID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Save user to database
	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user"})
		return
	}

	c.JSON(http.StatusCreated, dto.ToUserResponse(user))
}

// GetUser returns a user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	// Get tenant ID from context
	tenantID, err := auth.GetCurrentTenantID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse user ID from URL
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get user by ID and tenant ID
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Verify the user belongs to the current tenant
	if user.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// UpdateUser updates a user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Get tenant ID from context
	tenantID, err := auth.GetCurrentTenantID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse user ID from URL
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get user by ID and tenant ID
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Verify the user belongs to the current tenant
	if user.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user fields if provided
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Password != nil {
		if err := user.SetPassword(*req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
			return
		}
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Save updated user
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Get tenant ID from context
	tenantID, err := auth.GetCurrentTenantID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse user ID from URL
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get user by ID and tenant ID
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Verify the user belongs to the current tenant
	if user.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Delete user
	if err := h.userRepo.Delete(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.Status(http.StatusNoContent)
}
