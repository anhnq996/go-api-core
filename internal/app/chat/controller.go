package chat

import (
	"net/http"

	model "api-core/internal/models"
	"api-core/pkg/i18n"
	"api-core/pkg/jwt"
	"api-core/pkg/response"
	"api-core/pkg/utils"
	"api-core/pkg/validator"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler chứa service của chat
type Handler struct {
	service *Service
}

// NewHandler tạo handler mới
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// GetOrCreateConversation - POST /chats/conversations
func (h *Handler) GetOrCreateConversation(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	userID := jwt.GetUserIDFromContext(r.Context())
	if userID == "" {
		response.Unauthorized(w, lang, response.CodeUnauthorized)
		return
	}

	user1ID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	var input GetOrCreateConversationRequest
	if !validator.ValidateAndRespond(w, r, &input) {
		return
	}

	user2ID, err := uuid.Parse(input.UserID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	conversation, err := h.service.GetOrCreateDirectConversation(r.Context(), user1ID, user2ID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(w, lang, response.CodeSuccess, conversation)
}

// SendMessage - POST /chats/messages
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	userID := jwt.GetUserIDFromContext(r.Context())
	if userID == "" {
		response.Unauthorized(w, lang, response.CodeUnauthorized)
		return
	}

	senderID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	var input SendMessageRequest
	if !validator.ValidateAndRespond(w, r, &input) {
		return
	}

	conversationID, err := uuid.Parse(input.ConversationID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	// Parse message type
	messageType := model.MessageTypeText
	if input.MessageType != "" {
		messageType = model.MessageType(input.MessageType)
	}

	// Parse reply_to_id
	var replyToID *uuid.UUID
	if input.ReplyToID != nil && *input.ReplyToID != "" {
		id, err := uuid.Parse(*input.ReplyToID)
		if err != nil {
			response.BadRequest(w, lang, response.CodeBadRequest, nil)
			return
		}
		replyToID = &id
	}

	message, err := h.service.SendMessage(r.Context(), conversationID, senderID, input.Content, messageType, replyToID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, err.Error())
		return
	}

	response.Created(w, lang, response.CodeCreated, message)
}

// GetMessages - GET /chats/conversations/{id}/messages
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	userID := jwt.GetUserIDFromContext(r.Context())
	if userID == "" {
		response.Unauthorized(w, lang, response.CodeUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	conversationIDStr := chi.URLParam(r, "id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	// Parse query parameters
	params := utils.ParseQueryParams(r)
	page := params.Page
	if page < 1 {
		page = 1
	}
	perPage := params.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	messages, pagination, err := h.service.GetMessages(r.Context(), conversationID, userUUID, page, perPage)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, err.Error())
		return
	}

	responseData := utils.PaginatedResponse(messages, pagination)
	response.Success(w, lang, response.CodeSuccess, responseData)
}

// GetConversations - GET /chats/conversations
func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	userID := jwt.GetUserIDFromContext(r.Context())
	if userID == "" {
		response.Unauthorized(w, lang, response.CodeUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	conversations, err := h.service.GetConversations(r.Context(), userUUID)
	if err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeSuccess, conversations)
}
