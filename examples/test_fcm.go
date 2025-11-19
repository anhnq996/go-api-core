package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"api-core/pkg/fcm"

	"firebase.google.com/go/v4/messaging"
)

var (
	credentialsFile string
	deviceToken     string
	deviceTokens    string
	topic           string
	condition       string
	testType        string
	dryRun          bool
)

func init() {
	flag.StringVar(&credentialsFile, "credentials", "keys/firebase-credentials.json", "Đường dẫn tới file Firebase credentials")
	flag.StringVar(&deviceToken, "token", "cgiBLa_jPg5EMCUWmqHMbD:APA91bHvpxmigWVs9uCEwSG2ib6T6b6w-ygwoqVxkDXAJpqw-dr589YZ2ijcUlqcI_u6JVkfTUrqdGOnBC68s8OG1Lf3_yyzGSvzDvko-ZNyN64rMnjBwKw", "Device token để gửi notification")
	flag.StringVar(&deviceTokens, "tokens", "", "Danh sách device tokens (phân cách bởi dấu phẩy)")
	flag.StringVar(&topic, "topic", "", "Topic để gửi notification")
	flag.StringVar(&condition, "condition", "", "Condition để gửi notification (ví dụ: 'news' in topics)")
	flag.StringVar(&testType, "test", "all", "Loại test: all, token, tokens, topic, condition, subscribe, unsubscribe, dryrun")
	flag.BoolVar(&dryRun, "dryrun", false, "Chạy ở chế độ dry-run (không gửi thực sự)")
}

func main() {
	flag.Parse()

	fmt.Println("🔥 FCM Test Script - ApiCore")
	fmt.Println("==============================")
	fmt.Println()

	// Kiểm tra file credentials
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		log.Fatalf("❌ Không tìm thấy file credentials: %s\n", credentialsFile)
		log.Fatalf("   Vui lòng đặt file Firebase service account vào thư mục keys/\n")
		log.Fatalf("   Xem hướng dẫn trong examples/README_FCM.md\n")
	}

	// Khởi tạo FCM client
	fmt.Printf("📋 Cấu hình:\n")
	fmt.Printf("   Credentials: %s\n", credentialsFile)
	fmt.Printf("   Timeout: 10s\n")
	fmt.Println()

	config := &fcm.Config{
		CredentialsFile: credentialsFile,
		Timeout:         10 * time.Second,
	}

	client, err := fcm.NewClient(config)
	if err != nil {
		log.Fatalf("❌ Không thể khởi tạo FCM client: %v\n", err)
	}

	fmt.Println("✅ FCM client đã được khởi tạo thành công!")
	fmt.Println()

	ctx := context.Background()

	// Chạy các test dựa trên flag
	switch testType {
	case "token":
		testSendToToken(ctx, client)
	case "tokens":
		testSendToTokens(ctx, client)
	case "topic":
		testSendToTopic(ctx, client)
	case "condition":
		testSendToCondition(ctx, client)
	case "subscribe":
		testSubscribeToTopic(ctx, client)
	case "unsubscribe":
		testUnsubscribeFromTopic(ctx, client)
	case "dryrun":
		testDryRun(ctx, client)
	case "all":
		testAllFeatures(ctx, client)
	default:
		log.Fatalf("❌ Loại test không hợp lệ: %s\n", testType)
	}
}

// Test 1: Gửi notification đến một device token
func testSendToToken(ctx context.Context, client *fcm.Client) {
	fmt.Println("🧪 Test 1: Send Notification to Token")
	fmt.Println("-----------------------------------")

	if deviceToken == "" {
		fmt.Println("⚠️  Chưa có device token, sử dụng token mẫu (sẽ thất bại nếu token không hợp lệ)")
		deviceToken = "YOUR_DEVICE_TOKEN_HERE"
	}

	// Notification đơn giản
	notification := fcm.NewNotificationBuilder().
		SetTitle("Test Notification").
		SetBody("Đây là notification test từ ApiCore FCM").
		Build()

	data := map[string]string{
		"type":      "test",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		"test_id":   "test_send_to_token",
		"action":    "open_app",
		"deep_link": "apicore://test",
	}

	fmt.Printf("📤 Gửi notification đến token: %s\n", deviceToken[:20]+"...")
	fmt.Printf("   Title: %s\n", notification.Title)
	fmt.Printf("   Body: %s\n", notification.Body)
	fmt.Printf("   Data: %v\n", data)
	fmt.Println()

	messageID, err := client.SendToToken(ctx, deviceToken, notification, data)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Thành công! Message ID: %s\n", messageID)
}

