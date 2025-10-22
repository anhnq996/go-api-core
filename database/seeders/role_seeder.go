package seeders

import (
	model "anhnq/api-core/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedRoles seed roles data
func SeedRoles(db *gorm.DB) error {
	roles := []model.Role{
		{
			ID:          uuid.New(),
			Name:        "admin",
			DisplayName: "Administrator",
			Description: "Full system access with all permissions",
		},
		{
			ID:          uuid.New(),
			Name:        "moderator",
			DisplayName: "Moderator",
			Description: "Can manage content and users",
		},
		{
			ID:          uuid.New(),
			Name:        "user",
			DisplayName: "User",
			Description: "Regular user with basic permissions",
		},
	}

	for _, role := range roles {
		var existing model.Role
		if err := db.Where("name = ?", role.Name).First(&existing).Error; err != nil {
			// Role doesn't exist, create it
			if err := db.Create(&role).Error; err != nil {
				return err
			}
		} else {
			// Role exists, update it
			role.ID = existing.ID
			if err := db.Model(&existing).Updates(role).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
