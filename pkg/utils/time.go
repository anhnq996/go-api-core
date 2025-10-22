package utils

import (
	"fmt"
	"time"
)

// Now trả về thời gian hiện tại
func Now() time.Time {
	return time.Now()
}

// Today trả về ngày hôm nay (00:00:00)
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Tomorrow trả về ngày mai (00:00:00)
func Tomorrow() time.Time {
	return Today().AddDate(0, 0, 1)
}

// Yesterday trả về ngày hôm qua (00:00:00)
func Yesterday() time.Time {
	return Today().AddDate(0, 0, -1)
}

// StartOfMonth trả về đầu tháng
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth trả về cuối tháng
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Second)
}

// StartOfYear trả về đầu năm
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear trả về cuối năm
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// FormatDateTime format datetime (2006-01-02 15:04:05)
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate format date (2006-01-02)
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatTime format time (15:04:05)
func FormatTime(t time.Time) string {
	return t.Format("15:04:05")
}

// ParseDateTime parse datetime string
func ParseDateTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}

// ParseDate parse date string
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// DiffDays tính số ngày giữa 2 thời điểm
func DiffDays(t1, t2 time.Time) int {
	diff := t2.Sub(t1)
	return int(diff.Hours() / 24)
}

// DiffHours tính số giờ giữa 2 thời điểm
func DiffHours(t1, t2 time.Time) int {
	diff := t2.Sub(t1)
	return int(diff.Hours())
}

// IsToday kiểm tra có phải hôm nay không
func IsToday(t time.Time) bool {
	today := Today()
	return t.Year() == today.Year() && t.Month() == today.Month() && t.Day() == today.Day()
}

// IsPast kiểm tra có phải trong quá khứ không
func IsPast(t time.Time) bool {
	return t.Before(time.Now())
}

// IsFuture kiểm tra có phải trong tương lai không
func IsFuture(t time.Time) bool {
	return t.After(time.Now())
}

// AddDays thêm số ngày
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddMonths thêm số tháng
func AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// AddYears thêm số năm
func AddYears(t time.Time, years int) time.Time {
	return t.AddDate(years, 0, 0)
}

// Age tính tuổi từ ngày sinh
func Age(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	if now.Month() < birthDate.Month() || (now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}

	return age
}

// IsWeekend kiểm tra có phải cuối tuần không
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWeekday kiểm tra có phải ngày trong tuần không
func IsWeekday(t time.Time) bool {
	return !IsWeekend(t)
}

// TimeAgo format thời gian thành "X ago" (1 hour ago, 2 days ago, etc.)
func TimeAgo(t time.Time) string {
	duration := time.Since(t)

	seconds := int(duration.Seconds())
	minutes := int(duration.Minutes())
	hours := int(duration.Hours())
	days := hours / 24
	months := days / 30
	years := days / 365

	switch {
	case seconds < 60:
		return "just now"
	case minutes < 60:
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours < 24:
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case days < 30:
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case months < 12:
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
