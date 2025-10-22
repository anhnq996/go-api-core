package fcm

import (
	"time"

	"firebase.google.com/go/v4/messaging"
)

// Notification định nghĩa cấu trúc notification cơ bản
type Notification struct {
	Title    string                   // Tiêu đề notification
	Body     string                   // Nội dung notification
	ImageURL string                   // URL hình ảnh (optional)
	Android  *messaging.AndroidConfig // Cấu hình riêng cho Android
	APNS     *messaging.APNSConfig    // Cấu hình riêng cho iOS
	Webpush  *messaging.WebpushConfig // Cấu hình riêng cho Web
}

// NotificationBuilder giúp xây dựng notification một cách dễ dàng
type NotificationBuilder struct {
	notification *Notification
}

// NewNotificationBuilder tạo builder mới
func NewNotificationBuilder() *NotificationBuilder {
	return &NotificationBuilder{
		notification: &Notification{},
	}
}

// SetTitle đặt tiêu đề
func (b *NotificationBuilder) SetTitle(title string) *NotificationBuilder {
	b.notification.Title = title
	return b
}

// SetBody đặt nội dung
func (b *NotificationBuilder) SetBody(body string) *NotificationBuilder {
	b.notification.Body = body
	return b
}

// SetImageURL đặt URL hình ảnh
func (b *NotificationBuilder) SetImageURL(imageURL string) *NotificationBuilder {
	b.notification.ImageURL = imageURL
	return b
}

// SetAndroidConfig đặt cấu hình Android
func (b *NotificationBuilder) SetAndroidConfig(config *messaging.AndroidConfig) *NotificationBuilder {
	b.notification.Android = config
	return b
}

// SetAPNSConfig đặt cấu hình iOS
func (b *NotificationBuilder) SetAPNSConfig(config *messaging.APNSConfig) *NotificationBuilder {
	b.notification.APNS = config
	return b
}

// SetWebpushConfig đặt cấu hình Web
func (b *NotificationBuilder) SetWebpushConfig(config *messaging.WebpushConfig) *NotificationBuilder {
	b.notification.Webpush = config
	return b
}

// Build xây dựng notification
func (b *NotificationBuilder) Build() *Notification {
	return b.notification
}

// WithAndroidPriority helper để set priority cho Android
func (b *NotificationBuilder) WithAndroidPriority(priority string) *NotificationBuilder {
	if b.notification.Android == nil {
		b.notification.Android = &messaging.AndroidConfig{}
	}
	b.notification.Android.Priority = priority
	return b
}

// WithAndroidTTL helper để set TTL cho Android (ví dụ: 3600 cho 1 giờ)
func (b *NotificationBuilder) WithAndroidTTL(ttlSeconds int) *NotificationBuilder {
	if b.notification.Android == nil {
		b.notification.Android = &messaging.AndroidConfig{}
	}
	ttl := time.Duration(ttlSeconds) * time.Second
	b.notification.Android.TTL = &ttl
	return b
}

// WithAndroidSound helper để set sound cho Android
func (b *NotificationBuilder) WithAndroidSound(sound string) *NotificationBuilder {
	if b.notification.Android == nil {
		b.notification.Android = &messaging.AndroidConfig{}
	}
	if b.notification.Android.Notification == nil {
		b.notification.Android.Notification = &messaging.AndroidNotification{}
	}
	b.notification.Android.Notification.Sound = sound
	return b
}

// WithAndroidColor helper để set màu cho Android notification
func (b *NotificationBuilder) WithAndroidColor(color string) *NotificationBuilder {
	if b.notification.Android == nil {
		b.notification.Android = &messaging.AndroidConfig{}
	}
	if b.notification.Android.Notification == nil {
		b.notification.Android.Notification = &messaging.AndroidNotification{}
	}
	b.notification.Android.Notification.Color = color
	return b
}

// WithAndroidIcon helper để set icon cho Android
func (b *NotificationBuilder) WithAndroidIcon(icon string) *NotificationBuilder {
	if b.notification.Android == nil {
		b.notification.Android = &messaging.AndroidConfig{}
	}
	if b.notification.Android.Notification == nil {
		b.notification.Android.Notification = &messaging.AndroidNotification{}
	}
	b.notification.Android.Notification.Icon = icon
	return b
}

