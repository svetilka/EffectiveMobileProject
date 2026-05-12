package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	ServerPort string

	LogLevel string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found, using environment variables")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres123"),
		DBName:     getEnv("DB_NAME", "subscriptions"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func SetupLogger(level string) {
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(logLevel)
	}

	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

func (c *Config) GetDBConnectionString() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

func (c *Config) GetServerAddress() string {
	return ":" + c.ServerPort
}

func parseIntOrDefault(str string, defaultValue int) int {
	if val, err := strconv.Atoi(str); err == nil {
		return val
	}
	return defaultValue
}
