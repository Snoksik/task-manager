package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	repository "github.com/Snoksik/task-manager/internal/Repository"
	"github.com/Snoksik/task-manager/internal/model"
	"github.com/Snoksik/task-manager/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTaskCRUDIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	parsedURL, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("invalid test database URL: %v", err)
	}

	if parsedURL.Path != "/task_manager_test" {
		t.Fatal("integration test must use task_manager_test")
	}
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)

	if err != nil {
		t.Fatalf("failed to create database pool: %v", err)
	}

	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	repo := repository.NewPostgresRepository(pool)
	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks", h.GetAll)
	mux.HandleFunc("GET /tasks/{id}", h.GetByID)
	mux.HandleFunc("POST /tasks", h.Create)
	mux.HandleFunc("PUT /tasks/{id}", h.Update)
	mux.HandleFunc("DELETE /tasks/{id}", h.Delete)

	wrapped := RequestIDMiddleware(mux)

	body := `{"name":"Integration test task","user_id":1}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/tasks",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status created, got %d with body %v", rec.Code, rec.Body.String())
	}

	var resPost model.Task
	err = json.NewDecoder(rec.Body).Decode(&resPost)
	if err != nil {
		t.Fatalf("expected zero errors in decode, got %v", err)

	}
	if resPost.ID <= 0 {
		t.Fatalf("expected id > 0, got %d", resPost.ID)
	}

	t.Cleanup(func() {
		_, err := pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", resPost.ID)
		if err != nil {
			t.Errorf("Expected delete task in clean, got %v", err)
		}
	})

	path := "/tasks/" + strconv.Itoa(resPost.ID)

	req2 := httptest.NewRequest(http.MethodGet, path, nil)
	rec2 := httptest.NewRecorder()

	wrapped.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected status ok, got %d with body %v", rec2.Code, rec2.Body.String())
	}

	var resGetById model.Task
	err = json.NewDecoder(rec2.Body).Decode(&resGetById)
	if err != nil {
		t.Fatalf("expected zero errors in decode, got %v", err)
	}
	if resGetById.Name != "Integration test task" {
		t.Fatalf("unexpected task name: %q", resGetById.Name)
	}

	if resPost.UserID != resGetById.UserID {
		t.Fatalf("expected equal user_id, got not %v and %v", resPost.UserID, resGetById.UserID)
	}

	if resPost.Name != resGetById.Name {
		t.Fatalf("expected equal Name, got not %v and %v", resPost.Name, resGetById.Name)
	}

	body1 := `{"name":"Updated integration task","user_id":1}`

	req3 := httptest.NewRequest(
		http.MethodPut,
		path,
		strings.NewReader(body1),
	)

	req3.Header.Set("Content-Type", "application/json")
	rec3 := httptest.NewRecorder()

	wrapped.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusOK {
		t.Fatalf("expected status OK, got %d with body %v", rec3.Code, rec3.Body.String())
	}

	var resPut model.Task
	err = json.NewDecoder(rec3.Body).Decode(&resPut)
	if err != nil {
		t.Fatalf("expected zero errors in decode, got %v", err)

	}
	if resPut.ID != resPost.ID {
		t.Fatalf("expected equal id, got %d and %d", resPut.ID, resPost.ID)
	}
	if resPut.UserID != resPost.UserID {
		t.Fatalf("expectet equal user_id, got not %v and %v", resPut.UserID, resPost.UserID)
	}
	if resPut.Name != "Updated integration task" {
		t.Fatalf("expected  Updated integration task, got %v", resPut.Name)
	}

	req4 := httptest.NewRequest(http.MethodGet, path, nil)
	rec4 := httptest.NewRecorder()

	wrapped.ServeHTTP(rec4, req4)

	if rec4.Code != http.StatusOK {
		t.Fatalf("expected status ok, got %d with body %v", rec4.Code, rec4.Body.String())
	}

	var resGetById2 model.Task
	err = json.NewDecoder(rec4.Body).Decode(&resGetById2)
	if err != nil {
		t.Fatalf("expected zero errors in decode, got %v", err)
	}
	if resGetById2.ID != resPost.ID {
		t.Fatalf("expected task ID %d, got %d", resPost.ID, resGetById2.ID)
	}
	if resGetById2.Name != "Updated integration task" {
		t.Fatalf("unexpected task name: %q", resGetById2.Name)
	}

	if resPut.UserID != resGetById2.UserID {
		t.Fatalf("expected equal user_id, got not %v and %v", resPut.UserID, resGetById2.UserID)
	}

	if resPut.Name != resGetById2.Name {
		t.Fatalf("expected equal Name, got not %v and %v", resPut.Name, resGetById2.Name)
	}

	req5 := httptest.NewRequest(http.MethodDelete, path, nil)
	rec5 := httptest.NewRecorder()

	wrapped.ServeHTTP(rec5, req5)

	if rec5.Code != http.StatusNoContent {
		t.Fatalf("expected status ok, got %d with body %v", rec5.Code, rec5.Body.String())
	}
	if rec5.Body.Len() != 0 {
		t.Fatalf("expected empty DELETE response, got %q", rec5.Body.String())
	}

	req6 := httptest.NewRequest(http.MethodGet, path, nil)
	rec6 := httptest.NewRecorder()

	wrapped.ServeHTTP(rec6, req6)

	if rec6.Code != http.StatusNotFound {
		t.Fatalf("expected status not found, got %d with body %v", rec6.Code, rec6.Body.String())
	}

	var errorResponse ErrorResponse

	if err := json.NewDecoder(rec6.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errorResponse.Error.Code != "TASK_NOT_FOUND" {
		t.Fatalf(
			"expected TASK_NOT_FOUND, got %q",
			errorResponse.Error.Code,
		)
	}
	if rec5.Body.Len() != 0 {
		t.Fatalf("expected empty DELETE response, got %q", rec5.Body.String())
	}
	requestID := rec6.Header().Get("X-Request-ID")

	if requestID == "" {
		t.Fatal("expected X-Request-ID header")
	}

	if errorResponse.RequestID != requestID {
		t.Fatalf("request IDs do not match")
	}
}
