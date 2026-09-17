// Package web 内嵌静态页面资源：API 文档页与前端应用。
package web

import (
	"embed"
)

// IndexHTML API 文档页（/api-doc）。
//
//go:embed index.html
var IndexHTML string

// AppFS 前端应用目录（index.html / app.css / app.js）。
//
//go:embed app
var AppFS embed.FS
