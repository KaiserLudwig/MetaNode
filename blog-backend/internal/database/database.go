// Package database 数据库连接、自动迁移与演示数据。
package database

import (
	"log/slog"

	"blog-backend/internal/config"
	"blog-backend/internal/models"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect 打开 SQLite 数据库并自动迁移所有模型。
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // 只打印警告与错误
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}); err != nil {
		return nil, err
	}
	slog.Info("数据库连接成功并完成迁移", "path", cfg.DBPath)
	return db, nil
}

// Seed 在空库中写入演示账号与文章，方便快速体验（可通过 SEED_DEMO=false 关闭）。
func Seed(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("demo123456"), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("生成演示密码失败", "error", err)
		return
	}
	demo := models.User{Username: "demo", Password: string(hashed), Email: "demo@example.com"}
	if err := db.Create(&demo).Error; err != nil {
		slog.Error("创建演示用户失败", "error", err)
		return
	}

	posts := []models.Post{
		{Title: "我的第一篇博客", Content: "欢迎来到我的个人博客！这是一个基于 Gin + GORM 构建的演示项目。", UserID: demo.ID},
		{Title: "Go 并发模型入门", Content: "goroutine 是 Go 并发编程的核心，配合 channel 可以实现优雅的协程间通信。", UserID: demo.ID},
		{Title: "GORM 使用技巧", Content: "预加载（Preload）、钩子函数（Hook）、事务（Transaction）是 GORM 最常用的高级特性。", UserID: demo.ID},
	}
	if err := db.Create(&posts).Error; err != nil {
		slog.Error("创建演示文章失败", "error", err)
		return
	}

	comments := []models.Comment{
		{Content: "写得太好了，收藏！", UserID: demo.ID, PostID: posts[0].ID},
		{Content: "期待更多内容！", UserID: demo.ID, PostID: posts[0].ID},
		{Content: "讲解得很清楚，学到了。", UserID: demo.ID, PostID: posts[1].ID},
	}
	if err := db.Create(&comments).Error; err != nil {
		slog.Error("创建演示评论失败", "error", err)
		return
	}
	slog.Info("已写入演示数据", "账号", "demo / demo123456")
}
