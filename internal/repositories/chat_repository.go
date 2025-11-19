package repository

import (
	"context"
	"time"

	model "api-core/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationRepository interface
type ConversationRepository interface {
	Repository[model.Conversation]

	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Conversation, error)
	FindDirectConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (*model.Conversation, error)
	FindByIDWithParticipants(ctx context.Context, id uuid.UUID) (*model.Conversation, error)
}

// ConversationParticipantRepository interface
type ConversationParticipantRepository interface {
	Repository[model.ConversationParticipant]

	FindByConversationID(ctx context.Context, conversationID uuid.UUID) ([]model.ConversationParticipant, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ConversationParticipant, error)
	FindByConversationAndUser(ctx context.Context, conversationID, userID uuid.UUID) (*model.ConversationParticipant, error)
	UpdateLastReadAt(ctx context.Context, conversationID, userID uuid.UUID) error
}

// MessageRepository interface
type MessageRepository interface {
	Repository[model.Message]

	FindByConversationID(ctx context.Context, conversationID uuid.UUID, page, perPage int) ([]model.Message, int64, error)
	FindLatestByConversationID(ctx context.Context, conversationID uuid.UUID, limit int) ([]model.Message, error)
	FindUnreadCount(ctx context.Context, conversationID, userID uuid.UUID) (int64, error)
}

// conversationRepository implementation
type conversationRepository struct {
	*BaseRepository[model.Conversation]
}

// NewConversationRepository tạo conversation repository mới
func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{
		BaseRepository: NewBaseRepository[model.Conversation](db, true),
	}
}

// FindByUserID tìm tất cả conversations của user
func (r *conversationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Conversation, error) {
	var conversations []model.Conversation
	err := r.DB().WithContext(ctx).
		Joins("INNER JOIN conversation_participants ON conversations.id = conversation_participants.conversation_id").
		Where("conversation_participants.user_id = ? AND conversation_participants.deleted_at IS NULL", userID).
		Preload("Participants.User").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		Find(&conversations).Error
	return conversations, err
}

// FindDirectConversation tìm direct conversation giữa 2 user
func (r *conversationRepository) FindDirectConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (*model.Conversation, error) {
	var conversation model.Conversation
	err := r.DB().WithContext(ctx).
		Where("type = ?", model.ConversationTypeDirect).
		Joins("INNER JOIN conversation_participants cp1 ON conversations.id = cp1.conversation_id AND cp1.user_id = ?", user1ID).
		Joins("INNER JOIN conversation_participants cp2 ON conversations.id = cp2.conversation_id AND cp2.user_id = ?", user2ID).
		Where("cp1.deleted_at IS NULL AND cp2.deleted_at IS NULL").
		Preload("Participants.User").
		First(&conversation).Error

	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// FindByIDWithParticipants tìm conversation theo ID kèm participants
func (r *conversationRepository) FindByIDWithParticipants(ctx context.Context, id uuid.UUID) (*model.Conversation, error) {
	var conversation model.Conversation
	err := r.DB().WithContext(ctx).
		Preload("Participants.User").
		Preload("Creator").
		Where("id = ?", id).
		First(&conversation).Error

	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// conversationParticipantRepository implementation
type conversationParticipantRepository struct {
	*BaseRepository[model.ConversationParticipant]
}

// NewConversationParticipantRepository tạo conversation participant repository mới
func NewConversationParticipantRepository(db *gorm.DB) ConversationParticipantRepository {
	return &conversationParticipantRepository{
		BaseRepository: NewBaseRepository[model.ConversationParticipant](db, true),
	}
}

// FindByConversationID tìm tất cả participants của conversation
func (r *conversationParticipantRepository) FindByConversationID(ctx context.Context, conversationID uuid.UUID) ([]model.ConversationParticipant, error) {
	return r.FindWhere(ctx, "conversation_id = ?", conversationID)
}

// FindByUserID tìm tất cả conversations mà user tham gia
func (r *conversationParticipantRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ConversationParticipant, error) {
	return r.FindWhere(ctx, "user_id = ?", userID)
}

// FindByConversationAndUser tìm participant theo conversation và user
func (r *conversationParticipantRepository) FindByConversationAndUser(ctx context.Context, conversationID, userID uuid.UUID) (*model.ConversationParticipant, error) {
	return r.FirstWhere(ctx, "conversation_id = ? AND user_id = ?", conversationID, userID)
}

// UpdateLastReadAt cập nhật thời gian đọc tin nhắn cuối
func (r *conversationParticipantRepository) UpdateLastReadAt(ctx context.Context, conversationID, userID uuid.UUID) error {
	now := time.Now()
	return r.UpdateWhere(ctx, "conversation_id = ? AND user_id = ?", map[string]interface{}{
		"last_read_at": now,
	}, conversationID, userID)
}

// messageRepository implementation
type messageRepository struct {
	*BaseRepository[model.Message]
}

// NewMessageRepository tạo message repository mới
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{
		BaseRepository: NewBaseRepository[model.Message](db, true),
	}
}

// FindByConversationID tìm messages của conversation với pagination
func (r *messageRepository) FindByConversationID(ctx context.Context, conversationID uuid.UUID, page, perPage int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	// Set defaults
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	// Count total
	if err := r.DB().WithContext(ctx).
		Model(&model.Message{}).
		Where("conversation_id = ?", conversationID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get messages with pagination
	offset := (page - 1) * perPage
	err := r.DB().WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Preload("Sender").
		Preload("ReplyTo").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&messages).Error

	// Reverse để có thứ tự từ cũ đến mới
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, total, err
}

// FindLatestByConversationID tìm tin nhắn mới nhất của conversation
func (r *messageRepository) FindLatestByConversationID(ctx context.Context, conversationID uuid.UUID, limit int) ([]model.Message, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var messages []model.Message
	err := r.DB().WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Preload("Sender").
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
}

// FindUnreadCount đếm số tin nhắn chưa đọc
func (r *messageRepository) FindUnreadCount(ctx context.Context, conversationID, userID uuid.UUID) (int64, error) {
	var count int64

	// Lấy last_read_at của user trong conversation
	var participant model.ConversationParticipant
	err := r.DB().WithContext(ctx).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		First(&participant).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Nếu chưa có participant, đếm tất cả messages
			err := r.DB().WithContext(ctx).
				Model(&model.Message{}).
				Where("conversation_id = ?", conversationID).
				Count(&count).Error
			return count, err
		}
		return 0, err
	}

	// Đếm messages sau last_read_at
	query := r.DB().WithContext(ctx).
		Model(&model.Message{}).
		Where("conversation_id = ? AND sender_id != ?", conversationID, userID)

	if participant.LastReadAt != nil {
		query = query.Where("created_at > ?", participant.LastReadAt)
	}

	err = query.Count(&count).Error
	return count, err
}
