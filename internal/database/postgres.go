package database

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/svetilka/EffectiveMobileProject/internal/config"
	"github.com/svetilka/EffectiveMobileProject/internal/models"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
}

func NewDatabase(cfg *config.Config) (*DB, error) {
	log.WithFields(log.Fields{
		"host": cfg.DBHost,
		"port": cfg.DBPort,
		"db":   cfg.DBName,
	}).Info("Connecting to database")

	dsn := cfg.GetDBConnectionString()

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.WithError(err).Error("Failed to connect to database")
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Successfully connected to database")

	return &DB{db}, nil
}

func (db *DB) RunMigrations() error {
	log.Info("Running database migrations")

	// Auto migrate
	if err := db.AutoMigrate(&models.Subscription{}); err != nil {
		log.WithError(err).Error("Failed to auto migrate")
		return err
	}

	// Run SQL migrations
	migrationFiles := []string{
		"internal/migrations/001_create_subscriptions_table.sql",
	}

	for _, file := range migrationFiles {
		migrationSQL, err := ioutil.ReadFile(file)
		if err != nil {
			log.WithError(err).Warnf("Failed to read migration file: %s", file)
			continue
		}

		statements := strings.Split(string(migrationSQL), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if err := db.Exec(stmt).Error; err != nil {
				// Ignore "already exists" errors
				if !strings.Contains(err.Error(), "already exists") {
					log.WithError(err).Warnf("Failed to execute migration: %s", stmt)
				}
			}
		}
	}

	log.Info("Migrations completed successfully")
	return nil
}

func (db *DB) Close() {
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.WithError(err).Error("Failed to get underlying SQL DB")
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.WithError(err).Error("Failed to close database connection")
	} else {
		log.Info("Database connection closed")
	}
}
