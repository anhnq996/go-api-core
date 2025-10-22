package model

import (
	"time"

	"github.com/google/uuid"
)

// Role model
type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string       `gorm:"type:varchar(50);not null;unique" json:"name"`
	DisplayName string       `gorm:"type:varchar(100);not null" json:"display_name"`
	Description string       `gorm:"type:text" json:"description"`
	Permissions []Permission `gorm:"many2many:role_has_permissions;" json:"permissions,omitempty"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName chỉ định tên bảng
func (Role) TableName() string {
	return "roles"
}
