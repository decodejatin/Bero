package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/decodejatin/bero-backend/config"
	"github.com/decodejatin/bero-backend/pkg/database"
)

func main() {
	// Setup env for Supabase before loading config
	os.Setenv("DB_HOST", "db.pjmllcautezxszhwrwnd.supabase.co")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "Gagan@1008Gagan")
	os.Setenv("DB_NAME", "postgres")

	cfg := config.LoadConfig()

	// This connects and runs GORM AutoMigrate
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Unable to connect and automigrate: %v\n", err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql db: %v\n", err)
	}
	defer sqlDB.Close()

	migrationsDir := "../../migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		migrationsDir = "migrations"
		files, err = os.ReadDir(migrationsDir)
		if err != nil {
			log.Fatalf("Unable to read migrations directory: %v\n", err)
		}
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}

	sort.Strings(sqlFiles)

	for _, file := range sqlFiles {
		log.Printf("Applying migration %s...\n", file)
		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			log.Fatalf("Unable to read file %s: %v\n", file, err)
		}

		err = db.Exec(string(content)).Error
		if err != nil {
			log.Fatalf("Error executing migration %s: %v\n", file, err)
		}
		log.Printf("Successfully applied %s\n", file)
	}

	fmt.Println("All migrations applied successfully!")
}
