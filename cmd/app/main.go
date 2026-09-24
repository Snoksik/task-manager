package main

import (
	"log"
	"net/http"

	handler "github.com/Snoksik/task-manager/internal/Handler"
	repository "github.com/Snoksik/task-manager/internal/Repository"
	"github.com/Snoksik/task-manager/internal/service"
)

func main() {
	repo := repository.NewTaskRepository()
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks", h.GetAll)
	mux.HandleFunc("GET /tasks/{id}", h.GetByID)
	mux.HandleFunc("POST /tasks", h.Create)
	mux.HandleFunc("DELETE /tasks/{id}", h.Delete)
	mux.HandleFunc("PUT /tasks/{id}", h.Update)

	log.Println("starting server on http://localhost:8080")

	if err := http.ListenAndServe("localhost:8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
