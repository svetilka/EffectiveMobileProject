package database

import (
	"fmt"

	"github.com/svetilka/EffectiveMobileProject/internal/config"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres" // драйвер для GORM
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
	log.Info("Running database migrations via golang-migrate")

	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Оставляем структуру пустой, так как дефолтные настройки подходят
	// и это гарантирует успешную компиляцию
	driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate instance: %w", err)
	}

	// Выполняем накат миграций
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.WithError(err).Error("Migration execution failed")
		return fmt.Errorf("migration up failed: %w", err)
	}

	// ВАЖНО: Мы НЕ вызываем m.Close().
	// Это предотвращает закрытие sqlDB для последующего запуска HTTP-сервера GORM.

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
