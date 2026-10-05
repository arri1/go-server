// Package repository — слой доступа к данным.
// Handlers не пишут SQL сами: они вызывают методы репозитория.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go-server/internal/models"
)

// ErrNotFound — сторожевая ошибка ("sentinel error").
// Её сравнивают через errors.Is, а не по тексту строки.
//
// var + errors.New — обычный способ завести такую ошибку на уровне пакета.
var ErrNotFound = errors.New("не найдено")

// PostRepository описывает операции с постами.
//
// Это интерфейс: набор методов без реализации.
// Handler зависит от интерфейса, а не от конкретной структуры.
// Так проще подменить реализацию в тестах (мок вместо Postgres).
type PostRepository interface {
	Create(ctx context.Context, title, content string) (models.Post, error)
	GetByID(ctx context.Context, id int64) (models.Post, error)
	List(ctx context.Context) ([]models.Post, error)
	Update(ctx context.Context, id int64, title, content string) (models.Post, error)
	Delete(ctx context.Context, id int64) error
}

// PostgresPostRepository — реализация PostRepository через database/sql.
//
// Встраивать *sql.DB в структуру — частый приём: методы получают доступ к БД.
type PostgresPostRepository struct {
	db *sql.DB
}

// NewPostgresPostRepository — конструктор.
// В Go нет ключевых слов constructor/new как в Java, поэтому пишут функции New....
func NewPostgresPostRepository(db *sql.DB) *PostgresPostRepository {
	// Возвращаем указатель (*T), чтобы не копировать структуру при каждом вызове.
	return &PostgresPostRepository{db: db}
}

// Create вставляет пост и сразу читает его обратно через RETURNING.
func (r *PostgresPostRepository) Create(ctx context.Context, title, content string) (models.Post, error) {
	const query = `
		INSERT INTO posts (title, content)
		VALUES ($1, $2)
		RETURNING id, title, content, created_at, updated_at
	`

	var post models.Post
	// QueryRowContext ждёт ровно одну строку.
	// Scan раскладывает колонки по полям структуры по порядку.
	err := r.db.QueryRowContext(ctx, query, title, content).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return models.Post{}, fmt.Errorf("создать пост: %w", err)
	}

	return post, nil
}

// GetByID ищет пост по первичному ключу.
func (r *PostgresPostRepository) GetByID(ctx context.Context, id int64) (models.Post, error) {
	const query = `
		SELECT id, title, content, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	var post models.Post
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		// sql.ErrNoRows — стандартная ошибка "запрос не вернул строк".
		if errors.Is(err, sql.ErrNoRows) {
			return models.Post{}, ErrNotFound
		}
		return models.Post{}, fmt.Errorf("получить пост %d: %w", id, err)
	}

	return post, nil
}

// List возвращает все посты, новые сверху.
func (r *PostgresPostRepository) List(ctx context.Context) ([]models.Post, error) {
	const query = `
		SELECT id, title, content, created_at, updated_at
		FROM posts
		ORDER BY id DESC
	`

	// QueryContext возвращает курсор (*sql.Rows), его обязательно закрывать.
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("список постов: %w", err)
	}
	// defer выполнит Close() при выходе из функции — даже если дальше будет return с ошибкой.
	defer rows.Close()

	// Пустой слайс (не nil), чтобы JSON стал [] а не null.
	posts := make([]models.Post, 0)

	// rows.Next() двигает курсор. Когда строки кончились — вернёт false.
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("прочитать строку поста: %w", err)
		}
		// append добавляет элемент в слайс (при необходимости выделяет новую память).
		posts = append(posts, post)
	}

	// После цикла проверяем ошибку итерации (обрыв соединения и т.п.).
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("обойти посты: %w", err)
	}

	return posts, nil
}

// Update меняет title/content и обновляет updated_at на стороне БД.
func (r *PostgresPostRepository) Update(ctx context.Context, id int64, title, content string) (models.Post, error) {
	const query = `
		UPDATE posts
		SET title = $1, content = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, title, content, created_at, updated_at
	`

	var post models.Post
	err := r.db.QueryRowContext(ctx, query, title, content, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Post{}, ErrNotFound
		}
		return models.Post{}, fmt.Errorf("обновить пост %d: %w", id, err)
	}

	return post, nil
}

// Delete удаляет пост. Если строка не найдена — тоже ErrNotFound.
func (r *PostgresPostRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM posts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("удалить пост %d: %w", id, err)
	}

	// RowsAffected говорит, сколько строк задел запрос.
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("узнать число удалённых строк: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}