// Test 2: Gửi notification đến nhiều device tokens (Multicast)
func testSendToTokens(ctx context.Context, client *fcm.Client) {
	fmt.Println("\n🧪 Test 2: Send Notification to Multiple Tokens (Multicast)")
	fmt.Println("-----------------------------------------------------------")

	var tokens []string
	if deviceTokens == "" {
		fmt.Println("⚠️  Chưa có danh sách tokens, sử dụng tokens mẫu (sẽ thất bại nếu tokens không hợp lệ)")
		tokens = []string{
			"YOUR_DEVICE_TOKEN_1",
			"YOUR_DEVICE_TOKEN_2",
		}
	} else {
		tokens = strings.Split(deviceTokens, ",")
		for i, t := range tokens {
			tokens[i] = strings.TrimSpace(t)
		}
	}

	notification := fcm.NewNotificationBuilder().
		SetTitle("Bulk Notification").
		SetBody(fmt.Sprintf("Gửi đến %d thiết bị", len(tokens))).
		SetImageURL("https://via.placeholder.com/150").
		Build()

	data := map[string]string{
		"type":      "bulk_notification",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		"test_id":   "test_send_to_tokens",
	}

	fmt.Printf("📤 Gửi notification đến %d tokens\n", len(tokens))
	fmt.Println()

	response, err := client.SendToTokens(ctx, tokens, notification, data)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Kết quả:\n")
	fmt.Printf("   Thành công: %d\n", response.SuccessCount)
	fmt.Printf("   Thất bại: %d\n", response.FailureCount)
	fmt.Println()

	// Hiển thị chi tiết lỗi nếu có
	if response.FailureCount > 0 {
		fmt.Println("📋 Chi tiết lỗi:")
		for i, resp := range response.Responses {
			if !resp.Success {
				fmt.Printf("   Token %d (%s...): %v\n", i+1, tokens[i][:min(20, len(tokens[i]))], resp.Error)
			}
		}
	}
}

// Test 3: Gửi notification đến topic
func testSendToTopic(ctx context.Context, client *fcm.Client) {
	fmt.Println("\n🧪 Test 3: Send Notification to Topic")
	fmt.Println("------------------------------------")

	if topic == "" {
		topic = "test_topic"
		fmt.Printf("⚠️  Chưa có topic, sử dụng topic mặc định: %s\n", topic)
	}

	notification := fcm.NewNotificationBuilder().
		SetTitle("Topic Notification").
		SetBody(fmt.Sprintf("Tin nhắn gửi đến topic: %s", topic)).
		Build()

	data := map[string]string{
		"type":      "topic_notification",
		"topic":     topic,
		"test_id":   "test_send_to_topic",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}

	fmt.Printf("📤 Gửi notification đến topic: %s\n", topic)
	fmt.Println()

	messageID, err := client.SendToTopic(ctx, topic, notification, data)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Thành công! Message ID: %s\n", messageID)
}

// Test 4: Gửi notification với condition
func testSendToCondition(ctx context.Context, client *fcm.Client) {
	fmt.Println("\n🧪 Test 4: Send Notification with Condition")
	fmt.Println("-------------------------------------------")

	if condition == "" {
		condition = "'news' in topics || 'sports' in topics"
		fmt.Printf("⚠️  Chưa có condition, sử dụng condition mặc định: %s\n", condition)
	}

	notification := fcm.NewNotificationBuilder().
		SetTitle("Conditional Notification").
		SetBody(fmt.Sprintf("Notification với condition: %s", condition)).
		Build()

	data := map[string]string{
		"type":      "conditional_notification",
		"condition": condition,
		"test_id":   "test_send_to_condition",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}

	fmt.Printf("📤 Gửi notification với condition: %s\n", condition)
	fmt.Println()

	messageID, err := client.SendToCondition(ctx, condition, notification, data)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Thành công! Message ID: %s\n", messageID)
}

