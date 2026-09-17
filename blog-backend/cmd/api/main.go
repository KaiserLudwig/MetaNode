// 个人博客系统后端入口：Gin + GORM + JWT。
package main

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/database"
	"blog-backend/internal/handlers"
	"blog-backend/internal/middleware"
	"blog-backend/internal/utils"
	"blog-backend/internal/web"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	setupLogger(cfg)

	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("数据库连接失败", "error", err)
		os.Exit(1)
	}
	if cfg.SeedDemo {
		database.Seed(db)
	}

	router := setupRouter(db, cfg)
	srv := &http.Server{
		Addr:    "0.0.0.0:" + cfg.Port,
		Handler: router,
	}

	go func() {
		slog.Info("博客服务已启动", "addr", "http://"+srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("服务启动失败", "error", err)
			os.Exit(1)
		}
	}()

	// 优雅关停：等待 SIGINT/SIGTERM，10 秒内完成在途请求。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("收到退出信号，正在优雅关停...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("优雅关停出错", "error", err)
	}
	slog.Info("服务已退出")
}

// setupLogger 配置结构化日志（log/slog，JSON 格式）：同时输出到控制台与日志文件。
func setupLogger(cfg *config.Config) {
	if dir := filepath.Dir(cfg.LogFile); dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	level := slog.LevelInfo
	if os.Getenv("DEBUG") == "1" {
		level = slog.LevelDebug
	}
	writer := io.Writer(os.Stdout)
	if f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
		writer = io.MultiWriter(os.Stdout, f)
	} else {
		slog.Warn("日志文件打开失败，仅输出到控制台", "error", err)
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})))
}

// setupRouter 组装路由与中间件。
func setupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode) // 关闭 gin 自带调试日志，统一走 slog
	r := gin.New()
	r.Use(middleware.Recovery(), middleware.RequestLogger())

	// 前端应用（SPA）：/ 返回应用页面，/static/* 服务 css/js。
	appFS, err := fs.Sub(web.AppFS, "app")
	if err != nil {
		panic("加载前端资源失败: " + err.Error())
	}
	r.GET("/", func(c *gin.Context) {
		index, err := fs.ReadFile(appFS, "index.html")
		if err != nil {
			utils.HandleError(c, err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
	r.GET("/static/*filepath", gin.WrapH(http.StripPrefix("/static/", http.FileServer(http.FS(appFS)))))

	// API 文档页。
	r.GET("/api-doc", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(web.IndexHTML))
	})

	// 未知路径：统一返回 JSON 404（而不是 gin 默认文本）。
	r.NoRoute(func(c *gin.Context) {
		utils.HandleError(c, utils.NotFound("接口不存在: "+c.Request.Method+" "+c.Request.URL.Path))
	})

	auth := handlers.Auth{DB: db, Secret: cfg.JWTSecret}
	postH := handlers.Post{DB: db}
	commentH := handlers.Comment{DB: db}
	requireAuth := middleware.Auth(cfg.JWTSecret)

	api := r.Group("/api/v1")
	{
		api.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"time":   time.Now().Format(time.RFC3339),
			})
		})

		authG := api.Group("/auth")
		{
			authG.POST("/register", auth.Register)
			authG.POST("/login", auth.Login)
			authG.GET("/me", requireAuth, auth.Me)
		}

		posts := api.Group("/posts")
		{
			posts.GET("", postH.List)                 // 文章列表（公开）
			posts.GET("/:id", postH.Get)              // 文章详情（公开）
			posts.GET("/:id/comments", commentH.List) // 评论列表（公开）
			posts.POST("", requireAuth, postH.Create) // 创建文章（需认证）
			posts.PUT("/:id", requireAuth, postH.Update)
			posts.DELETE("/:id", requireAuth, postH.Delete)
			posts.POST("/:id/comments", requireAuth, commentH.Create) // 发表评论（需认证）
		}
	}
	return r
}
