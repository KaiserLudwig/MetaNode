package tasks

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newGormDB 创建内存 SQLite 数据库（每次运行全新数据，无副作用）。
// cache=shared 保证连接池内所有连接共享同一个内存库。
func newGormDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 关闭 GORM 日志，输出由演示代码控制
	})
	if err != nil {
		panic("打开内存数据库失败: " + err.Error())
	}
	return db
}

// createBlogTables 创建博客系统三张表。
func createBlogTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Post{}, &Comment{})
}

// seedBlog 插入演示数据：
// 张三：2 篇文章（第 1 篇 3 条评论、第 2 篇 1 条评论）
// 李四：1 篇文章（2 条评论）
func seedBlog(db *gorm.DB) {
	zhang := User{Name: "张三"}
	li := User{Name: "李四"}
	if err := db.Create(&zhang).Error; err != nil {
		panic(err)
	}
	if err := db.Create(&li).Error; err != nil {
		panic(err)
	}
	posts := []Post{
		{Title: "Go 并发模型入门", UserID: zhang.ID, CommentStatus: "有评论"},
		{Title: "GORM 使用技巧", UserID: zhang.ID, CommentStatus: "有评论"},
		{Title: "区块链基础", UserID: li.ID, CommentStatus: "有评论"},
	}
	if err := db.Create(&posts).Error; err != nil {
		panic(err)
	}
	comments := []Comment{
		{Content: "写得太好了", PostID: posts[0].ID},
		{Content: "收藏了", PostID: posts[0].ID},
		{Content: "期待续篇", PostID: posts[0].ID},
		{Content: "学到了", PostID: posts[1].ID},
		{Content: "讲得清楚", PostID: posts[2].ID},
		{Content: "赞一个", PostID: posts[2].ID},
	}
	if err := db.Create(&comments).Error; err != nil {
		panic(err)
	}
}

// tableDDL 读取数据库中的建表语句（真实 DDL）。
func tableDDL(db *gorm.DB) []string {
	type master struct {
		SQL string
	}
	var rows []master
	if err := db.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name").Scan(&rows).Error; err != nil {
		return nil
	}
	var out []string
	for _, r := range rows {
		out = append(out, r.SQL)
	}
	return out
}

// 统计表行数（用于演示输出）。
func tableCounts(db *gorm.DB) map[string]int64 {
	counts := map[string]int64{}
	for _, t := range []string{"users", "posts", "comments"} {
		var n int64
		if err := db.Table(t).Count(&n).Error; err == nil {
			counts[t] = n
		}
	}
	return counts
}
