// Package config читает настройки приложения из переменных окружения.
//
// В Go пакет = папка. Имя пакета обычно совпадает с именем папки.
// internal/ — особая папка: её можно импортировать только внутри этого модуля.
// Так мы прячем детали реализации от внешнего мира.
package config

import (
	"fmt"
	"os"
)

// Config — обычная структура (struct). Это набор полей с типами.
// В других языках это похоже на класс без методов (методы можно добавить отдельно).
type Config struct {
	// Port — порт, на котором слушает HTTP-сервер. Строка, потому что
	// gin.Run() принимает адрес вида ":8080".
	Port string

	// DatabaseURL — одна строка со всеми параметрами Postgres
	// (пользователь, пароль, хост, порт, имя БД).
	DatabaseURL string
}

// Load собирает Config из окружения.
//
// В Go функции часто возвращают (результат, error).
// Если error != nil — дальше результат обычно не используют.
func Load() (Config, error) {
	cfg := Config{
		// getenv возвращает значение или запасной вариант, если переменной нет.
		Port:        getenv("APP_PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/go_server?sslmode=disable"),
	}

	// Пустая строка подключения — явная ошибка, а не "потом разберёмся".
	if cfg.DatabaseURL == "" {
		// fmt.Errorf создаёт ошибку с текстом. %w не нужен, если мы не оборачиваем другую ошибку.
		return Config{}, fmt.Errorf("DATABASE_URL не задан")
	}

	return cfg, nil
}

// getenv — маленькая вспомогательная функция.
// Имена с маленькой буквы = неэкспортируемые (private) внутри пакета.
// С большой буквы (Load, Config) = экспортируемые (public), их видят другие пакеты.
func getenv(key, fallback string) string {
	// os.LookupEnv возвращает (значение, ok).
	// ok == false, если переменной нет в окружении.
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
