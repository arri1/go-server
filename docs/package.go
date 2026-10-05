// Package docs — OpenAPI-спецификация для Swagger UI.
//
// Файлы docs.go, swagger.json и swagger.yaml создаёт утилита swag
// из комментариев вида @Summary / @Router над хендлерами.
// Их не правят руками: после изменений аннотаций запускай `make swagger`.
package docs
