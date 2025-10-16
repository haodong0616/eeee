package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort  string
	DBType      string // mysql or sqlite
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	JWTSecret   string
	CORSOrigins string
}

func Load() (*Config, error) {
	godotenv.Load()

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8383"),
		// MySQL 配置（使用共享 Docker MySQL）
		DBType:     "mysql",
		DBHost:     "localhost",
		DBPort:     "3308",
		DBUser:     "referral_user",
		DBPassword: "referral123456",
		DBName:     "expchange",
		// JWT 配置
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