// Test 5: Subscribe tokens vào topic
func testSubscribeToTopic(ctx context.Context, client *fcm.Client) {
	fmt.Println("\n🧪 Test 5: Subscribe Tokens to Topic")
	fmt.Println("------------------------------------")

	var tokens []string
	if deviceTokens == "" {
		if deviceToken == "" {
			log.Fatal("❌ Cần ít nhất một token để subscribe. Sử dụng -token hoặc -tokens")
		}
		tokens = []string{deviceToken}
	} else {
		tokens = strings.Split(deviceTokens, ",")
		for i, t := range tokens {
			tokens[i] = strings.TrimSpace(t)
		}
	}

	if topic == "" {
		topic = "test_topic"
		fmt.Printf("⚠️  Chưa có topic, sử dụng topic mặc định: %s\n", topic)
	}

	fmt.Printf("📤 Subscribe %d token(s) vào topic: %s\n", len(tokens), topic)
	fmt.Println()

	response, err := client.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Thành công!\n")
	fmt.Printf("   Success Count: %d\n", response.SuccessCount)
	if len(response.Errors) > 0 {
		fmt.Printf("   Errors:\n")
		for _, err := range response.Errors {
			fmt.Printf("     - %v\n", err)
		}
	}
}

// Test 6: Unsubscribe tokens khỏi topic
func testUnsubscribeFromTopic(ctx context.Context, client *fcm.Client) {
	fmt.Println("\n🧪 Test 6: Unsubscribe Tokens from Topic")
	fmt.Println("----------------------------------------")

	var tokens []string
	if deviceTokens == "" {
		if deviceToken == "" {
			log.Fatal("❌ Cần ít nhất một token để unsubscribe. Sử dụng -token hoặc -tokens")
		}
		tokens = []string{deviceToken}
	} else {
		tokens = strings.Split(deviceTokens, ",")
		for i, t := range tokens {
			tokens[i] = strings.TrimSpace(t)
		}
	}

	if topic == "" {
		topic = "test_topic"
		fmt.Printf("⚠️  Chưa có topic, sử dụng topic mặc định: %s\n", topic)
	}

	fmt.Printf("📤 Unsubscribe %d token(s) khỏi topic: %s\n", len(tokens), topic)
	fmt.Println()

	response, err := client.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Thành công!\n")
	fmt.Printf("   Success Count: %d\n", response.SuccessCount)
	if len(response.Errors) > 0 {
		fmt.Printf("   Errors:\n")
		for _, err := range response.Errors {
			fmt.Printf("     - %v\n", err)
		}
	}
}

// Test 7: Dry run (test mà không gửi thực sự)
func testDryRun(ctx context.Context, client *fcm.Client) {
	fmt.Println("\n🧪 Test 7: Dry Run (Test Notification)")
	fmt.Println("--------------------------------------")

	if deviceToken == "" {
		deviceToken = "YOUR_DEVICE_TOKEN_HERE"
	}

	notification := fcm.NewNotificationBuilder().
		SetTitle("Dry Run Test").
		SetBody("Đây là test dry-run, notification sẽ không được gửi thực sự").
		Build()

	data := map[string]string{
		"type":      "dry_run_test",
		"test_id":   "test_dry_run",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}

	fmt.Printf("📤 Dry-run notification đến token: %s\n", deviceToken[:min(20, len(deviceToken))]+"...")
	fmt.Println()

	messageID, err := client.SendDryRun(ctx, deviceToken, notification, data)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Dry-run thành công! Message ID: %s\n", messageID)
	fmt.Println("   (Notification này không được gửi thực sự đến thiết bị)")
}

// Test 8: Notification với Android config
func testAndroidConfig(ctx context.Context, client *fcm.Client) {
	fmt.Println("\n🧪 Test 8: Notification với Android Config")
	fmt.Println("------------------------------------------")

	if deviceToken == "" {
		deviceToken = "YOUR_DEVICE_TOKEN_HERE"
	}

	notification := fcm.NewNotificationBuilder().
		SetTitle("Android Custom Notification").
		SetBody("Notification với cấu hình Android tùy chỉnh").
		WithAndroidPriority("high").
		WithAndroidSound("default").
		WithAndroidColor("#FF5722").
		WithAndroidIcon("ic_notification").
		WithAndroidClickAction("OPEN_MAIN_ACTIVITY").
		WithAndroidTTL(3600).
		Build()

	data := map[string]string{
		"type":      "android_custom",
		"test_id":   "test_android_config",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}

	fmt.Printf("📤 Gửi Android custom notification\n")
	fmt.Printf("   Priority: high\n")
	fmt.Printf("   Sound: default\n")
	fmt.Printf("   Color: #FF5722\n")
	fmt.Printf("   Icon: ic_notification\n")
	fmt.Printf("   TTL: 3600s\n")
	fmt.Println()

	messageID, err := client.SendToToken(ctx, deviceToken, notification, data)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Thành công! Message ID: %s\n", messageID)
}

