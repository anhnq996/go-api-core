package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User entity
type User struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name            string         `json:"name" gorm:"type:varchar(255);not null"`
	Email           string         `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Password        string         `json:"-" gorm:"type:varchar(255)"` // Không trả về trong JSON
	Avatar          *string        `json:"avatar" gorm:"type:varchar(500)"`
	RoleID          *uuid.UUID     `json:"role_id" gorm:"type:uuid"`
	Role            *Role          `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at"`
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	LastLoginAt     *time.Time     `json:"last_login_at"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"` // Soft delete
}

// TableName override tên bảng
func (User) TableName() string {
	return "users"
}

// UserWithPermissions User model kèm permissions
type UserWithPermissions struct {
	User
	Permissions []string `json:"permissions" gorm:"-"` // Không map vào database
}
