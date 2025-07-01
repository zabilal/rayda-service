package database

import (
	"context"
	"fmt"
	"time"

	"github.com/rayda/rayda-service/internal/config"
	"github.com/rayda/rayda-service/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection
var DB *gorm.DB

// Connect initializes the database connection and runs migrations
func Connect(cfg *config.Config) error {
	// Create DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	// Configure GORM logger
	gormLogger := logger.Default
	if cfg.Server.Environment == "production" {
		gormLogger = gormLogger.LogMode(logger.Silent)
	}

	// Connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	// Set the global DB instance
	DB = db

	return nil
}

// Migrate runs database migrations
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Enable UUID extension if not exists
	if err := DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	// Run migrations for all models
	err := DB.AutoMigrate(
		&model.Organization{},
		&model.User{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// WithTenantScope returns a database session scoped to a specific tenant
func WithTenantScope(db *gorm.DB, tenantID string) *gorm.DB {
	return db.Where("tenant_id = ?", tenantID)
}

// BeginTransaction starts a new database transaction
func BeginTransaction(ctx context.Context) (*gorm.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Set the transaction in the context
	return tx.WithContext(ctx), nil
}

// CommitTransaction commits a transaction
func CommitTransaction(tx *gorm.DB) error {
	return tx.Commit().Error
}

// RollbackTransaction rolls back a transaction
func RollbackTransaction(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	return sqlDB.Close()
}
