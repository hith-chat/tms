package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/db"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/migrate/main.go [up|down|version]")
	}

	command := os.Args[1]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	switch command {
	case "up":
		err = runMigrations(database)
		if err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully")
	case "down":
		// In a real system, you might want to implement rollback functionality
		fmt.Println("Migration rollback not implemented")
	case "version":
		version, err := getCurrentMigrationVersion(database)
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		fmt.Printf("Current migration version: %s\n", version)
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func runMigrations(database *db.DB) error {
	// Create migration tracking table if it doesn't exist
	_, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migration tracking table: %w", err)
	}

	// This is a simple approach - in production you might want to use
	// a proper migration library like golang-migrate
	migrations := []string{
		"migrations/022_knowledge_management_system.sql",
		"migrations/023_nullable_embeddings.sql",
		"migrations/024_prevent_duplicate_pages.sql",
		"migrations/025_project_level_deduplication.sql",
		"migrations/026_add_content_hash.sql",
		"migrations/027_alarm_system.sql",
		"migrations/028_agent_enhancements.sql",
		"migrations/029_credits_system.sql",
		"migrations/030_knowledge_faq.sql",
		"migrations/031_public_projects.sql",
	}

	for _, migration := range migrations {
		// Check if migration has already been applied
		var count int
		migrationName := strings.TrimPrefix(migration, "migrations/")
		migrationName = strings.TrimSuffix(migrationName, ".sql")

		err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", migrationName).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", migrationName, err)
		}

		if count > 0 {
			fmt.Printf("Skipping already applied migration: %s\n", migration)
			continue
		}

		fmt.Printf("Running migration: %s\n", migration)
		content, err := os.ReadFile(migration)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migration, err)
		}

		// Parse goose format - extract SQL between +goose Up and +goose Down
		sql := parseGooseMigration(string(content))
		if sql == "" {
			return fmt.Errorf("no valid SQL found in migration %s", migration)
		}

		_, err = database.Exec(sql)
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migration, err)
		}

		// Record migration as applied
		_, err = database.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migrationName)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
		}
	}

	return nil
}

// parseGooseMigration extracts the SQL content between +goose Up and +goose Down markers
func parseGooseMigration(content string) string {
	lines := strings.Split(content, "\n")
	var sqlLines []string
	inUpSection := false
	inStatementBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "-- +goose Up") {
			inUpSection = true
			continue
		}

		if strings.Contains(line, "-- +goose Down") {
			break
		}

		if strings.Contains(line, "-- +goose StatementBegin") {
			inStatementBlock = true
			continue
		}

		if strings.Contains(line, "-- +goose StatementEnd") {
			inStatementBlock = false
			continue
		}

		if inUpSection && (inStatementBlock || !strings.HasPrefix(line, "-- +goose")) {
			sqlLines = append(sqlLines, line)
		}
	}

	return strings.Join(sqlLines, "\n")
}

func getCurrentMigrationVersion(database *db.DB) (string, error) {
	// This is a simple implementation - in production you'd track migration versions properly
	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'agents'").Scan(&count)
	if err != nil {
		return "", err
	}

	if count > 0 {
		return "003_seed_data", nil
	}

	return "none", nil
}
