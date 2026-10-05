// Package database отвечает только за подключение к Postgres.
// SQL-запросы живут в repository — так проще читать и тестировать.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Анонимный импорт: пакет нужен ради побочного эффекта (регистрации драйвера),
	// а не ради типов. Подчёркивание `_` говорит компилятору: "импортируй, но не используй имя".
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect открывает пул соединений к Postgres и проверяет, что БД жива.
//
// *sql.DB — это не одно соединение, а пул. Его обычно создают один раз
// при старте приложения и передают дальше.
func Connect(ctx context.Context, databaseURL string) (*sql.DB, error) {
	// sql.Open ещё не ходит в сеть — только готовит пул с указанным драйвером.
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		// %w оборачивает исходную ошибку: её потом можно достать через errors.Is / errors.As.
		return nil, fmt.Errorf("открыть пул соединений: %w", err)
	}

	// Небольшие лимиты, чтобы не держать слишком много соединений к Postgres.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	// PingContext реально проверяет сеть и авторизацию.
	// context.Context позволяет отменить ожидание по таймауту.
	if err := db.PingContext(ctx); err != nil {
		// Закрываем пул, если пинг не прошёл — иначе утечёт ресурс.
		_ = db.Close()
		return nil, fmt.Errorf("проверить соединение с postgres: %w", err)
	}

	return db, nil
}

// Migrate создаёт таблицу posts, если её ещё нет.
//
// Для учебного проекта этого достаточно. В серьёзном сервисе обычно
// берут golang-migrate и хранят историю версий схемы отдельно.
func Migrate(ctx context.Context, db *sql.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS posts (
			id         BIGSERIAL PRIMARY KEY,
			title      TEXT        NOT NULL,
			content    TEXT        NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`

	// ExecContext выполняет SQL без возврата строк (DDL, INSERT без RETURNING и т.п.).
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("применить миграцию posts: %w", err)
	}

	return nil
}
