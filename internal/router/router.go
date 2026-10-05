// Package router собирает все HTTP-маршруты в одном месте.
package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go-server/docs"
	"go-server/internal/handler"
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

	// Host в сгенерированных docs зашит как localhost:8080.
	// Подменяем на Host текущего запроса: локально останется localhost,
	// на VPS Try it out пойдёт на 185.233.185.109:8080, а не на машину клиента.
	r.GET("/swagger/*any", func(c *gin.Context) {
		docs.SwaggerInfo.Host = c.Request.Host
		if c.Request.TLS != nil {
			docs.SwaggerInfo.Schemes = []string{"https"}
		} else {
			docs.SwaggerInfo.Schemes = []string{"http"}
		}
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
	})

	return r
}
