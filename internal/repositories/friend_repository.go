package repository

import (
	"context"

	model "api-core/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FriendRequestRepository interface
type FriendRequestRepository interface {
	Repository[model.FriendRequest]

	FindBySenderAndReceiver(ctx context.Context, senderID, receiverID uuid.UUID) (*model.FriendRequest, error)
	FindPendingByReceiver(ctx context.Context, receiverID uuid.UUID) ([]model.FriendRequest, error)
	FindPendingBySender(ctx context.Context, senderID uuid.UUID) ([]model.FriendRequest, error)
	FindByStatus(ctx context.Context, senderID uuid.UUID, status model.FriendRequestStatus) ([]model.FriendRequest, error)
}

// FriendshipRepository interface
type FriendshipRepository interface {
	Repository[model.Friendship]

	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Friendship, error)
	FindFriendship(ctx context.Context, userID, friendID uuid.UUID) (*model.Friendship, error)
	IsFriend(ctx context.Context, userID, friendID uuid.UUID) (bool, error)
}

// friendRequestRepository implementation
type friendRequestRepository struct {
	*BaseRepository[model.FriendRequest]
}

// NewFriendRequestRepository tạo friend request repository mới
func NewFriendRequestRepository(db *gorm.DB) FriendRequestRepository {
	return &friendRequestRepository{
		BaseRepository: NewBaseRepository[model.FriendRequest](db, true),
	}
}

// FindBySenderAndReceiver tìm friend request theo sender và receiver
func (r *friendRequestRepository) FindBySenderAndReceiver(ctx context.Context, senderID, receiverID uuid.UUID) (*model.FriendRequest, error) {
	return r.FirstWhere(ctx, "(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		senderID, receiverID, receiverID, senderID)
}

// FindPendingByReceiver tìm các lời mời pending mà user nhận được
func (r *friendRequestRepository) FindPendingByReceiver(ctx context.Context, receiverID uuid.UUID) ([]model.FriendRequest, error) {
	return r.FindWhere(ctx, "receiver_id = ? AND status = ?", receiverID, model.FriendRequestStatusPending)
}

// FindPendingBySender tìm các lời mời pending mà user đã gửi
func (r *friendRequestRepository) FindPendingBySender(ctx context.Context, senderID uuid.UUID) ([]model.FriendRequest, error) {
	return r.FindWhere(ctx, "sender_id = ? AND status = ?", senderID, model.FriendRequestStatusPending)
}

// FindByStatus tìm friend requests theo status
func (r *friendRequestRepository) FindByStatus(ctx context.Context, senderID uuid.UUID, status model.FriendRequestStatus) ([]model.FriendRequest, error) {
	return r.FindWhere(ctx, "(sender_id = ? OR receiver_id = ?) AND status = ?", senderID, senderID, status)
}

// friendshipRepository implementation
type friendshipRepository struct {
	*BaseRepository[model.Friendship]
}

// NewFriendshipRepository tạo friendship repository mới
func NewFriendshipRepository(db *gorm.DB) FriendshipRepository {
	return &friendshipRepository{
		BaseRepository: NewBaseRepository[model.Friendship](db, true),
	}
}

// FindByUserID tìm tất cả bạn bè của user
func (r *friendshipRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Friendship, error) {
	return r.FindWhere(ctx, "user_id = ? OR friend_id = ?", userID, userID)
}

// FindFriendship tìm quan hệ bạn bè giữa 2 user
func (r *friendshipRepository) FindFriendship(ctx context.Context, userID, friendID uuid.UUID) (*model.Friendship, error) {
	return r.FirstWhere(ctx, "(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friendID, friendID, userID)
}

// IsFriend kiểm tra 2 user có phải bạn bè không
func (r *friendshipRepository) IsFriend(ctx context.Context, userID, friendID uuid.UUID) (bool, error) {
	friendship, err := r.FindFriendship(ctx, userID, friendID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return friendship != nil, nil
}
