package tasks

import (
	"fmt"
)

// RunGormQueryDemo 任务12：关联查询。
//
// ① Preload 预加载：查询某个用户的所有文章及其评论。
// ② 聚合查询：评论数量最多的文章。
func RunGormQueryDemo() {
	db := newGormDB()
	if err := createBlogTables(db); err != nil {
		fmt.Println("建表失败:", err)
		return
	}
	seedBlog(db)
	fmt.Println("演示数据：张三 2 篇文章（3+1 条评论），李四 1 篇文章（2 条评论）")

	fmt.Println("\n① 查询用户“张三”的所有文章及其评论（Preload 预加载）:")
	var zhang User
	if err := db.Preload("Posts.Comments").First(&zhang, "name = ?", "张三").Error; err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	fmt.Printf("   用户: %s\n", zhang.Name)
	for _, p := range zhang.Posts {
		fmt.Printf("     ├─ 文章《%s》 共 %d 条评论\n", p.Title, len(p.Comments))
		for _, c := range p.Comments {
			fmt.Printf("     │    └─ 评论: %s\n", c.Content)
		}
	}

	fmt.Println("\n② 查询评论数量最多的文章（聚合: COUNT + GROUP BY）:")
	type topPost struct {
		Title string
		Cnt   int
	}
	var top topPost
	err := db.Model(&Post{}).
		Select("posts.title AS title, COUNT(comments.id) AS cnt").
		Joins("LEFT JOIN comments ON comments.post_id = posts.id").
		Group("posts.id").
		Order("cnt DESC").
		Limit(1).
		Scan(&top).Error
	if err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	fmt.Printf("   🏆 《%s》 评论数最多，共 %d 条\n", top.Title, top.Cnt)
}
