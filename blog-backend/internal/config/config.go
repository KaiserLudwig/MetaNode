// Package config 应用配置：支持环境变量覆盖，便于不同环境部署。
package config

import "os"

// Config 应用配置。
type Config struct {
	Port      string // 监听端口
	DBPath    string // SQLite 数据库文件路径
	JWTSecret string // JWT 签名密钥（生产环境务必通过环境变量注入）
	LogFile   string // 日志文件路径
	SeedDemo  bool   // 是否在空库中写入演示数据
}

// Load 从环境变量加载配置（缺失时使用开发默认值）。
func Load() *Config {
	return &Config{
		Port:      getenv("PORT", "8080"),
		DBPath:    getenv("DB_PATH", "blog.db"),
		JWTSecret: getenv("JWT_SECRET", "dev-secret-change-me-in-production"),
		LogFile:   getenv("LOG_FILE", "logs/app.log"),
		SeedDemo:  getenv("SEED_DEMO", "true") == "true",
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
