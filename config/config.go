package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3Endpoint         string
	S3Bucket           string
	ServerPort         string
	MaxUploadSize      int
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		AWSRegion:          getEnv("AWS_REGION", "ap-southeast-1"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", "test"),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", "test"),
		S3Endpoint:         getEnv("S3_ENDPOINT", "http://localhost:4566"),
		S3Bucket:           getEnv("S3_BUCKET", "my-bucket"),
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		MaxUploadSize:      getEnvInt("MAX_UPLOAD_SIZE", 100),
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
