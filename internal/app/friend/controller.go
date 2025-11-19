package friend

import (
	"net/http"

	"api-core/pkg/i18n"
	"api-core/pkg/jwt"
	"api-core/pkg/response"
	"api-core/pkg/validator"

	"github.com/google/uuid"
)

// Handler chứa service của friend
type Handler struct {
	service *Service
}

// NewHandler tạo handler mới
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// SendFriendRequest - POST /friends/requests
func (h *Handler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
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

	var input SendFriendRequestRequest
	if !validator.ValidateAndRespond(w, r, &input) {
		return
	}

	receiverID, err := uuid.Parse(input.ReceiverID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	friendRequest, err := h.service.SendFriendRequest(r.Context(), senderID, receiverID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, err.Error())
		return
	}

	response.Created(w, lang, response.CodeCreated, friendRequest)
}

// AcceptFriendRequest - POST /friends/requests/accept
func (h *Handler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	userID := jwt.GetUserIDFromContext(r.Context())
	if userID == "" {
		response.Unauthorized(w, lang, response.CodeUnauthorized)
		return
	}

	receiverID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	var input AcceptFriendRequestRequest
	if !validator.ValidateAndRespond(w, r, &input) {
		return
	}

	requestID, err := uuid.Parse(input.RequestID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	if err := h.service.AcceptFriendRequest(r.Context(), requestID, receiverID); err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(w, lang, response.CodeSuccess, map[string]string{
		"message": "Đã chấp nhận lời mời kết bạn",
	})
}

// RejectFriendRequest - POST /friends/requests/reject
func (h *Handler) RejectFriendRequest(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	userID := jwt.GetUserIDFromContext(r.Context())
	if userID == "" {
		response.Unauthorized(w, lang, response.CodeUnauthorized)
		return
	}

	receiverID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	var input RejectFriendRequestRequest
	if !validator.ValidateAndRespond(w, r, &input) {
		return
	}

	requestID, err := uuid.Parse(input.RequestID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	if err := h.service.RejectFriendRequest(r.Context(), requestID, receiverID); err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(w, lang, response.CodeSuccess, map[string]string{
		"message": "Đã từ chối lời mời kết bạn",
	})
}

// CancelFriendRequest - POST /friends/requests/cancel
func (h *Handler) CancelFriendRequest(w http.ResponseWriter, r *http.Request) {
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

	var input CancelFriendRequestRequest
	if !validator.ValidateAndRespond(w, r, &input) {
		return
	}

	requestID, err := uuid.Parse(input.RequestID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return
	}

	if err := h.service.CancelFriendRequest(r.Context(), requestID, senderID); err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(w, lang, response.CodeSuccess, map[string]string{
		"message": "Đã hủy lời mời kết bạn",
	})
}

// GetFriendsList - GET /friends
func (h *Handler) GetFriendsList(w http.ResponseWriter, r *http.Request) {
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

	friends, err := h.service.GetFriendsList(r.Context(), userUUID)
	if err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeSuccess, friends)
}

// GetPendingRequests - GET /friends/requests/pending
func (h *Handler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
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

	requests, err := h.service.GetPendingRequests(r.Context(), userUUID)
	if err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeSuccess, requests)
}

// GetSentRequests - GET /friends/requests/sent
func (h *Handler) GetSentRequests(w http.ResponseWriter, r *http.Request) {
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

	requests, err := h.service.GetSentRequests(r.Context(), userUUID)
	if err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeSuccess, requests)
}
