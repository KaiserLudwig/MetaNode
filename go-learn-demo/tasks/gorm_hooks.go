package tasks

import (
	"fmt"

	"gorm.io/gorm"
)

// AfterCreate 钩子（Post）：文章创建成功后，自动把作者的文章统计 +1。
// 钩子运行在 GORM 事务内，失败会回滚整个创建操作。
func (p *Post) AfterCreate(tx *gorm.DB) error {
	return tx.Model(&User{}).
		Where("id = ?", p.UserID).
		UpdateColumn("post_count", gorm.Expr("post_count + 1")).Error
}

// AfterDelete 钩子（Comment）：评论删除后，检查该文章剩余评论数，
// 若为 0 则把文章状态更新为"无评论"。
func (c *Comment) AfterDelete(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return tx.Model(&Post{}).
			Where("id = ?", c.PostID).
			Update("comment_status", "无评论").Error
	}
	return nil
}

// RunGormHooksDemo 任务13：钩子函数演示。
func RunGormHooksDemo() {
	db := newGormDB()
	if err := createBlogTables(db); err != nil {
		fmt.Println("建表失败:", err)
		return
	}
	seedBlog(db)

	var zhang User
	db.First(&zhang, "name = ?", "张三")
	fmt.Println("① 种子数据：张三已有 2 篇文章（AfterCreate 钩子已自动维护 PostCount）")
	fmt.Printf("   张三 PostCount = %d ✓\n", zhang.PostCount)

	fmt.Println("\n② 张三再发布 1 篇文章，AfterCreate 钩子自动触发：")
	newPost := Post{Title: "GORM 钩子实战", UserID: zhang.ID, CommentStatus: "有评论"}
	if err := db.Create(&newPost).Error; err != nil {
		fmt.Println("创建失败:", err)
		return
	}
	db.First(&zhang, "name = ?", "张三")
	fmt.Printf("   新文章《%s》创建成功，张三 PostCount 自动变为 %d ✓\n", newPost.Title, zhang.PostCount)

	fmt.Println("\n③ 给文章添加 2 条评论，然后逐条删除：")
	c1 := Comment{Content: "钩子真有用", PostID: newPost.ID}
	c2 := Comment{Content: "学到了", PostID: newPost.ID}
	db.Create(&c1)
	db.Create(&c2)
	showPostStatus(db, newPost.ID)

	fmt.Println("   删除第 1 条评论（还剩 1 条，状态不变）...")
	db.Delete(&c1)
	showPostStatus(db, newPost.ID)

	fmt.Println("   删除最后 1 条评论，AfterDelete 钩子触发：")
	db.Delete(&c2)
	showPostStatus(db, newPost.ID)

	fmt.Println("\n④ 结论：评论删光后，文章状态被钩子自动更新为“无评论” ✓")
}

// showPostStatus 打印文章的评论数与状态。
func showPostStatus(db *gorm.DB, postID uint) {
	var p Post
	var cnt int64
	db.First(&p, postID)
	db.Model(&Comment{}).Where("post_id = ?", postID).Count(&cnt)
	fmt.Printf("   《%s》评论数=%d 状态=%s\n", p.Title, cnt, p.CommentStatus)
}
