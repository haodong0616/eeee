package database

import (
	"expchange-backend/models"
	"log"
	"strconv"
	"sync"
)

// SystemConfigManager 系统配置管理器（支持热更新）
type SystemConfigManager struct {
	configs map[string]string
	mu      sync.RWMutex
}

var (
	sysConfigInstance *SystemConfigManager
	sysConfigOnce     sync.Once
)

// GetSystemConfigManager 获取系统配置管理器单例
func GetSystemConfigManager() *SystemConfigManager {
	sysConfigOnce.Do(func() {
		sysConfigInstance = &SystemConfigManager{
			configs: make(map[string]string),
		}
		sysConfigInstance.LoadFromDB()
	})
	return sysConfigInstance
}

// LoadFromDB 从数据库加载配置
func (m *SystemConfigManager) LoadFromDB() {
	var configs []models.SystemConfig
	DB.Find(&configs)

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, config := range configs {
		m.configs[config.Key] = config.Value
	}

	log.Printf("✅ 已加载 %d 个系统配置", len(configs))
}

// Get 获取配置值
func (m *SystemConfigManager) Get(key string, defaultValue string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if value, exists := m.configs[key]; exists {
		return value
	}
	return defaultValue
}

// GetInt 获取整数配置
func (m *SystemConfigManager) GetInt(key string, defaultValue int) int {
	value := m.Get(key, "")
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetBool 获取布尔配置
func (m *SystemConfigManager) GetBool(key string, defaultValue bool) bool {
	value := m.Get(key, "")
	if value == "" {
		return defaultValue
	}

	return value == "true" || value == "1"
}

// Set 设置配置值（热更新）
func (m *SystemConfigManager) Set(key, value string) {
	m.mu.Lock()
	m.configs[key] = value
	m.mu.Unlock()

	log.Printf("🔄 配置已热更新: %s = %s", key, value)
}

// Reload 重新加载所有配置
func (m *SystemConfigManager) Reload() {
	m.LoadFromDB()
}



