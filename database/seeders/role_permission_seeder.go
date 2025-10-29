package seeders

import (
	model "api-core/internal/models"

	"gorm.io/gorm"
)

// SeedRolePermissions seed role-permission relationships
func SeedRolePermissions(db *gorm.DB) error {
	// Clear existing role-permission relationships
	db.Where("1 = 1").Delete(&model.RoleHasPermission{})

	// Define role-permission mapping by name
	rolePermissionMap := map[string][]string{
		"admin": {
			// Admin có tất cả permissions
			"users.view",
			"users.create",
			"users.update",
			"users.delete",
			"roles.view",
			"roles.manage",
			"permissions.view",
			"permissions.manage",
			"profile.view",
			"profile.update",
		},
		"moderator": {
			// Moderator có quyền hạn chế
			"users.view",
			"users.update",
			"profile.view",
			"profile.update",
		},
		"user": {
			// User chỉ có quyền cơ bản
			"profile.view",
			"profile.update",
		},
	}

	// Get all roles and permissions from database
	var roles []model.Role
	if err := db.Find(&roles).Error; err != nil {
		return err
	}

	var permissions []model.Permission
	if err := db.Find(&permissions).Error; err != nil {
		return err
	}

	// Create maps for quick lookup
	roleMap := make(map[string]model.Role)
	for _, role := range roles {
		roleMap[role.Name] = role
	}

	permissionMap := make(map[string]model.Permission)
	for _, permission := range permissions {
		permissionMap[permission.Name] = permission
	}

	// Seed role-permission relationships
	for roleName, permissionNames := range rolePermissionMap {
		role, roleExists := roleMap[roleName]
		if !roleExists {
			// Skip if role doesn't exist
			continue
		}

		for _, permName := range permissionNames {
			permission, permExists := permissionMap[permName]
			if !permExists {
				// Skip if permission doesn't exist
				continue
			}

			// Create relationship
			roleHasPermission := model.RoleHasPermission{
				RoleID:       role.ID,
				PermissionID: permission.ID,
			}

			if err := db.Create(&roleHasPermission).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
