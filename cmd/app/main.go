package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	handler "github.com/Snoksik/task-manager/internal/Handler"
	repository "github.com/Snoksik/task-manager/internal/Repository"
	"github.com/Snoksik/task-manager/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)

	if err != nil {
		log.Fatalf("failed to create database pool: %v", err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	repo := repository.NewPostgresRepository(pool)
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks", h.GetAll)
	mux.HandleFunc("GET /tasks/{id}", h.GetByID)
	mux.HandleFunc("POST /tasks", h.Create)
	mux.HandleFunc("DELETE /tasks/{id}", h.Delete)
	mux.HandleFunc("PUT /tasks/{id}", h.Update)

	log.Println("starting server on http://localhost:8080")

	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: handler.RequestIDMiddleware(mux),

		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverError := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				serverError <- err
			}
			// slog.Debug("Server Closed")
		}
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-signalCtx.Done():
		slog.Info("shutting down server")
	case err := <-serverError:
		slog.Error("server stopped unexpectedly", "error", err)
		return
	}

	shutdownctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	err = srv.Shutdown(shutdownctx)
	if err != nil {
		slog.Error("failed to shutdown server", "error", err)
		return
	}

	slog.Info("server stopped successfully")
}
