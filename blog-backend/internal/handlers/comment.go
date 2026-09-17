package handlers

import (
	"blog-backend/internal/middleware"
	"blog-backend/internal/models"
	"blog-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Comment 评论相关处理器。
type Comment struct {
	DB *gorm.DB
}

// List GET /api/v1/posts/:id/comments 某篇文章的全部评论（含评论作者，按时间升序）。
func (h *Comment) List(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var post models.Post
	if err := h.DB.First(&post, id).Error; err != nil {
		utils.HandleError(c, utils.NotFound("文章不存在"))
		return
	}

	var comments []models.Comment
	if err := h.DB.Preload("User").
		Where("post_id = ?", id).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		utils.HandleError(c, err)
		return
	}
	items := make([]models.CommentResponse, 0, len(comments))
	for _, cm := range comments {
		items = append(items, models.CommentResponse{
			ID:        cm.ID,
			Content:   cm.Content,
			Author:    models.UserBrief{ID: cm.User.ID, Username: cm.User.Username},
			CreatedAt: cm.CreatedAt,
		})
	}
	utils.OK(c, gin.H{"post_id": id, "total": len(items), "items": items})
}

// Create POST /api/v1/posts/:id/comments 发表评论（需认证，文章必须存在）。
func (h *Comment) Create(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, utils.BadRequest("参数校验失败: "+err.Error()))
		return
	}

	var post models.Post
	if err := h.DB.First(&post, id).Error; err != nil {
		utils.HandleError(c, utils.NotFound("文章不存在"))
		return
	}
	comment := models.Comment{Content: req.Content, UserID: middleware.UserID(c), PostID: post.ID}
	if err := h.DB.Create(&comment).Error; err != nil {
		utils.HandleError(c, err)
		return
	}
	utils.Created(c, "评论成功", gin.H{"id": comment.ID})
}
