package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationType enum type
type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct"
	ConversationTypeGroup  ConversationType = "group"
)

// Conversation entity
type Conversation struct {
	ID        uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type      ConversationType `json:"type" gorm:"type:conversation_type;default:'direct'"`
	Name      *string          `json:"name" gorm:"type:varchar(255)"`
	Avatar    *string          `json:"avatar" gorm:"type:varchar(500)"`
	CreatedBy *uuid.UUID       `json:"created_by" gorm:"type:uuid"`
	CreatedAt time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt   `json:"-" gorm:"index"`

	// Relations
	Creator      *User                     `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Participants []ConversationParticipant `json:"participants,omitempty" gorm:"foreignKey:ConversationID"`
	Messages     []Message                 `json:"messages,omitempty" gorm:"foreignKey:ConversationID"`
}

// TableName override tên bảng
func (Conversation) TableName() string {
	return "conversations"
}
