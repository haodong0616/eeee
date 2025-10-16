package database

import (
	"context"
	"fmt"

	"expchange-backend/config"
	"expchange-backend/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
)

func InitDB(cfg *config.Config) error {
	// 使用SQLite数据库
	dbPath := cfg.DBName
	if dbPath == "" {
		dbPath = "expchange.db"
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate models
	err = DB.AutoMigrate(
		&models.User{},
		&models.TradingPair{},
		&models.Balance{},
		&models.Order{},
		&models.Trade{},
		&models.Kline{},
		&models.FeeConfig{},
		&models.FeeRecord{},
		&models.DepositRecord{},
		&models.WithdrawRecord{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

func InitRedis(cfg *config.Config) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	})

	ctx := context.Background()
	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	return nil
}
