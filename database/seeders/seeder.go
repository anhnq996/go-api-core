package seeders

import (
	"fmt"

	"gorm.io/gorm"
)

// Seeder định nghĩa một seeder
type Seeder struct {
	Name string
	Run  func(*gorm.DB) error
}

// AllSeeders danh sách tất cả seeders
var AllSeeders = []Seeder{
	{
		Name: "RoleSeeder",
		Run:  SeedRoles,
	},
	{
		Name: "PermissionSeeder",
		Run:  SeedPermissions,
	},
	{
		Name: "RolePermissionSeeder",
		Run:  SeedRolePermissions,
	},
	{
		Name: "UserSeeder",
		Run:  SeedUsers,
	},
}

// RunSeeders chạy tất cả seeders
func RunSeeders(db *gorm.DB) error {
	fmt.Println("Running seeders...")

	for _, seeder := range AllSeeders {
		fmt.Printf("\n📦 Running seeder: %s\n", seeder.Name)
		if err := seeder.Run(db); err != nil {
			return fmt.Errorf("failed to run seeder %s: %w", seeder.Name, err)
		}
	}

	fmt.Println("\n✅ All seeders completed successfully")
	return nil
}