// WithAndroidClickAction helper để set click action cho Android
func (b *NotificationBuilder) WithAndroidClickAction(clickAction string) *NotificationBuilder {
	if b.notification.Android == nil {
		b.notification.Android = &messaging.AndroidConfig{}
	}
	if b.notification.Android.Notification == nil {
		b.notification.Android.Notification = &messaging.AndroidNotification{}
	}
	b.notification.Android.Notification.ClickAction = clickAction
	return b
}

// WithIOSBadge helper để set badge cho iOS
func (b *NotificationBuilder) WithIOSBadge(badge int) *NotificationBuilder {
	if b.notification.APNS == nil {
		b.notification.APNS = &messaging.APNSConfig{}
	}
	if b.notification.APNS.Payload == nil {
		b.notification.APNS.Payload = &messaging.APNSPayload{}
	}
	if b.notification.APNS.Payload.Aps == nil {
		b.notification.APNS.Payload.Aps = &messaging.Aps{}
	}
	badgePtr := &badge
	b.notification.APNS.Payload.Aps.Badge = badgePtr
	return b
}

// WithIOSSound helper để set sound cho iOS
func (b *NotificationBuilder) WithIOSSound(sound string) *NotificationBuilder {
	if b.notification.APNS == nil {
		b.notification.APNS = &messaging.APNSConfig{}
	}
	if b.notification.APNS.Payload == nil {
		b.notification.APNS.Payload = &messaging.APNSPayload{}
	}
	if b.notification.APNS.Payload.Aps == nil {
		b.notification.APNS.Payload.Aps = &messaging.Aps{}
	}
	b.notification.APNS.Payload.Aps.Sound = sound
	return b
}

// WithIOSCategory helper để set category cho iOS
func (b *NotificationBuilder) WithIOSCategory(category string) *NotificationBuilder {
	if b.notification.APNS == nil {
		b.notification.APNS = &messaging.APNSConfig{}
	}
	if b.notification.APNS.Payload == nil {
		b.notification.APNS.Payload = &messaging.APNSPayload{}
	}
	if b.notification.APNS.Payload.Aps == nil {
		b.notification.APNS.Payload.Aps = &messaging.Aps{}
	}
	b.notification.APNS.Payload.Aps.Category = category
	return b
}

// WithIOSThreadID helper để set thread ID cho iOS
func (b *NotificationBuilder) WithIOSThreadID(threadID string) *NotificationBuilder {
	if b.notification.APNS == nil {
		b.notification.APNS = &messaging.APNSConfig{}
	}
	if b.notification.APNS.Payload == nil {
		b.notification.APNS.Payload = &messaging.APNSPayload{}
	}
	if b.notification.APNS.Payload.Aps == nil {
		b.notification.APNS.Payload.Aps = &messaging.Aps{}
	}
	b.notification.APNS.Payload.Aps.ThreadID = threadID
	return b
}

// WithIOSContentAvailable helper để set content-available cho iOS
func (b *NotificationBuilder) WithIOSContentAvailable() *NotificationBuilder {
	if b.notification.APNS == nil {
		b.notification.APNS = &messaging.APNSConfig{}
	}
	if b.notification.APNS.Payload == nil {
		b.notification.APNS.Payload = &messaging.APNSPayload{}
	}
	if b.notification.APNS.Payload.Aps == nil {
		b.notification.APNS.Payload.Aps = &messaging.Aps{}
	}
	b.notification.APNS.Payload.Aps.ContentAvailable = true
	return b
}

// WithIOSMutableContent helper để set mutable-content cho iOS
func (b *NotificationBuilder) WithIOSMutableContent() *NotificationBuilder {
	if b.notification.APNS == nil {
		b.notification.APNS = &messaging.APNSConfig{}
	}
	if b.notification.APNS.Payload == nil {
		b.notification.APNS.Payload = &messaging.APNSPayload{}
	}
	if b.notification.APNS.Payload.Aps == nil {
		b.notification.APNS.Payload.Aps = &messaging.Aps{}
	}
	b.notification.APNS.Payload.Aps.MutableContent = true
	return b
}
