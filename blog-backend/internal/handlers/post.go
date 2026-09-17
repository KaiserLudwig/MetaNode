package handlers

import (
	"strconv"

	"blog-backend/internal/middleware"
	"blog-backend/internal/models"
	"blog-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Post 文章相关处理器。
type Post struct {
	DB *gorm.DB
}

// parseID 解析路径中的文章 ID。
func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.HandleError(c, utils.BadRequest("无效的 ID: "+c.Param("id")))
		return 0, false
	}
	return uint(id), true
}

// List GET /api/v1/posts 文章列表（分页 + 作者 + 评论数）。
func (h *Post) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	var total int64
	h.DB.Model(&models.Post{}).Count(&total)

	var posts []models.Post
	if err := h.DB.Preload("User").
		Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&posts).Error; err != nil {
		utils.HandleError(c, err)
		return
	}
	items := make([]models.PostBrief, 0, len(posts))
	for _, p := range posts {
		items = append(items, toPostBrief(h.DB, p))
	}
	utils.OK(c, models.PostListResponse{Items: items, Total: total, Page: page, Size: size})
}

// Get GET /api/v1/posts/:id 文章详情（含作者与评论数）。
func (h *Post) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var post models.Post
	if err := h.DB.Preload("User").First(&post, id).Error; err != nil {
		utils.HandleError(c, utils.NotFound("文章不存在"))
		return
	}
	var cnt int64
	h.DB.Model(&models.Comment{}).Where("post_id = ?", post.ID).Count(&cnt)
	utils.OK(c, models.PostResponse{
		ID:           post.ID,
		Title:        post.Title,
		Content:      post.Content,
		Author:       models.UserBrief{ID: post.User.ID, Username: post.User.Username},
		CommentCount: cnt,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	})
}

// Create POST /api/v1/posts 创建文章（需认证）。
func (h *Post) Create(c *gin.Context) {
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, utils.BadRequest("参数校验失败: "+err.Error()))
		return
	}
	post := models.Post{Title: req.Title, Content: req.Content, UserID: middleware.UserID(c)}
	if err := h.DB.Create(&post).Error; err != nil {
		utils.HandleError(c, err)
		return
	}
	utils.Created(c, "文章创建成功", gin.H{"id": post.ID, "title": post.Title})
}

// Update PUT /api/v1/posts/:id 更新文章（仅作者本人，支持部分字段更新）。
func (h *Post) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, utils.BadRequest("参数校验失败: "+err.Error()))
		return
	}

	var post models.Post
	if err := h.DB.First(&post, id).Error; err != nil {
		utils.HandleError(c, utils.NotFound("文章不存在"))
		return
	}
	if post.UserID != middleware.UserID(c) {
		utils.HandleError(c, utils.Forbidden("只能修改自己发布的文章"))
		return
	}

	updates := map[string]any{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if len(updates) == 0 {
		utils.HandleError(c, utils.BadRequest("没有需要更新的字段"))
		return
	}
	if err := h.DB.Model(&post).Updates(updates).Error; err != nil {
		utils.HandleError(c, err)
		return
	}
	utils.OK(c, gin.H{"message": "文章更新成功", "id": post.ID})
}

// Delete DELETE /api/v1/posts/:id 删除文章（仅作者本人，事务内级联删除评论）。
func (h *Post) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var post models.Post
	if err := h.DB.First(&post, id).Error; err != nil {
		utils.HandleError(c, utils.NotFound("文章不存在"))
		return
	}
	if post.UserID != middleware.UserID(c) {
		utils.HandleError(c, utils.Forbidden("只能删除自己发布的文章"))
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", post.ID).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&post).Error
	})
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	utils.OK(c, gin.H{"message": "文章删除成功", "id": post.ID})
}

// toPostBrief 组装文章摘要（列表项），附带评论数。
func toPostBrief(db *gorm.DB, p models.Post) models.PostBrief {
	var cnt int64
	db.Model(&models.Comment{}).Where("post_id = ?", p.ID).Count(&cnt)
	return models.PostBrief{
		ID:           p.ID,
		Title:        p.Title,
		Author:       models.UserBrief{ID: p.User.ID, Username: p.User.Username},
		CommentCount: cnt,
		CreatedAt:    p.CreatedAt,
	}
}
