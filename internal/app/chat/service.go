package chat

import (
	"context"
	"fmt"
	"time"

	model "api-core/internal/models"
	repository "api-core/internal/repositories"
	"api-core/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service xử lý business logic cho chat
type Service struct {
	conversationRepo            repository.ConversationRepository
	conversationParticipantRepo repository.ConversationParticipantRepository
	messageRepo                 repository.MessageRepository
	friendshipRepo              repository.FriendshipRepository
	userRepo                    repository.UserRepository
	db                          *gorm.DB
}

// NewService tạo chat service mới
func NewService(
	conversationRepo repository.ConversationRepository,
	conversationParticipantRepo repository.ConversationParticipantRepository,
	messageRepo repository.MessageRepository,
	friendshipRepo repository.FriendshipRepository,
	userRepo repository.UserRepository,
	db *gorm.DB,
) *Service {
	return &Service{
		conversationRepo:            conversationRepo,
		conversationParticipantRepo: conversationParticipantRepo,
		messageRepo:                 messageRepo,
		friendshipRepo:              friendshipRepo,
		userRepo:                    userRepo,
		db:                          db,
	}
}

// GetOrCreateDirectConversation lấy hoặc tạo direct conversation giữa 2 user
func (s *Service) GetOrCreateDirectConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (*model.Conversation, error) {
	// Kiểm tra không thể chat với chính mình
	if user1ID == user2ID {
		return nil, fmt.Errorf("không thể chat với chính mình")
	}

	// Kiểm tra 2 user có phải bạn bè không (nếu cần)
	isFriend, err := s.friendshipRepo.IsFriend(ctx, user1ID, user2ID)
	if err != nil {
		return nil, fmt.Errorf("lỗi kiểm tra quan hệ bạn bè: %w", err)
	}
	if !isFriend {
		return nil, fmt.Errorf("chỉ có thể chat với bạn bè")
	}

	// Tìm conversation đã tồn tại
	conversation, err := s.conversationRepo.FindDirectConversation(ctx, user1ID, user2ID)
	if err == nil && conversation != nil {
		// Preload participants
		conversation, err = s.conversationRepo.FindByIDWithParticipants(ctx, conversation.ID)
		if err != nil {
			return nil, fmt.Errorf("lỗi lấy conversation: %w", err)
		}
		return conversation, nil
	}

	// Tạo conversation mới
	return s.createDirectConversation(ctx, user1ID, user2ID)
}

// createDirectConversation tạo direct conversation mới
func (s *Service) createDirectConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (*model.Conversation, error) {
	var conversation *model.Conversation

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Tạo conversation
		conv := model.Conversation{
			Type: model.ConversationTypeDirect,
		}
		if err := tx.WithContext(ctx).Create(&conv).Error; err != nil {
			return fmt.Errorf("lỗi tạo conversation: %w", err)
		}

		// Tạo participants
		participants := []model.ConversationParticipant{
			{
				ConversationID: conv.ID,
				UserID:         user1ID,
			},
			{
				ConversationID: conv.ID,
				UserID:         user2ID,
			},
		}

		for _, p := range participants {
			if err := tx.WithContext(ctx).Create(&p).Error; err != nil {
				return fmt.Errorf("lỗi tạo participant: %w", err)
			}
		}

		conversation = &conv
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Preload participants
	conversation, err = s.conversationRepo.FindByIDWithParticipants(ctx, conversation.ID)
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy conversation: %w", err)
	}

	return conversation, nil
}

// SendMessage gửi tin nhắn
func (s *Service) SendMessage(ctx context.Context, conversationID, senderID uuid.UUID, content string, messageType model.MessageType, replyToID *uuid.UUID) (*model.Message, error) {
	// Kiểm tra conversation có tồn tại không
	conversation, err := s.conversationRepo.FindByIDWithParticipants(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation không tồn tại")
	}

	// Kiểm tra sender có tham gia conversation không
	isParticipant := false
	for _, p := range conversation.Participants {
		if p.UserID == senderID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, fmt.Errorf("bạn không tham gia conversation này")
	}

	// Kiểm tra reply_to message có tồn tại không
	if replyToID != nil {
		replyTo, err := s.messageRepo.FindByID(ctx, *replyToID)
		if err != nil {
			return nil, fmt.Errorf("tin nhắn được trả lời không tồn tại")
		}
		if replyTo.ConversationID != conversationID {
			return nil, fmt.Errorf("tin nhắn được trả lời không thuộc conversation này")
		}
	}

	// Tạo message
	message := model.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		MessageType:    messageType,
		ReplyToID:      replyToID,
	}

	if err := s.messageRepo.Create(ctx, &message); err != nil {
		return nil, fmt.Errorf("lỗi gửi tin nhắn: %w", err)
	}

	// Preload relations
	message.Sender, _ = s.userRepo.FindByID(ctx, senderID)
	if replyToID != nil {
		message.ReplyTo, _ = s.messageRepo.FindByID(ctx, *replyToID)
	}

	// Cập nhật updated_at của conversation
	now := time.Now()
	s.db.WithContext(ctx).Model(&model.Conversation{}).
		Where("id = ?", conversationID).
		Update("updated_at", now)

	return &message, nil
}

// GetMessages lấy danh sách tin nhắn của conversation
func (s *Service) GetMessages(ctx context.Context, conversationID, userID uuid.UUID, page, perPage int) ([]model.Message, *utils.Pagination, error) {
	// Kiểm tra user có tham gia conversation không
	_, err := s.conversationParticipantRepo.FindByConversationAndUser(ctx, conversationID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("bạn không tham gia conversation này")
	}

	// Lấy messages
	messages, total, err := s.messageRepo.FindByConversationID(ctx, conversationID, page, perPage)
	if err != nil {
		return nil, nil, fmt.Errorf("lỗi lấy tin nhắn: %w", err)
	}

	// Preload sender và reply_to
	for i := range messages {
		messages[i].Sender, _ = s.userRepo.FindByID(ctx, messages[i].SenderID)
		if messages[i].ReplyToID != nil {
			messages[i].ReplyTo, _ = s.messageRepo.FindByID(ctx, *messages[i].ReplyToID)
		}
	}

	// Tạo pagination
	pagination := utils.NewPagination(page, perPage, total)

	// Cập nhật last_read_at
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.conversationParticipantRepo.UpdateLastReadAt(ctx, conversationID, userID)
	}()

	return messages, pagination, nil
}

// GetConversations lấy danh sách conversations của user
func (s *Service) GetConversations(ctx context.Context, userID uuid.UUID) ([]model.Conversation, error) {
	conversations, err := s.conversationRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy danh sách conversations: %w", err)
	}

	return conversations, nil
}
