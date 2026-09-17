// Package handlers HTTP 处理器：认证、文章、评论。
package handlers

import (
	"errors"
	"log/slog"

	"blog-backend/internal/middleware"
	"blog-backend/internal/models"
	"blog-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Auth 认证相关处理器。
type Auth struct {
	DB     *gorm.DB
	Secret string // JWT 签名密钥
}

// Register POST /api/v1/auth/register 用户注册。
// 密码使用 bcrypt 加密后入库，绝不保存明文。
func (h *Auth) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, utils.BadRequest("参数校验失败: "+err.Error()))
		return
	}

	// 唯一性预检：给出友好提示（数据库唯一索引是最终防线）
	var count int64
	h.DB.Model(&models.User{}).Where("username = ? OR email = ?", req.Username, req.Email).Count(&count)
	if count > 0 {
		utils.HandleError(c, utils.Conflict("用户名或邮箱已被占用"))
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	user := models.User{Username: req.Username, Password: string(hashed), Email: req.Email}
	if err := h.DB.Create(&user).Error; err != nil {
		utils.HandleError(c, err)
		return
	}
	slog.Info("新用户注册", "id", user.ID, "username", user.Username)
	utils.Created(c, "注册成功", gin.H{"id": user.ID, "username": user.Username})
}

// Login POST /api/v1/auth/login 用户登录。
// 验证用户名与密码，成功后签发 24 小时有效的 JWT。
func (h *Auth) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, utils.BadRequest("参数校验失败: "+err.Error()))
		return
	}

	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.HandleError(c, utils.Unauthorized("用户名或密码错误"))
			return
		}
		utils.HandleError(c, err)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.HandleError(c, utils.Unauthorized("用户名或密码错误"))
		return
	}

	token, err := utils.GenerateToken(h.Secret, user.ID, user.Username)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	slog.Info("用户登录", "id", user.ID, "username", user.Username)
	utils.OK(c, models.LoginResponse{
		Token: token,
		User:  models.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email, CreatedAt: user.CreatedAt},
	})
}

// Me GET /api/v1/auth/me 当前登录用户信息（需认证）。
func (h *Auth) Me(c *gin.Context) {
	var user models.User
	if err := h.DB.First(&user, middleware.UserID(c)).Error; err != nil {
		utils.HandleError(c, err)
		return
	}
	utils.OK(c, models.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email, CreatedAt: user.CreatedAt})
}
