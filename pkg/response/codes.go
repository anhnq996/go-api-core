package response

// Response codes
// Các mã này sẽ được dịch sang message tương ứng theo ngôn ngữ

const (
	// Success codes (2xx)
	CodeSuccess   = "SUCCESS"
	CodeCreated   = "CREATED"
	CodeUpdated   = "UPDATED"
	CodeDeleted   = "DELETED"
	CodeNoContent = "NO_CONTENT"

	// Client errors (4xx)
	CodeBadRequest       = "BAD_REQUEST"
	CodeInvalidInput     = "INVALID_INPUT"
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeResourceNotFound = "RESOURCE_NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeDuplicateEntry   = "DUPLICATE_ENTRY"
	CodeTooManyRequests  = "TOO_MANY_REQUESTS"

	// Authentication & Authorization
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeTokenInvalid       = "TOKEN_INVALID"
	CodeTokenMissing       = "TOKEN_MISSING"
	CodePermissionDenied   = "PERMISSION_DENIED"
	CodeAccountDisabled    = "ACCOUNT_DISABLED"
	CodeAccountNotVerified = "ACCOUNT_NOT_VERIFIED"

	// Server errors (5xx)
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeDatabaseError       = "DATABASE_ERROR"
	CodeCacheError          = "CACHE_ERROR"

	// Business logic errors
	CodeInsufficientBalance = "INSUFFICIENT_BALANCE"
	CodeOperationFailed     = "OPERATION_FAILED"
	CodeInvalidOperation    = "INVALID_OPERATION"
	CodeLimitExceeded       = "LIMIT_EXCEEDED"

	// File & Upload errors
	CodeFileUploadFailed = "FILE_UPLOAD_FAILED"
	CodeFileNotFound     = "FILE_NOT_FOUND"
	CodeFileTooLarge     = "FILE_TOO_LARGE"
	CodeInvalidFileType  = "INVALID_FILE_TYPE"

	// User specific
	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeUserAlreadyExists  = "USER_ALREADY_EXISTS"
	CodeEmailAlreadyExists = "EMAIL_ALREADY_EXISTS"
	CodePhoneAlreadyExists = "PHONE_ALREADY_EXISTS"

	// Pagination
	CodeInvalidPage     = "INVALID_PAGE"
	CodeInvalidPageSize = "INVALID_PAGE_SIZE"

	// Authentication Success
	CodeLoginSuccess   = "LOGIN_SUCCESS"
	CodeLogoutSuccess  = "LOGOUT_SUCCESS"
	CodeTokenRefreshed = "TOKEN_REFRESHED"

	// Rate limit
	CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"

	// Friend errors
	CodeCannotSendRequestToSelf       = "CANNOT_SEND_REQUEST_TO_SELF"
	CodeUserInactive                  = "USER_INACTIVE"
	CodeAlreadyFriends                = "ALREADY_FRIENDS"
	CodeFriendRequestPending          = "FRIEND_REQUEST_PENDING"
	CodeFriendRequestNotFound         = "FRIEND_REQUEST_NOT_FOUND"
	CodeNotRequestReceiver            = "NOT_REQUEST_RECEIVER"
	CodeNotRequestSender              = "NOT_REQUEST_SENDER"
	CodeFriendRequestNotPending       = "FRIEND_REQUEST_NOT_PENDING"
	CodeCannotCancelNonPendingRequest = "CANNOT_CANCEL_NON_PENDING_REQUEST"
	CodeSendFriendRequestFailed       = "SEND_FRIEND_REQUEST_FAILED"
	CodeAcceptFriendRequestFailed     = "ACCEPT_FRIEND_REQUEST_FAILED"
	CodeRejectFriendRequestFailed     = "REJECT_FRIEND_REQUEST_FAILED"
	CodeCancelFriendRequestFailed     = "CANCEL_FRIEND_REQUEST_FAILED"
	CodeGetFriendsListFailed          = "GET_FRIENDS_LIST_FAILED"
	CodeGetPendingRequestsFailed      = "GET_PENDING_REQUESTS_FAILED"
	CodeGetSentRequestsFailed         = "GET_SENT_REQUESTS_FAILED"
	CodeCheckFriendshipFailed         = "CHECK_FRIENDSHIP_FAILED"

	// Chat errors
	CodeCannotChatWithSelf            = "CANNOT_CHAT_WITH_SELF"
	CodeNotFriend                     = "NOT_FRIEND"
	CodeConversationNotFound          = "CONVERSATION_NOT_FOUND"
	CodeNotParticipant                = "NOT_PARTICIPANT"
	CodeMessageNotFound               = "MESSAGE_NOT_FOUND"
	CodeReplyMessageNotInConversation = "REPLY_MESSAGE_NOT_IN_CONVERSATION"
	CodeSendMessageFailed             = "SEND_MESSAGE_FAILED"
	CodeGetMessagesFailed             = "GET_MESSAGES_FAILED"
	CodeGetConversationsFailed        = "GET_CONVERSATIONS_FAILED"
	CodeCreateConversationFailed      = "CREATE_CONVERSATION_FAILED"
	CodeGetConversationFailed         = "GET_CONVERSATION_FAILED"
)

