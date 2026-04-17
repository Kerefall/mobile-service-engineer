package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ServerPort     string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	JWTSecret      string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
	MinioPublicURL string
	// StorageDriver: "disk" (по умолчанию) или "minio"
	StorageDriver string
	FCMServerKey  string
	AdminSecret   string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "mobile_engineer"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key-change-me"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    getEnv("MINIO_BUCKET", "mobile-engineer"),
		MinioUseSSL:    getEnvBool("MINIO_USE_SSL", false),
		MinioPublicURL: strings.TrimRight(getEnv("MINIO_PUBLIC_URL", "http://127.0.0.1:9000"), "/"),
		StorageDriver:  strings.ToLower(getEnv("STORAGE_DRIVER", "disk")),
		FCMServerKey:   getEnv("FCM_SERVER_KEY", ""),
		AdminSecret:    getEnv("ADMIN_SECRET", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		v, err := strconv.ParseBool(value)
		if err == nil {
			return v
		}
	}
	return defaultValue
}
