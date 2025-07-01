package model

import (
	"time"

	"github.com/google/uuid"
)

// Base contains common fields for all models
type Base struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"not null"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" sql:"index"`
}

// TenantModel contains fields for tenant isolation
type TenantModel struct {
	Base
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
}

// BeforeCreate is a hook that runs before creating a record
func (b *Base) BeforeCreate() error {
	now := time.Now()
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	b.CreatedAt = now
	b.UpdatedAt = now
	return nil
}

// BeforeUpdate is a hook that runs before updating a record
func (b *Base) BeforeUpdate() error {
	b.UpdatedAt = time.Now()
	return nil
}
