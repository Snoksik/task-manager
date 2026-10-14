package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

	if err := http.ListenAndServe("localhost:8080", handler.RequestIDMiddleware(mux)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
