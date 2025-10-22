package seeders

import (
	model "anhnq/api-core/internal/models"
	"anhnq/api-core/pkg/utils"
	"fmt"

	"gorm.io/gorm"
)

// SeedUsers tạo dữ liệu mẫu cho users
func SeedUsers(db *gorm.DB) error {
	// Hash password: Password123!
	hashedPassword, err := utils.HashPassword("Password123!")
	if err != nil {
		return err
	}

	// Query all roles from database
	var roles []model.Role
	if err := db.Find(&roles).Error; err != nil {
		return fmt.Errorf("failed to query roles: %w", err)
	}

	// Create role map for quick lookup by name
	roleMap := make(map[string]model.Role)
	for _, role := range roles {
		roleMap[role.Name] = role
	}

	// Define users with role names
	type UserSeed struct {
		Name     string
		Email    string
		RoleName string
	}

	userSeeds := []UserSeed{
		{
			Name:     "Admin User",
			Email:    "admin@example.com",
			RoleName: "admin",
		},
		{
			Name:     "Moderator User",
			Email:    "moderator@example.com",
			RoleName: "moderator",
		},
		{
			Name:     "Regular User",
			Email:    "user@example.com",
			RoleName: "user",
		},
		{
			Name:     "John Doe",
			Email:    "john@example.com",
			RoleName: "user",
		},
		{
			Name:     "Jane Smith",
			Email:    "jane@example.com",
			RoleName: "user",
		},
	}

	for _, userSeed := range userSeeds {
		// Get role from map
		role, roleExists := roleMap[userSeed.RoleName]
		if !roleExists {
			fmt.Printf("  ⚠️  Role '%s' not found, skipping user %s\n", userSeed.RoleName, userSeed.Email)
			continue
		}

		// Check if user already exists
		var existingUser model.User
		if err := db.Where("email = ?", userSeed.Email).First(&existingUser).Error; err == nil {
			// User exists, update if needed
			existingUser.Name = userSeed.Name
			existingUser.Password = hashedPassword
			existingUser.RoleID = &role.ID
			existingUser.IsActive = true

			if err := db.Model(&existingUser).Updates(existingUser).Error; err != nil {
				return fmt.Errorf("failed to update user %s: %w", userSeed.Email, err)
			}
			fmt.Printf("  ✅ Updated user: %s (%s) - Role: %s\n", existingUser.Name, existingUser.Email, role.Name)
			continue
		}

		// Create new user
		newUser := model.User{
			Name:     userSeed.Name,
			Email:    userSeed.Email,
			Password: hashedPassword,
			RoleID:   &role.ID,
			IsActive: true,
		}

		if err := db.Create(&newUser).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", userSeed.Email, err)
		}
		fmt.Printf("  ✅ Created user: %s (%s) - Role: %s\n", newUser.Name, newUser.Email, role.Name)
	}

	return nil
}

// ClearUsers xóa tất cả users
func ClearUsers(db *gorm.DB) error {
	return db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE").Error
}
