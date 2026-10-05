// Package handler — HTTP-слой: принимает запрос, вызывает репозиторий, отдаёт JSON.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-server/internal/models"
	"go-server/internal/repository"
)

// PostHandler держит зависимость на интерфейс, а не на Postgres напрямую.
type PostHandler struct {
	repo repository.PostRepository
}

// NewPostHandler создаёт обработчик постов.
func NewPostHandler(repo repository.PostRepository) *PostHandler {
	return &PostHandler{repo: repo}
}

// ListPosts отдаёт все посты.
//
// Комментарии с @ — не просто документация для людей.
// Утилита `swag` читает их и собирает OpenAPI (то, что рисует Swagger UI).
// godoc в первой строке — соглашение: имя функции + короткое описание.
//
// ListPosts godoc
// @Summary      Список постов
// @Description  Возвращает все посты, новые сверху
// @Tags         posts
// @Produce      json
// @Success      200  {array}   models.Post
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/posts [get]
func (h *PostHandler) ListPosts(c *gin.Context) {
	// c.Request.Context() связан с жизнью HTTP-запроса:
	// если клиент отключился, запрос к БД тоже можно отменить.
	posts, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "не удалось получить посты"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// GetPost godoc
// @Summary      Получить пост
// @Description  Возвращает один пост по id
// @Tags         posts
// @Produce      json
// @Param        id   path      int  true  "ID поста"
// @Success      200  {object}  models.Post
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	post, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		writeRepoError(c, err)
		return
	}

	c.JSON(http.StatusOK, post)
}

// CreatePost godoc
// @Summary      Создать пост
// @Description  Создаёт новый пост
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post  body      models.CreatePostRequest  true  "Данные поста"
// @Success      201   {object}  models.Post
// @Failure      400   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /api/v1/posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req models.CreatePostRequest

	// ShouldBindJSON читает body и проверяет теги binding.
	// Если JSON кривой или нет обязательных полей — вернётся ошибка.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "некорректное тело запроса: нужны title и content"})
		return
	}

	post, err := h.repo.Create(c.Request.Context(), req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "не удалось создать пост"})
		return
	}

	// 201 Created — стандартный код для успешного создания ресурса.
	c.JSON(http.StatusCreated, post)
}

// UpdatePost godoc
// @Summary      Обновить пост
// @Description  Полностью заменяет title и content
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id    path      int                       true  "ID поста"
// @Param        post  body      models.UpdatePostRequest  true  "Новые данные"
// @Success      200   {object}  models.Post
// @Failure      400   {object}  models.ErrorResponse
// @Failure      404   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /api/v1/posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "некорректное тело запроса: нужны title и content"})
		return
	}

	post, err := h.repo.Update(c.Request.Context(), id, req.Title, req.Content)
	if err != nil {
		writeRepoError(c, err)
		return
	}

	c.JSON(http.StatusOK, post)
}

// DeletePost godoc
// @Summary      Удалить пост
// @Description  Удаляет пост по id
// @Tags         posts
// @Produce      json
// @Param        id   path  int  true  "ID поста"
// @Success      204  "Пост удалён"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		writeRepoError(c, err)
		return
	}

	// 204 No Content: успех, но тела ответа нет.
	c.Status(http.StatusNoContent)
}

// parseID достаёт :id из URL и превращает строку в int64.
// Второй результат (ok) — идиома Go: "получилось ли?".
func parseID(c *gin.Context) (int64, bool) {
	raw := c.Param("id")

	// strconv.ParseInt(строка, основание, битность).
	// 10 = десятичная система, 64 = int64.
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "id должен быть положительным числом"})
		return 0, false
	}

	return id, true
}

// writeRepoError переводит ошибки репозитория в HTTP-коды.
func writeRepoError(c *gin.Context, err error) {
	// errors.Is сравнивает цепочку обёрнутых ошибок с целевой.
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "пост не найден"})
		return
	}

	c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "внутренняя ошибка сервера"})
}