// GetHTTPStatusCode trả về HTTP status code tương ứng với response code
func GetHTTPStatusCode(code string) int {
	statusMap := map[string]int{
		// Success
		CodeSuccess:   200,
		CodeCreated:   201,
		CodeUpdated:   200,
		CodeDeleted:   200,
		CodeNoContent: 204,

		// Client errors
		CodeBadRequest:       400,
		CodeInvalidInput:     400,
		CodeValidationFailed: 422,
		CodeUnauthorized:     401,
		CodeForbidden:        403,
		CodeNotFound:         404,
		CodeResourceNotFound: 404,
		CodeConflict:         409,
		CodeDuplicateEntry:   409,
		CodeTooManyRequests:  429,

		// Auth errors
		CodeInvalidCredentials: 401,
		CodeTokenExpired:       401,
		CodeTokenInvalid:       401,
		CodeTokenMissing:       401,
		CodePermissionDenied:   403,
		CodeAccountDisabled:    403,
		CodeAccountNotVerified: 403,

		// Server errors
		CodeInternalServerError: 500,
		CodeServiceUnavailable:  503,
		CodeDatabaseError:       500,
		CodeCacheError:          500,

		// Business logic
		CodeInsufficientBalance: 400,
		CodeOperationFailed:     400,
		CodeInvalidOperation:    400,
		CodeLimitExceeded:       400,

		// File errors
		CodeFileUploadFailed: 400,
		CodeFileNotFound:     404,
		CodeFileTooLarge:     413,
		CodeInvalidFileType:  400,

		// User errors
		CodeUserNotFound:       404,
		CodeUserAlreadyExists:  409,
		CodeEmailAlreadyExists: 409,
		CodePhoneAlreadyExists: 409,

		// Pagination
		CodeInvalidPage:     400,
		CodeInvalidPageSize: 400,

		// Rate limit
		CodeRateLimitExceeded: 429,

		// Friend errors
		CodeCannotSendRequestToSelf:       400,
		CodeUserInactive:                  403,
		CodeAlreadyFriends:                409,
		CodeFriendRequestPending:          409,
		CodeFriendRequestNotFound:         404,
		CodeNotRequestReceiver:            403,
		CodeNotRequestSender:              403,
		CodeFriendRequestNotPending:       400,
		CodeCannotCancelNonPendingRequest: 400,
		CodeSendFriendRequestFailed:       500,
		CodeAcceptFriendRequestFailed:     500,
		CodeRejectFriendRequestFailed:     500,
		CodeCancelFriendRequestFailed:     500,
		CodeGetFriendsListFailed:          500,
		CodeGetPendingRequestsFailed:      500,
		CodeGetSentRequestsFailed:         500,
		CodeCheckFriendshipFailed:         500,

		// Chat errors
		CodeCannotChatWithSelf:            400,
		CodeNotFriend:                     403,
		CodeConversationNotFound:          404,
		CodeNotParticipant:                403,
		CodeMessageNotFound:               404,
		CodeReplyMessageNotInConversation: 400,
		CodeSendMessageFailed:             500,
		CodeGetMessagesFailed:             500,
		CodeGetConversationsFailed:        500,
		CodeCreateConversationFailed:      500,
		CodeGetConversationFailed:         500,
	}

	if status, ok := statusMap[code]; ok {
		return status
	}

	// Default to 500 for unknown codes
	return 500
}