// Test 9: Notification với iOS config
func testIOSConfig(ctx context.Context, client *fcm.Client) {
	fmt.Println("\n🧪 Test 9: Notification với iOS Config")
	fmt.Println("-------------------------------------")

	if deviceToken == "" {
		deviceToken = "YOUR_DEVICE_TOKEN_HERE"
	}

	notification := fcm.NewNotificationBuilder().
		SetTitle("iOS Custom Notification").
		SetBody("Notification với cấu hình iOS tùy chỉnh").
		WithIOSBadge(5).
		WithIOSSound("default").
		WithIOSCategory("NEW_MESSAGE").
		WithIOSThreadID("thread-123").
		WithIOSContentAvailable().
		WithIOSMutableContent().
		Build()

	data := map[string]string{
		"type":      "ios_custom",
		"test_id":   "test_ios_config",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}

	fmt.Printf("📤 Gửi iOS custom notification\n")
	fmt.Printf("   Badge: 5\n")
	fmt.Printf("   Sound: default\n")
	fmt.Printf("   Category: NEW_MESSAGE\n")
	fmt.Printf("   Thread ID: thread-123\n")
	fmt.Printf("   Content Available: true\n")
	fmt.Printf("   Mutable Content: true\n")
	fmt.Println()

	messageID, err := client.SendToToken(ctx, deviceToken, notification, data)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Thành công! Message ID: %s\n", messageID)
}

// Test 10: Data-only notification (silent notification)
func testDataOnly(ctx context.Context, client *fcm.Client) {
	fmt.Println("\n🧪 Test 10: Data-Only Notification (Silent)")
	fmt.Println("--------------------------------------------")

	if deviceToken == "" {
		deviceToken = "YOUR_DEVICE_TOKEN_HERE"
	}

	data := map[string]string{
		"type":      "sync",
		"sync_id":   "abc123",
		"action":    "refresh_data",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		"test_id":   "test_data_only",
	}

	fmt.Printf("📤 Gửi data-only notification (không có title/body)\n")
	fmt.Printf("   Data: %v\n", data)
	fmt.Println()

	// Gửi với notification = nil để chỉ gửi data
	messageID, err := client.SendToToken(ctx, deviceToken, nil, data)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Thành công! Message ID: %s\n", messageID)
	fmt.Println("   (Đây là silent notification, sẽ không hiển thị trên thiết bị)")
}

// Test 11: Gửi nhiều messages khác nhau (SendAll)
func testSendAll(ctx context.Context, client *fcm.Client) {
	fmt.Println("\n🧪 Test 11: Send Multiple Different Messages (SendAll)")
	fmt.Println("-------------------------------------------------------")

	var tokens []string
	if deviceTokens == "" {
		if deviceToken == "" {
			log.Fatal("❌ Cần ít nhất một token. Sử dụng -token hoặc -tokens")
		}
		tokens = []string{deviceToken, deviceToken} // Dùng lại token để demo
	} else {
		tokens = strings.Split(deviceTokens, ",")
		for i, t := range tokens {
			tokens[i] = strings.TrimSpace(t)
		}
	}

	// Tạo nhiều messages khác nhau
	var messages []*messaging.Message
	for i, token := range tokens {
		notification := fcm.NewNotificationBuilder().
			SetTitle(fmt.Sprintf("Custom Message %d", i+1)).
			SetBody(fmt.Sprintf("Nội dung tùy chỉnh cho message %d", i+1)).
			Build()

		message := &messaging.Message{
			Token: token,
			Notification: &messaging.Notification{
				Title: notification.Title,
				Body:  notification.Body,
			},
			Data: map[string]string{
				"message_id": fmt.Sprintf("msg_%d", i+1),
				"type":       "custom_message",
				"timestamp":  fmt.Sprintf("%d", time.Now().Unix()),
			},
		}
		messages = append(messages, message)
	}

	fmt.Printf("📤 Gửi %d messages khác nhau\n", len(messages))
	fmt.Println()

	response, err := client.SendAll(ctx, messages)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("✅ Kết quả:\n")
	fmt.Printf("   Thành công: %d\n", response.SuccessCount)
	fmt.Printf("   Thất bại: %d\n", response.FailureCount)
}

