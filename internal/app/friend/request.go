package friend

// SendFriendRequestRequest request cho gửi lời mời kết bạn
type SendFriendRequestRequest struct {
	ReceiverID string `json:"receiver_id" validate:"required,uuid"`
}

// AcceptFriendRequestRequest request cho chấp nhận lời mời
type AcceptFriendRequestRequest struct {
	RequestID string `json:"request_id" validate:"required,uuid"`
}

// RejectFriendRequestRequest request cho từ chối lời mời
type RejectFriendRequestRequest struct {
	RequestID string `json:"request_id" validate:"required,uuid"`
}

// CancelFriendRequestRequest request cho hủy lời mời
type CancelFriendRequestRequest struct {
	RequestID string `json:"request_id" validate:"required,uuid"`
}
