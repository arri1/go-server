// Package models хранит структуры данных, которыми оперирует API.
// Их же используем в JSON и в Swagger.
package models

import "time"

// Post — запись в блоге / новостях.
//
// Теги в обратных кавычках (`json:"..."`) — это метаданные для библиотек.
// encoding/json смотрит на json-тег и понимает, как назвать поле в JSON.
// Пример: поле CreatedAt станет "created_at" в ответе API.
type Post struct {
	ID        int64     `json:"id" example:"1"`
	Title     string    `json:"title" example:"Привет, Go"`
	Content   string    `json:"content" example:"Текст поста"`
	CreatedAt time.Time `json:"created_at" example:"2026-10-05T11:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-10-05T11:00:00Z"`
}

// CreatePostRequest — тело POST /posts.
// Отдельную структуру для запроса делают специально:
// клиент не должен присылать id и даты, их выставляет сервер.
type CreatePostRequest struct {
	// binding:"required" — тег Gin: поле обязательно, иначе 400.
	Title   string `json:"title" binding:"required" example:"Привет, Go"`
	Content string `json:"content" binding:"required" example:"Текст поста"`
}

// UpdatePostRequest — тело PUT /posts/:id.
// Те же поля, но отдельный тип, чтобы потом легко добавить частичное обновление.
type UpdatePostRequest struct {
	Title   string `json:"title" binding:"required" example:"Обновлённый заголовок"`
	Content string `json:"content" binding:"required" example:"Обновлённый текст"`
}

// ErrorResponse — единый формат ошибки для API и Swagger.
type ErrorResponse struct {
	Error string `json:"error" example:"пост не найден"`
}
