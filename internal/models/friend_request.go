package model

import (
	"time"

	"github.com/google/uuid"
)

// FriendRequestStatus enum type
type FriendRequestStatus string

const (
	FriendRequestStatusPending   FriendRequestStatus = "pending"
	FriendRequestStatusAccepted  FriendRequestStatus = "accepted"
	FriendRequestStatusRejected  FriendRequestStatus = "rejected"
	FriendRequestStatusCancelled FriendRequestStatus = "cancelled"
)

// FriendRequest entity
type FriendRequest struct {
	ID         uuid.UUID           `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SenderID   uuid.UUID           `json:"sender_id" gorm:"type:uuid;not null"`
	ReceiverID uuid.UUID           `json:"receiver_id" gorm:"type:uuid;not null"`
	Status     FriendRequestStatus `json:"status" gorm:"type:friend_request_status;default:'pending'"`
	CreatedAt  time.Time           `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time           `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Sender   *User `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	Receiver *User `json:"receiver,omitempty" gorm:"foreignKey:ReceiverID"`
}

// TableName override tên bảng
func (FriendRequest) TableName() string {
	return "friend_requests"
}
