package seeders

import (
	model "anhnq/api-core/internal/models"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedUsers tạo dữ liệu mẫu cho users
func SeedUsers(db *gorm.DB) error {
	users := []model.User{
		{
			ID:    uuid.New().String(),
			Name:  "Admin User",
			Email: "admin@example.com",
		},
		{
			ID:    uuid.New().String(),
			Name:  "John Doe",
			Email: "john@example.com",
		},
		{
			ID:    uuid.New().String(),
			Name:  "Jane Smith",
			Email: "jane@example.com",
		},
		{
			ID:    uuid.New().String(),
			Name:  "Nguyễn Văn A",
			Email: "nguyenvana@example.com",
		},
		{
			ID:    uuid.New().String(),
			Name:  "Trần Thị B",
			Email: "tranthib@example.com",
		},
	}

	for _, user := range users {
		// Check if user already exists
		var existingUser model.User
		if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			fmt.Printf("  - User %s already exists, skipping\n", user.Email)
			continue
		}

		// Create user
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Email, err)
		}
		fmt.Printf("  ✅ Created user: %s (%s)\n", user.Name, user.Email)
	}

	return nil
}

// ClearUsers xóa tất cả users
func ClearUsers(db *gorm.DB) error {
	return db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE").Error
}
