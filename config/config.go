package config

import "os"

type Config struct {
	Port         string
	DatabasePath string
	AppID        string
	AppSecret    string
}

func LoadConfig() *Config {
	cfg := &Config{}

	// 获取环境变量，使用默认值
	cfg.Port = getEnvOrDefault("PORT", "8080")
	cfg.DatabasePath = getEnvOrDefault("DATABASE_PATH", "./data/feishu_api.db")
	cfg.AppID = os.Getenv("APP_ID")         // 必须通过环境变量设置
	cfg.AppSecret = os.Getenv("APP_SECRET") // 必须通过环境变量设置

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
