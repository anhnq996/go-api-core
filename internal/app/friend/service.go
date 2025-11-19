package friend

import (
	"context"

	model "api-core/internal/models"
	repository "api-core/internal/repositories"
	"api-core/pkg/i18n"
	"api-core/pkg/response"

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
func (s *Service) SendFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Kiểm tra không thể tự gửi lời mời cho chính mình
	if senderID == receiverID {
		return response.BadRequestResponse(lang, response.CodeCannotSendRequestToSelf, nil)
	}

	// Kiểm tra receiver có tồn tại không
	receiver, err := s.userRepo.FindByID(ctx, receiverID)
	if err != nil {
		return response.NotFoundResponse(lang, response.CodeUserNotFound)
	}
	if !receiver.IsActive {
		return response.ForbiddenResponse(lang, response.CodeUserInactive)
	}

	// Kiểm tra đã là bạn chưa
	isFriend, err := s.friendshipRepo.IsFriend(ctx, senderID, receiverID)
	if err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeCheckFriendshipFailed)
	}
	if isFriend {
		return response.ConflictResponse(lang, response.CodeAlreadyFriends)
	}

	// Kiểm tra đã có lời mời pending chưa
	existingRequest, err := s.friendRequestRepo.FindBySenderAndReceiver(ctx, senderID, receiverID)
	if err == nil && existingRequest != nil {
		if existingRequest.Status == model.FriendRequestStatusPending {
			return response.ConflictResponse(lang, response.CodeFriendRequestPending)
		}
	}

	// Tạo friend request mới
	friendRequest := model.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Status:     model.FriendRequestStatusPending,
	}

	if err := s.friendRequestRepo.Create(ctx, &friendRequest); err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeSendFriendRequestFailed)
	}

	// Preload relations
	friendRequest.Sender, _ = s.userRepo.FindByID(ctx, senderID)
	friendRequest.Receiver = receiver

	return response.SuccessResponse(lang, response.CodeCreated, friendRequest)
}

// AcceptFriendRequest chấp nhận lời mời kết bạn
func (s *Service) AcceptFriendRequest(ctx context.Context, requestID, receiverID uuid.UUID) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Lấy friend request
	request, err := s.friendRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return response.NotFoundResponse(lang, response.CodeFriendRequestNotFound)
	}

	// Kiểm tra receiver có phải người nhận không
	if request.ReceiverID != receiverID {
		return response.ForbiddenResponse(lang, response.CodeNotRequestReceiver)
	}

	// Kiểm tra status
	if request.Status != model.FriendRequestStatusPending {
		return response.BadRequestResponse(lang, response.CodeFriendRequestNotPending, nil)
	}

	// Transaction: cập nhật status và tạo friendship
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Cập nhật status thành accepted
		request.Status = model.FriendRequestStatusAccepted
		if err := tx.WithContext(ctx).Save(request).Error; err != nil {
			return err
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
			return err
		}

		return nil
	})

	if err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeAcceptFriendRequestFailed)
	}

	return response.SuccessResponse(lang, response.CodeSuccess, map[string]string{
		"message": "Đã chấp nhận lời mời kết bạn",
	})
}

// RejectFriendRequest từ chối lời mời kết bạn
func (s *Service) RejectFriendRequest(ctx context.Context, requestID, receiverID uuid.UUID) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Lấy friend request
	request, err := s.friendRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return response.NotFoundResponse(lang, response.CodeFriendRequestNotFound)
	}

	// Kiểm tra receiver có phải người nhận không
	if request.ReceiverID != receiverID {
		return response.ForbiddenResponse(lang, response.CodeNotRequestReceiver)
	}

	// Kiểm tra status
	if request.Status != model.FriendRequestStatusPending {
		return response.BadRequestResponse(lang, response.CodeFriendRequestNotPending, nil)
	}

	// Cập nhật status thành rejected
	request.Status = model.FriendRequestStatusRejected
	if err := s.friendRequestRepo.Update(ctx, requestID, request); err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeRejectFriendRequestFailed)
	}

	return response.SuccessResponse(lang, response.CodeSuccess, map[string]string{
		"message": "Đã từ chối lời mời kết bạn",
	})
}

// CancelFriendRequest hủy lời mời kết bạn
func (s *Service) CancelFriendRequest(ctx context.Context, requestID, senderID uuid.UUID) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Lấy friend request
	request, err := s.friendRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return response.NotFoundResponse(lang, response.CodeFriendRequestNotFound)
	}

	// Kiểm tra sender có phải người gửi không
	if request.SenderID != senderID {
		return response.ForbiddenResponse(lang, response.CodeNotRequestSender)
	}

	// Kiểm tra status
	if request.Status != model.FriendRequestStatusPending {
		return response.BadRequestResponse(lang, response.CodeCannotCancelNonPendingRequest, nil)
	}

	// Cập nhật status thành cancelled
	request.Status = model.FriendRequestStatusCancelled
	if err := s.friendRequestRepo.Update(ctx, requestID, request); err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeCancelFriendRequestFailed)
	}

	return response.SuccessResponse(lang, response.CodeSuccess, map[string]string{
		"message": "Đã hủy lời mời kết bạn",
	})
}

// GetFriendsList lấy danh sách bạn bè
func (s *Service) GetFriendsList(ctx context.Context, userID uuid.UUID) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Lấy tất cả friendships
	friendships, err := s.friendshipRepo.FindByUserID(ctx, userID)
	if err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeGetFriendsListFailed)
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

	return response.SuccessResponse(lang, response.CodeSuccess, friends)
}

// GetPendingRequests lấy danh sách lời mời đang chờ (nhận được)
func (s *Service) GetPendingRequests(ctx context.Context, userID uuid.UUID) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	requests, err := s.friendRequestRepo.FindPendingByReceiver(ctx, userID)
	if err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeGetPendingRequestsFailed)
	}

	// Preload sender
	for i := range requests {
		requests[i].Sender, _ = s.userRepo.FindByID(ctx, requests[i].SenderID)
	}

	return response.SuccessResponse(lang, response.CodeSuccess, requests)
}

// GetSentRequests lấy danh sách lời mời đã gửi
func (s *Service) GetSentRequests(ctx context.Context, userID uuid.UUID) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	requests, err := s.friendRequestRepo.FindPendingBySender(ctx, userID)
	if err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeGetSentRequestsFailed)
	}

	// Preload receiver
	for i := range requests {
		requests[i].Receiver, _ = s.userRepo.FindByID(ctx, requests[i].ReceiverID)
	}

	return response.SuccessResponse(lang, response.CodeSuccess, requests)
}
