package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/model"
)

// Claims represents the JWT claims
type Claims struct {
	UserID    uuid.UUID    `json:"user_id"`
	TenantID  uuid.UUID    `json:"tenant_id"`
	Email     string       `json:"email"`
	Role      model.Role   `json:"role"`
	ExpiresAt int64        `json:"exp"`
	jwt.RegisteredClaims
}

// TokenPair contains access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// Service handles authentication and authorization operations
type Service struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewService creates a new auth service
func NewService(secretKey string, accessExpiry, refreshExpiry time.Duration) *Service {
	return &Service{
		secretKey:     []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateTokenPair generates a new access and refresh token pair
func (s *Service) GenerateTokenPair(user *model.User) (*TokenPair, error) {
	now := time.Now()
	accessExpiresAt := now.Add(s.accessExpiry).Unix()
	refreshExpiresAt := now.Add(s.refreshExpiry).Unix()

	// Create access token
	accessClaims := &Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		ExpiresAt: accessExpiresAt,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(accessExpiresAt, 0)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		return nil, err
	}

	// Create refresh token
	refreshClaims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Unix(refreshExpiresAt, 0)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Subject:   user.ID.String(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

// ParseToken parses and validates a JWT token
func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new access token using a refresh token
func (s *Service) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	token, err := jwt.ParseWithClaims(refreshTokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// In a real application, you would look up the user by ID and validate the refresh token
		// For now, we'll just generate a new token pair with the same user ID
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return nil, errors.New("invalid user ID in token")
		}

		// Create a minimal user with just the ID to generate tokens
		// In a real app, you would fetch the full user from the database
		user := &model.User{
			TenantModel: model.TenantModel{
				Base: model.Base{
					ID: userID,
				},
			},
			// These would be set from the database in a real app
			Email: "",
			Role:  model.RoleMember,
		}

		return s.GenerateTokenPair(user)
	}

	return nil, errors.New("invalid refresh token")
}

// AccessExpiry returns the duration for which access tokens are valid
func (s *Service) AccessExpiry() time.Duration {
	return s.accessExpiry
}

// HasPermission checks if a user has the required role
func (s *Service) HasPermission(requiredRole model.Role, userRole model.Role) bool {
	roleHierarchy := map[model.Role]int{
		model.RoleSuperAdmin: 3,
		model.RoleAdmin:      2,
		model.RoleMember:     1,
	}

	requiredLevel, ok := roleHierarchy[requiredRole]
	if !ok {
		return false
	}

	userLevel, ok := roleHierarchy[userRole]
	if !ok {
		return false
	}

	return userLevel >= requiredLevel
}