// Test tất cả các tính năng
func testAllFeatures(ctx context.Context, client *fcm.Client) {
	fmt.Println("🚀 Chạy tất cả các test FCM...")
	fmt.Println()

	// Chỉ chạy các test không cần token thật nếu không có token
	if deviceToken == "" && deviceTokens == "" {
		fmt.Println("⚠️  Không có device token, sẽ bỏ qua các test cần token thật")
		fmt.Println()
	}

	// Test 1: Send to Token
	if deviceToken != "" {
		testSendToToken(ctx, client)
		time.Sleep(1 * time.Second)
	}

	// Test 2: Send to Multiple Tokens
	if deviceTokens != "" {
		testSendToTokens(ctx, client)
		time.Sleep(1 * time.Second)
	}

	// Test 3: Send to Topic
	testSendToTopic(ctx, client)
	time.Sleep(1 * time.Second)

	// Test 4: Send with Condition
	testSendToCondition(ctx, client)
	time.Sleep(1 * time.Second)

	// Test 5: Subscribe to Topic
	if deviceToken != "" || deviceTokens != "" {
		testSubscribeToTopic(ctx, client)
		time.Sleep(1 * time.Second)
	}

	// Test 6: Unsubscribe from Topic
	if deviceToken != "" || deviceTokens != "" {
		testUnsubscribeFromTopic(ctx, client)
		time.Sleep(1 * time.Second)
	}

	// Test 7: Dry Run
	if deviceToken != "" {
		testDryRun(ctx, client)
		time.Sleep(1 * time.Second)
	}

	// Test 8: Android Config
	if deviceToken != "" {
		testAndroidConfig(ctx, client)
		time.Sleep(1 * time.Second)
	}

	// Test 9: iOS Config
	if deviceToken != "" {
		testIOSConfig(ctx, client)
		time.Sleep(1 * time.Second)
	}

	// Test 10: Data-only
	if deviceToken != "" {
		testDataOnly(ctx, client)
		time.Sleep(1 * time.Second)
	}

	// Test 11: SendAll
	if deviceToken != "" || deviceTokens != "" {
		testSendAll(ctx, client)
	}

	fmt.Println("\n✅ Hoàn thành tất cả các test!")
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PrintUsage hiển thị hướng dẫn sử dụng
func printUsage() {
	fmt.Println("Cách sử dụng:")
	fmt.Println()
	fmt.Println("  # Test tất cả tính năng")
	fmt.Println("  go run examples/test_fcm.go -test=all")
	fmt.Println()
	fmt.Println("  # Gửi notification đến một token")
	fmt.Println("  go run examples/test_fcm.go -test=token -token=YOUR_DEVICE_TOKEN")
	fmt.Println()
	fmt.Println("  # Gửi notification đến nhiều tokens")
	fmt.Println("  go run examples/test_fcm.go -test=tokens -tokens=TOKEN1,TOKEN2,TOKEN3")
	fmt.Println()
	fmt.Println("  # Gửi notification đến topic")
	fmt.Println("  go run examples/test_fcm.go -test=topic -topic=news")
	fmt.Println()
	fmt.Println("  # Gửi notification với condition")
	fmt.Println("  go run examples/test_fcm.go -test=condition -condition=\"'news' in topics || 'sports' in topics\"")
	fmt.Println()
	fmt.Println("  # Subscribe tokens vào topic")
	fmt.Println("  go run examples/test_fcm.go -test=subscribe -tokens=TOKEN1,TOKEN2 -topic=news")
	fmt.Println()
	fmt.Println("  # Unsubscribe tokens khỏi topic")
	fmt.Println("  go run examples/test_fcm.go -test=unsubscribe -tokens=TOKEN1,TOKEN2 -topic=news")
	fmt.Println()
	fmt.Println("  # Dry run test")
	fmt.Println("  go run examples/test_fcm.go -test=dryrun -token=YOUR_DEVICE_TOKEN")
	fmt.Println()
	fmt.Println("  # Sử dụng credentials file khác")
	fmt.Println("  go run examples/test_fcm.go -credentials=keys/my-firebase-credentials.json -test=token -token=YOUR_TOKEN")
	fmt.Println()
}
