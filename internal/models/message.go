package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageType enum type
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeFile     MessageType = "file"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeLocation MessageType = "location"
	MessageTypeSystem   MessageType = "system"
)

// Message entity
type Message struct {
	ID             uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID              `json:"conversation_id" gorm:"type:uuid;not null"`
	SenderID       uuid.UUID              `json:"sender_id" gorm:"type:uuid;not null"`
	Content        string                 `json:"content" gorm:"type:text;not null"`
	MessageType    MessageType            `json:"message_type" gorm:"type:message_type;default:'text'"`
	ReplyToID      *uuid.UUID             `json:"reply_to_id" gorm:"type:uuid"`
	FileURL        *string                `json:"file_url" gorm:"type:varchar(500)"`
	FileName       *string                `json:"file_name" gorm:"type:varchar(255)"`
	FileSize       *int64                 `json:"file_size" gorm:"type:bigint"`
	Metadata       map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt      time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt         `json:"-" gorm:"index"`

	// Relations
	Conversation *Conversation `json:"conversation,omitempty" gorm:"foreignKey:ConversationID"`
	Sender       *User         `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	ReplyTo      *Message      `json:"reply_to,omitempty" gorm:"foreignKey:ReplyToID"`
}

// TableName override tên bảng
func (Message) TableName() string {
	return "messages"
}
