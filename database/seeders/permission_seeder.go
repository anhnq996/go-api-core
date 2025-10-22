package seeders

import (
	model "anhnq/api-core/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedPermissions seed permissions data
func SeedPermissions(db *gorm.DB) error {
	permissions := []model.Permission{
		// User permissions
		{
			ID:          uuid.New(),
			Name:        "users.view",
			DisplayName: "View Users",
			Description: "Can view user list and details",
			Module:      "users",
		},
		{
			ID:          uuid.New(),
			Name:        "users.create",
			DisplayName: "Create Users",
			Description: "Can create new users",
			Module:      "users",
		},
		{
			ID:          uuid.New(),
			Name:        "users.update",
			DisplayName: "Update Users",
			Description: "Can update user information",
			Module:      "users",
		},
		{
			ID:          uuid.New(),
			Name:        "users.delete",
			DisplayName: "Delete Users",
			Description: "Can delete users",
			Module:      "users",
		},

		// Role permissions
		{
			ID:          uuid.New(),
			Name:        "roles.view",
			DisplayName: "View Roles",
			Description: "Can view roles",
			Module:      "roles",
		},
		{
			ID:          uuid.New(),
			Name:        "roles.manage",
			DisplayName: "Manage Roles",
			Description: "Can create, update, delete roles",
			Module:      "roles",
		},

		// Permission permissions
		{
			ID:          uuid.New(),
			Name:        "permissions.view",
			DisplayName: "View Permissions",
			Description: "Can view permissions",
			Module:      "permissions",
		},
		{
			ID:          uuid.New(),
			Name:        "permissions.manage",
			DisplayName: "Manage Permissions",
			Description: "Can assign/revoke permissions",
			Module:      "permissions",
		},

		// Profile permissions
		{
			ID:          uuid.New(),
			Name:        "profile.view",
			DisplayName: "View Own Profile",
			Description: "Can view own profile",
			Module:      "profile",
		},
		{
			ID:          uuid.New(),
			Name:        "profile.update",
			DisplayName: "Update Own Profile",
			Description: "Can update own profile",
			Module:      "profile",
		},
	}

	for _, permission := range permissions {
		var existing model.Permission
		if err := db.Where("name = ?", permission.Name).First(&existing).Error; err != nil {
			// Permission doesn't exist, create it
			if err := db.Create(&permission).Error; err != nil {
				return err
			}
		} else {
			// Permission exists, update it
			permission.ID = existing.ID
			if err := db.Model(&existing).Updates(permission).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
