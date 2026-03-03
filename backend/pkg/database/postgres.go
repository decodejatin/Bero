package database

import (
	"fmt"
	"log"
	"time"

	"github.com/decodejatin/bero-backend/config"
	"github.com/decodejatin/bero-backend/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establishes database connection and runs migrations
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Kolkata",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Connection pool tuning for production throughput
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Enable PostGIS extension (idempotent)
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error; err != nil {
		log.Printf("⚠️  PostGIS extension not available (non-fatal): %v", err)
	}

	// Run auto migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	DB = db
	log.Println("Database connected successfully")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.WorkerProfile{},
		&domain.ClientProfile{},
		&domain.Job{},
		&domain.JobAcceptance{},
		&domain.JobCompletion{},
		&domain.OtpRequest{},
		&domain.Session{},
		&domain.ChatConversation{},
		&domain.ChatMessage{},
		&domain.SavedAddress{},
		&domain.MatchingWeights{},
		&domain.StabilityConfig{},
		&domain.StabilityEvent{},
		&domain.PricingConfig{},
		&domain.SurgeHistory{},
		&domain.MutualRating{},
	)
}
