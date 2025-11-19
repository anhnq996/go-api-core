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

	resp := h.service.SendFriendRequest(r.Context(), senderID, receiverID)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
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

	resp := h.service.AcceptFriendRequest(r.Context(), requestID, receiverID)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
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

	resp := h.service.RejectFriendRequest(r.Context(), requestID, receiverID)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
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

	resp := h.service.CancelFriendRequest(r.Context(), requestID, senderID)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
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

	resp := h.service.GetFriendsList(r.Context(), userUUID)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
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

	resp := h.service.GetPendingRequests(r.Context(), userUUID)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
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

	resp := h.service.GetSentRequests(r.Context(), userUUID)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
}
