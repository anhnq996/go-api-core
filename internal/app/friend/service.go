package friend

import (
	"context"
	"fmt"

	model "api-core/internal/models"
	repository "api-core/internal/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service xử lý business logic cho friend
type Service struct {
	friendRequestRepo repository.FriendRequestRepository
	friendshipRepo    repository.FriendshipRepository
	userRepo          repository.UserRepository
	db                *gorm.DB
}

// NewService tạo friend service mới
func NewService(
	friendRequestRepo repository.FriendRequestRepository,
	friendshipRepo repository.FriendshipRepository,
	userRepo repository.UserRepository,
	db *gorm.DB,
) *Service {
	return &Service{
		friendRequestRepo: friendRequestRepo,
		friendshipRepo:    friendshipRepo,
		userRepo:          userRepo,
		db:                db,
	}
}

// SendFriendRequest gửi lời mời kết bạn
func (s *Service) SendFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) (*model.FriendRequest, error) {
	// Kiểm tra không thể tự gửi lời mời cho chính mình
	if senderID == receiverID {
		return nil, fmt.Errorf("không thể gửi lời mời kết bạn cho chính mình")
	}

	// Kiểm tra receiver có tồn tại không
	receiver, err := s.userRepo.FindByID(ctx, receiverID)
	if err != nil {
		return nil, fmt.Errorf("người dùng không tồn tại")
	}
	if !receiver.IsActive {
		return nil, fmt.Errorf("người dùng không hoạt động")
	}

	// Kiểm tra đã là bạn chưa
	isFriend, err := s.friendshipRepo.IsFriend(ctx, senderID, receiverID)
	if err != nil {
		return nil, fmt.Errorf("lỗi kiểm tra quan hệ bạn bè: %w", err)
	}
	if isFriend {
		return nil, fmt.Errorf("đã là bạn bè")
	}

	// Kiểm tra đã có lời mời pending chưa
	existingRequest, err := s.friendRequestRepo.FindBySenderAndReceiver(ctx, senderID, receiverID)
	if err == nil && existingRequest != nil {
		if existingRequest.Status == model.FriendRequestStatusPending {
			return nil, fmt.Errorf("đã có lời mời kết bạn đang chờ")
		}
	}

	// Tạo friend request mới
	friendRequest := model.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Status:     model.FriendRequestStatusPending,
	}

	if err := s.friendRequestRepo.Create(ctx, &friendRequest); err != nil {
		return nil, fmt.Errorf("lỗi tạo lời mời kết bạn: %w", err)
	}

	// Preload relations
	friendRequest.Sender, _ = s.userRepo.FindByID(ctx, senderID)
	friendRequest.Receiver = receiver

	return &friendRequest, nil
}

// AcceptFriendRequest chấp nhận lời mời kết bạn
func (s *Service) AcceptFriendRequest(ctx context.Context, requestID, receiverID uuid.UUID) error {
	// Lấy friend request
	request, err := s.friendRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("lời mời kết bạn không tồn tại")
	}

	// Kiểm tra receiver có phải người nhận không
	if request.ReceiverID != receiverID {
		return fmt.Errorf("bạn không có quyền chấp nhận lời mời này")
	}

	// Kiểm tra status
	if request.Status != model.FriendRequestStatusPending {
		return fmt.Errorf("lời mời kết bạn không ở trạng thái pending")
	}

	// Transaction: cập nhật status và tạo friendship
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Cập nhật status thành accepted
		request.Status = model.FriendRequestStatusAccepted
		if err := tx.WithContext(ctx).Save(request).Error; err != nil {
			return fmt.Errorf("lỗi cập nhật status: %w", err)
		}

		// Tạo friendship (đảm bảo user_id < friend_id để tránh duplicate)
		userID := request.SenderID
		friendID := request.ReceiverID
		if userID.String() > friendID.String() {
			userID, friendID = friendID, userID
		}

		friendship := model.Friendship{
			UserID:   userID,
			FriendID: friendID,
		}

		if err := tx.WithContext(ctx).Create(&friendship).Error; err != nil {
			return fmt.Errorf("lỗi tạo quan hệ bạn bè: %w", err)
		}

		return nil
	})
}

// RejectFriendRequest từ chối lời mời kết bạn
func (s *Service) RejectFriendRequest(ctx context.Context, requestID, receiverID uuid.UUID) error {
	// Lấy friend request
	request, err := s.friendRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("lời mời kết bạn không tồn tại")
	}

	// Kiểm tra receiver có phải người nhận không
	if request.ReceiverID != receiverID {
		return fmt.Errorf("bạn không có quyền từ chối lời mời này")
	}

	// Kiểm tra status
	if request.Status != model.FriendRequestStatusPending {
		return fmt.Errorf("lời mời kết bạn không ở trạng thái pending")
	}

	// Cập nhật status thành rejected
	request.Status = model.FriendRequestStatusRejected
	if err := s.friendRequestRepo.Update(ctx, requestID, request); err != nil {
		return fmt.Errorf("lỗi cập nhật status: %w", err)
	}

	return nil
}

// CancelFriendRequest hủy lời mời kết bạn
func (s *Service) CancelFriendRequest(ctx context.Context, requestID, senderID uuid.UUID) error {
	// Lấy friend request
	request, err := s.friendRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("lời mời kết bạn không tồn tại")
	}

	// Kiểm tra sender có phải người gửi không
	if request.SenderID != senderID {
		return fmt.Errorf("bạn không có quyền hủy lời mời này")
	}

	// Kiểm tra status
	if request.Status != model.FriendRequestStatusPending {
		return fmt.Errorf("chỉ có thể hủy lời mời đang pending")
	}

	// Cập nhật status thành cancelled
	request.Status = model.FriendRequestStatusCancelled
	if err := s.friendRequestRepo.Update(ctx, requestID, request); err != nil {
		return fmt.Errorf("lỗi cập nhật status: %w", err)
	}

	return nil
}

// GetFriendsList lấy danh sách bạn bè
func (s *Service) GetFriendsList(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	// Lấy tất cả friendships
	friendships, err := s.friendshipRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy danh sách bạn bè: %w", err)
	}

	// Lấy thông tin user của từng bạn
	friends := make([]model.User, 0, len(friendships))
	for _, friendship := range friendships {
		var friendID uuid.UUID
		if friendship.UserID == userID {
			friendID = friendship.FriendID
		} else {
			friendID = friendship.UserID
		}

		friend, err := s.userRepo.FindByID(ctx, friendID)
		if err != nil {
			continue // Bỏ qua nếu không tìm thấy
		}
		friends = append(friends, *friend)
	}

	return friends, nil
}

// GetPendingRequests lấy danh sách lời mời đang chờ (nhận được)
func (s *Service) GetPendingRequests(ctx context.Context, userID uuid.UUID) ([]model.FriendRequest, error) {
	requests, err := s.friendRequestRepo.FindPendingByReceiver(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy danh sách lời mời: %w", err)
	}

	// Preload sender
	for i := range requests {
		requests[i].Sender, _ = s.userRepo.FindByID(ctx, requests[i].SenderID)
	}

	return requests, nil
}

// GetSentRequests lấy danh sách lời mời đã gửi
func (s *Service) GetSentRequests(ctx context.Context, userID uuid.UUID) ([]model.FriendRequest, error) {
	requests, err := s.friendRequestRepo.FindPendingBySender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy danh sách lời mời đã gửi: %w", err)
	}

	// Preload receiver
	for i := range requests {
		requests[i].Receiver, _ = s.userRepo.FindByID(ctx, requests[i].ReceiverID)
	}

	return requests, nil
}
