package main

import (
	"flag"
	"fmt"
	"os"

	"anhnq/api-core/config"
	"anhnq/api-core/database"
	"anhnq/api-core/database/seeders"

	"gorm.io/gorm"
)

func main() {
	// Subcommands
	upCmd := flag.NewFlagSet("up", flag.ExitOnError)
	downCmd := flag.NewFlagSet("down", flag.ExitOnError)
	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)
	forceCmd := flag.NewFlagSet("force", flag.ExitOnError)
	stepsCmd := flag.NewFlagSet("steps", flag.ExitOnError)
	seedCmd := flag.NewFlagSet("seed", flag.ExitOnError)

	// Force flags
	forceVersion := forceCmd.Int("version", 0, "Version to force")

	// Steps flags
	stepsN := stepsCmd.Int("n", 0, "Number of steps (positive=up, negative=down)")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Connect to database
	dbConfig := config.GetDefaultDatabaseConfig()
	db, err := config.ConnectDatabase(dbConfig)
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Parse subcommand
	command := os.Args[1]

	// Seed command không cần migrator
	if command == "seed" {
		seedCmd.Parse(os.Args[2:])
		runSeed(db)
		return
	}

	// Create migrator cho migration commands
	migrator, err := database.NewMigrator(db, "database/migrations")
	if err != nil {
		fmt.Printf("❌ Failed to create migrator: %v\n", err)
		os.Exit(1)
	}
	defer migrator.Close()

	switch command {
	case "up":
		upCmd.Parse(os.Args[2:])
		runUp(migrator)
	case "down":
		downCmd.Parse(os.Args[2:])
		runDown(migrator)
	case "version":
		versionCmd.Parse(os.Args[2:])
		showVersion(migrator)
	case "force":
		forceCmd.Parse(os.Args[2:])
		runForce(migrator, *forceVersion)
	case "steps":
		stepsCmd.Parse(os.Args[2:])
		runSteps(migrator, *stepsN)
	default:
		printUsage()
		os.Exit(1)
	}
}

func runUp(m *database.Migrator) {
	version, dirty, _ := m.Version()
	fmt.Printf("Current version: %d (dirty: %v)\n", version, dirty)

	fmt.Println("Running migrations up...")
	if err := m.Up(); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}

	version, dirty, _ = m.Version()
	fmt.Printf("✅ Migrations completed. New version: %d (dirty: %v)\n", version, dirty)
}

func runDown(m *database.Migrator) {
	version, dirty, _ := m.Version()
	fmt.Printf("Current version: %d (dirty: %v)\n", version, dirty)

	fmt.Println("Rolling back all migrations...")
	if err := m.Down(); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ All migrations rolled back")
}

func showVersion(m *database.Migrator) {
	version, dirty, err := m.Version()
	if err != nil {
		fmt.Println("Version: none (no migrations run yet)")
		return
	}

	fmt.Printf("Current version: %d\n", version)
	fmt.Printf("Dirty: %v\n", dirty)
}

func runForce(m *database.Migrator, version int) {
	if version == 0 {
		fmt.Println("❌ Please provide version with -version flag")
		os.Exit(1)
	}

	fmt.Printf("Forcing version to %d...\n", version)
	if err := m.Force(version); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Version forced to %d\n", version)
}

func runSteps(m *database.Migrator, n int) {
	if n == 0 {
		fmt.Println("❌ Please provide number of steps with -n flag")
		os.Exit(1)
	}

	version, dirty, _ := m.Version()
	fmt.Printf("Current version: %d (dirty: %v)\n", version, dirty)

	direction := "up"
	if n < 0 {
		direction = "down"
	}

	fmt.Printf("Running %d steps %s...\n", abs(n), direction)
	if err := m.Steps(n); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}

	version, dirty, _ = m.Version()
	fmt.Printf("✅ Steps completed. New version: %d (dirty: %v)\n", version, dirty)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func runSeed(db *gorm.DB) {
	fmt.Println("Running seeders...")
	if err := seeders.RunSeeders(db); err != nil {
		fmt.Printf("❌ Failed to run seeders: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Seeders completed successfully")
}

func printUsage() {
	fmt.Print(`
Database Tool - Quản lý migrations và seeders

Usage:
  go run cmd/migrate/main.go <command> [options]

Commands:
  up                Run all pending migrations
  down              Rollback all migrations
  version           Show current migration version
  force             Force set migration version (use when dirty)
  steps             Run N migration steps
  seed              Run database seeders

Examples:
  # Migrations
  go run cmd/migrate/main.go up
  go run cmd/migrate/main.go down
  go run cmd/migrate/main.go version
  go run cmd/migrate/main.go force -version 1
  go run cmd/migrate/main.go steps -n 1      # Run 1 migration up
  go run cmd/migrate/main.go steps -n -1     # Rollback 1 migration

  # Seeders
  go run cmd/migrate/main.go seed

Options:
  force -version <N>    Version number to force
  steps -n <N>          Number of steps (positive=up, negative=down)
`)
}
