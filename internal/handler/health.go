package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthResponse — простой ответ для проверки, что процесс жив.
// Docker и балансировщики дергают /health, чтобы понять: контейнер ещё отвечает.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// Health godoc
// @Summary      Проверка живости
// @Description  Используется Docker healthcheck и балансировщиками
// @Tags         system
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
