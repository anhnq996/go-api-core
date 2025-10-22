package model

import (
	"time"

	"github.com/google/uuid"
)

// RoleHasPermission model (bảng trung gian)
type RoleHasPermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;not null" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null" json:"permission_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName chỉ định tên bảng
func (RoleHasPermission) TableName() string {
	return "role_has_permissions"
}
