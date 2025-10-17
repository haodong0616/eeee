package database

import (
	"fmt"

	"expchange-backend/config"
	"expchange-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
)

func InitDB(cfg *config.Config) error {
	var err error
	var dialector gorm.Dialector

	// 根据配置选择数据库类型
	dbType := cfg.DBType
	if dbType == "" {
		dbType = "mysql"
	}

	switch dbType {
	case "mysql":
		// MySQL 连接字符串格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
		)
		dialector = mysql.Open(dsn)
	case "sqlite":
		// 使用 SQLite 数据库
		dbPath := cfg.DBName
		if dbPath == "" {
			dbPath = "expchange.db"
		}
		dialector = sqlite.Open(dbPath)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error), // 只记录错误级别的日志
	})
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
		&models.SystemConfig{},
		&models.ChainConfig{},
		&models.Task{},
		&models.TaskLog{},
		&models.MarketMakerPnL{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}
