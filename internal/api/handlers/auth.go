package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rayda/rayda-service/internal/api/dto"
	"github.com/rayda/rayda-service/internal/model"
	"github.com/rayda/rayda-service/internal/pkg/auth"
	"github.com/rayda/rayda-service/internal/repository"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authSvc *auth.Service
	userRepo repository.UserRepository
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authSvc *auth.Service, userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		authSvc:  authSvc,
		userRepo: userRepo,
	}
}

// RegisterRoutes registers the authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", h.Login)
		authGroup.POST("/register", h.Register)
		authGroup.POST("/refresh", h.RefreshToken)
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	user, err := h.userRepo.FindByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Generate tokens
	tokens, err := h.authSvc.GenerateTokenPair(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Convert to DTO token pair
	tokenPair := &dto.TokenPair{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	// Update last login time (in a real app, you would update this in the database)

	c.JSON(http.StatusOK, dto.NewAuthResponse(tokenPair, user, h.authSvc.AccessExpiry()))
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
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

	// Create a new organization for the user (in a real app, you might have a different flow)
	org := model.NewOrganization("My Organization", "", req.Email)

	// Create the user
	user, err := model.NewUser(
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
		model.RoleAdmin, // First user is an admin
		org.ID,          // Set the tenant ID to the organization ID
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Save the user to the database (in a real app, you would also save the organization)
	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user"})
		return
	}

	// Generate tokens
	tokens, err := h.authSvc.GenerateTokenPair(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Convert to DTO token pair
	tokenPair := &dto.TokenPair{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	c.JSON(http.StatusCreated, dto.NewAuthResponse(tokenPair, user, h.authSvc.AccessExpiry()))
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tokenPair, err := h.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Convert to DTO token pair
	tokenPairDTO := &dto.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}

	// Get the token expiration time from the claims
	token, _, err := new(jwt.Parser).ParseUnverified(tokenPair.AccessToken, &auth.Claims{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse token"})
		return
	}

	claims, ok := token.Claims.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid token claims"})
		return
	}

	// Get the user from the database
	user, err := h.userRepo.FindByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, dto.NewAuthResponse(tokenPairDTO, user, h.authSvc.AccessExpiry()))
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context
	userID, err := auth.GetCurrentUserID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get user from database
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, dto.UserInfo{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		TenantID:  user.TenantID.String(),
	})
}
