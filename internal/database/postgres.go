package database

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/dmehra2102/hr-management-system/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db     *gorm.DB
	config config.DatabaseConfig
}

func New(cfg config.DatabaseConfig) (*Database, error) {
	gormLogger := logger.Default
	gormLogger = gormLogger.LogMode(logger.Info)

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("filed to ping database: %w", err)
	}

	return &Database{
		db:     db,
		config: cfg,
	}, nil
}

func (d *Database) GetDB() *gorm.DB {
	return d.db
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB insatnce: %w", err)
	}
	return sqlDB.Close()
}

// Migrate run database migrations
func (d *Database) Migrate() error {
	migrationPath, err := filepath.Abs("internal/database/migrations")
	if err != nil {
		return fmt.Errorf("failed to get migrations path: %w", err)
	}

	m, err := migrate.New(fmt.Sprintf("file://%s", migrationPath), d.config.DSN())
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			log.Printf("migration close error: sourceErr=%v, dbErr=%v", sourceErr, dbErr)
		}
	}()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// MigrateDown rolls back all migrations
func (d *Database) MigrationDone() error {
	migrationPath, err := filepath.Abs("internal/database/migrations")
	if err != nil {
		return fmt.Errorf("failed to get migrations path: %w", err)
	}

	m, err := migrate.New(fmt.Sprintf("file://%s", migrationPath), d.config.DSN())
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			log.Printf("migration close error: sourceErr=%v, dbErr=%v", sourceErr, dbErr)
		}
	}()

	err = m.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// MigrateToVersion migrates to a specific version
func (d *Database) MigrationToVersion() (uint, bool, error) {
	migrationPath, err := filepath.Abs("internal/database/migrations")
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migrations path: %w", err)
	}

	m, err := migrate.New(fmt.Sprintf("file://%s", migrationPath), d.config.DSN())
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			log.Printf("migration close error: sourceErr=%v, dbErr=%v", sourceErr, dbErr)
		}
	}()

	version, dirty, err := m.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, err
}

// HealthCheck performs a health check on the database
func (d *Database) HealthCheck() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB instance: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
