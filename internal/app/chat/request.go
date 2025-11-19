package chat

// SendMessageRequest request cho gửi tin nhắn
type SendMessageRequest struct {
	ConversationID string  `json:"conversation_id" validate:"required,uuid"`
	Content        string  `json:"content" validate:"required,min=1,max=5000"`
	MessageType    string  `json:"message_type" validate:"omitempty,oneof=text image file audio video location system"`
	ReplyToID      *string `json:"reply_to_id" validate:"omitempty,uuid"`
}

// GetMessagesRequest request cho lấy tin nhắn
type GetMessagesRequest struct {
	Page    int `json:"page" validate:"omitempty,min=1"`
	PerPage int `json:"per_page" validate:"omitempty,min=1,max=100"`
}

// GetOrCreateConversationRequest request cho lấy/tạo conversation
type GetOrCreateConversationRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}
