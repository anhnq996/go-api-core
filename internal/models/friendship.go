package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Friendship entity
type Friendship struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	FriendID  uuid.UUID      `json:"friend_id" gorm:"type:uuid;not null"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User   *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Friend *User `json:"friend,omitempty" gorm:"foreignKey:FriendID"`
}

// TableName override tên bảng
func (Friendship) TableName() string {
	return "friendships"
}
