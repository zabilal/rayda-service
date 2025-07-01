package model

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleMember     Role = "member"
)

type User struct {
	TenantModel
	Email          string    `json:"email" gorm:"uniqueIndex:idx_tenant_email;not null"`
	HashedPassword string    `json:"-" gorm:"not null"`
	FirstName      string    `json:"first_name" gorm:"not null"`
	LastName       string    `json:"last_name" gorm:"not null"`
	Role           Role      `json:"role" gorm:"type:varchar(20);not null;default:'member'"`
	LastLoginAt    time.Time `json:"last_login_at,omitempty"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
}

// NewUser creates a new user with hashed password
func NewUser(email, password, firstName, lastName string, role Role, tenantID uuid.UUID) (*User, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		TenantModel: TenantModel{
			Base: Base{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			TenantID: tenantID,
		},
		Email:          email,
		HashedPassword: hashedPassword,
		FirstName:      firstName,
		LastName:       lastName,
		Role:           role,
		IsActive:       true,
	}

	return user, nil
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hashed, err := hashPassword(password)
	if err != nil {
		return err
	}
	u.HashedPassword = hashed
	return nil
}

// CheckPassword verifies if the provided password matches the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
