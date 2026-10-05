// Package main — точка входа программы.
// В Go исполняемый файл обязан называться package main и иметь func main().
//
// @title           Go Server API
// @version         1.0
// @description     Учебный REST API: CRUD постов, Postgres и Swagger.
// @host            localhost:8080
// @BasePath        /
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-server/internal/config"
	"go-server/internal/database"
	"go-server/internal/handler"
	"go-server/internal/repository"
	"go-server/internal/router"
)

func main() {
	// slog — стандартный структурированный логгер (с Go 1.21).
	// JSON удобнее читать в Docker, чем обычный fmt.Println.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		// slog.Error не останавливает программу. Для фатальной ошибки выходим сами.
		slog.Error("не удалось загрузить конфиг", "err", err)
		os.Exit(1)
	}

	// context.Background() — корневой контекст без дедлайна.
	// WithTimeout делает дочерний: через 15 секунд операции с этим ctx отменятся.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	cancel() // освобождаем таймер сразу после Connect
	if err != nil {
		slog.Error("не удалось подключиться к postgres", "err", err)
		os.Exit(1)
	}
	// defer выполнится при выходе из main: закроем пул соединений.
	defer db.Close()

	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := database.Migrate(migrateCtx, db); err != nil {
		migrateCancel()
		slog.Error("не удалось применить миграции", "err", err)
		os.Exit(1)
	}
	migrateCancel()

	// Собираем слои снизу вверх: БД → репозиторий → handler → router.
	posts := handler.NewPostHandler(repository.NewPostgresPostRepository(db))
	engine := router.New(posts)

	// http.Server даёт ручное управление: ListenAndServe + Shutdown.
	// gin.Engine.Run() так аккуратно остановиться не умеет.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Сервер запускаем в отдельной горутине (лёгкий поток Go).
	// main при этом может ждать сигнал ОС.
	errCh := make(chan error, 1)
	go func() {
		slog.Info("сервер слушает", "addr", srv.Addr)
		// ListenAndServe всегда возвращает ошибку: либо реальную, либо http.ErrServerClosed после Shutdown.
		errCh <- srv.ListenAndServe()
	}()

	// signal.NotifyContext завершит контекст на Ctrl+C или docker stop (SIGTERM).
	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-runCtx.Done():
		slog.Info("получен сигнал остановки, гасим сервер")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("сервер упал", "err", err)
			os.Exit(1)
		}
		return
	}

	// Даём активным запросам до 10 секунд завершиться.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("ошибка graceful shutdown", "err", err)
		os.Exit(1)
	}

	slog.Info("сервер остановлен")
}
