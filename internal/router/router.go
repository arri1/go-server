// Package router собирает все HTTP-маршруты в одном месте.
package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go-server/internal/handler"

	// Пакет docs создаёт swag командой `swag init`.
	// Импорт с подчёркиванием регистрирует спецификацию при старте.
	_ "go-server/docs"
)

// New создаёт движок Gin и навешивает маршруты.
func New(posts *handler.PostHandler) *gin.Engine {
	// gin.New() — чистый движок без логгера и recovery.
	// gin.Default() добавляет лог запросов и защиту от паник — удобнее для учёбы.
	r := gin.Default()

	r.GET("/health", handler.Health)

	// Группа путей: общий префикс /api/v1 для всех CRUD-ручек.
	// Так проще потом сделать /api/v2, не ломая старых клиентов.
	v1 := r.Group("/api/v1")
	{
		v1.GET("/posts", posts.ListPosts)
		v1.GET("/posts/:id", posts.GetPost)
		v1.POST("/posts", posts.CreatePost)
		v1.PUT("/posts/:id", posts.UpdatePost)
		v1.DELETE("/posts/:id", posts.DeletePost)
	}

	// UI Swagger: http://localhost:8080/swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
