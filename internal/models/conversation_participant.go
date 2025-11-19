package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationParticipant entity
type ConversationParticipant struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID      `json:"conversation_id" gorm:"type:uuid;not null"`
	UserID         uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	LastReadAt     *time.Time     `json:"last_read_at"`
	JoinedAt       time.Time      `json:"joined_at" gorm:"autoCreateTime"`
	LeftAt         *time.Time     `json:"left_at"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Conversation *Conversation `json:"conversation,omitempty" gorm:"foreignKey:ConversationID"`
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName override tên bảng
func (ConversationParticipant) TableName() string {
	return "conversation_participants"
}
