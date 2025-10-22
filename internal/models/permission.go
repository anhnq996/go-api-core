package model

import (
	"time"

	"github.com/google/uuid"
)

// Permission model
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	DisplayName string    `gorm:"type:varchar(150);not null" json:"display_name"`
	Description string    `gorm:"type:text" json:"description"`
	Module      string    `gorm:"type:varchar(50)" json:"module"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName chỉ định tên bảng
func (Permission) TableName() string {
	return "permissions"
}
