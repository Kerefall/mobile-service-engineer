package config

import (
    "os"
    "strconv"
    "github.com/joho/godotenv"
    "github.com/sirupsen/logrus"
)

type Config struct {
    ServerPort string
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    JWTSecret  string
    MinioEndpoint string
    MinioAccessKey string
    MinioSecretKey string
    MinioBucket string
}

func Load() *Config {
    // Загружаем .env файл если есть
    if err := godotenv.Load(); err != nil {
        logrus.Warn("No .env file found, using environment variables")
    }

    return &Config{
        ServerPort: getEnv("SERVER_PORT", "8080"),
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "5432"),
        DBUser:     getEnv("DB_USER", "postgres"),
        DBPassword: getEnv("DB_PASSWORD", "postgres"),
        DBName:     getEnv("DB_NAME", "mobile_engineer"),
        JWTSecret:  getEnv("JWT_SECRET", "your-secret-key-change-me"),
        MinioEndpoint: getEnv("MINIO_ENDPOINT", "localhost:9000"),
        MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
        MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
        MinioBucket: getEnv("MINIO_BUCKET", "mobile-engineer"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}