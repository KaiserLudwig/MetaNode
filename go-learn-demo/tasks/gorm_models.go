package tasks

import (
	"fmt"
)

// User 用户：与 Post 是一对多关系（一个用户可发布多篇文章）。
type User struct {
	ID        uint
	Name      string
	PostCount int // 文章数量统计（钩子演示用）
	Posts     []Post
}

// Post 文章：属于 User（belongsTo），拥有多个 Comment（hasMany）。
type Post struct {
	ID            uint
	Title         string
	UserID        uint
	User          User
	Comments      []Comment
	CommentStatus string // 评论状态："有评论" / "无评论"（钩子演示用）
}

// Comment 评论：属于 Post（belongsTo）。
type Comment struct {
	ID      uint
	Content string
	PostID  uint
	Post    Post
}

// RunGormModelDemo 任务11：模型定义 + AutoMigrate 建表。
func RunGormModelDemo() {
	db := newGormDB()

	fmt.Println("① 定义三个模型（关系由 GORM 标签自动推断）：")
	fmt.Println("   User    —— ID, Name, PostCount, Posts      (一对多: User → Post)")
	fmt.Println("   Post    —— ID, Title, UserID, CommentStatus, Comments")
	fmt.Println("                                                   (一对多: Post → Comment)")
	fmt.Println("   Comment —— ID, Content, PostID")

	fmt.Println("\n② 调用 AutoMigrate(&User{}, &Post{}, &Comment{}) 创建表...")
	if err := createBlogTables(db); err != nil {
		fmt.Println("建表失败:", err)
		return
	}

	fmt.Println("\n③ 数据库中的真实建表语句（GORM 生成的 DDL）：")
	for _, ddl := range tableDDL(db) {
		fmt.Println("  " + ddl)
	}

	fmt.Println("\n④ 建表完成：users / posts / comments 三张表已创建 ✓")
}
