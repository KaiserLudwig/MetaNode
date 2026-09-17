// Package models 定义数据库模型（GORM）与接口请求/响应 DTO。
package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表：存储用户信息。
type User struct {
	gorm.Model
	Username string `gorm:"size:32;uniqueIndex;not null"`  // 用户名，唯一
	Password string `gorm:"size:128;not null"`             // bcrypt 密文，绝不存明文
	Email    string `gorm:"size:128;uniqueIndex;not null"` // 邮箱，唯一
	Posts    []Post
	Comments []Comment
}

// Post 文章表：属于某个用户，可被多次评论。
type Post struct {
	gorm.Model
	Title    string `gorm:"size:128;not null"`
	Content  string `gorm:"type:text;not null"`
	UserID   uint   `gorm:"index;not null"` // 关联 users.id
	User     User
	Comments []Comment
}

// Comment 评论表：属于某用户与某文章。
type Comment struct {
	gorm.Model
	Content string `gorm:"type:text;not null"`
	UserID  uint   `gorm:"index;not null"` // 关联 users.id
	User    User
	PostID  uint `gorm:"index;not null"` // 关联 posts.id
	Post    Post
}

// ---- 请求 DTO ----

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8,max=64"`
	Email    string `json:"email" binding:"required,email"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreatePostRequest 创建文章请求。
type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=128"`
	Content string `json:"content" binding:"required,min=1"`
}

// UpdatePostRequest 更新文章请求（指针字段：缺省表示不修改）。
type UpdatePostRequest struct {
	Title   *string `json:"title" binding:"omitempty,min=1,max=128"`
	Content *string `json:"content" binding:"omitempty,min=1"`
}

// CreateCommentRequest 发表评论请求。
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

// ---- 响应 DTO ----

// UserBrief 作者/用户简要信息。
type UserBrief struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// UserResponse 用户完整信息。
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// PostBrief 文章摘要（列表用）。
type PostBrief struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Author       UserBrief `json:"author"`
	CommentCount int64     `json:"comment_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// PostResponse 文章详情。
type PostResponse struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Author       UserBrief `json:"author"`
	CommentCount int64     `json:"comment_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CommentResponse 评论信息。
type CommentResponse struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	Author    UserBrief `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// PostListResponse 文章列表（分页）。
type PostListResponse struct {
	Items []PostBrief `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// LoginResponse 登录成功响应：JWT + 用户信息。
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
