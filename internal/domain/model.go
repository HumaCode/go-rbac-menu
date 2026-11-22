package domain

import (
	"time"

	"go-rbac-menu/pkg/utils"
	"gorm.io/gorm"
)

// BaseModel contains common fields for all entities using ULID as PK
type BaseModel struct {
	ID        string    `gorm:"type:varchar(26);primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hook to generate ULID automatically if empty
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = utils.NewULID()
	}
	return nil
}

