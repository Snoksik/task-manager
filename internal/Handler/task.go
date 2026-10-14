package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"

	"github.com/Snoksik/task-manager/internal/model"
)

type TaskService interface {
	GetAll(ctx context.Context, filter model.TaskFilter) ([]model.Task, error)
	GetByID(ctx context.Context, id int) (model.Task, error)
	Create(ctx context.Context, task model.Task) (model.Task, error)
	Delete(ctx context.Context, id int) error
	Update(ctx context.Context, id int, task model.Task) (model.Task, error)
}

type TaskHandler struct {
	service TaskService
}

func NewTaskHandler(service TaskService) *TaskHandler {
	return &TaskHandler{
		service: service,
	}
}
func parseQueryParams(r *http.Request) (model.TaskFilter, error) {
	query := r.URL.Query()
	search := query.Get("search")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		parseLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			return model.TaskFilter{}, err
		}
		limit = parseLimit
	}
	if offsetStr != "" {
		parseOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return model.TaskFilter{}, err
		}
		offset = parseOffset
	}

	filter := model.TaskFilter{
		Search: search,
		Limit:  limit,
		Offset: offset,
	}

	return filter, nil

}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromContext(r.Context())

	filter, err := parseQueryParams(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_TASK_FILTER", "Invalid task filter", requestID)
		return
	}

	tasks, err := h.service.GetAll(r.Context(), filter)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidTaskFilter):
			writeAPIError(w, http.StatusBadRequest, "INVALID_TASK_FILTER", "Invalid task filter", requestID)

		default:
			slog.Error("failed to get tasks",
				"request_id", requestID,
				"error", err)
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		slog.Error(
			"failed to encode tasks response",
			"request_id", requestID,
			"error", err)
	}

}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromContext(r.Context())
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeAPIError(w, http.StatusBadRequest, "INVALID_TASK_ID", "Invalid task ID", requestID)
		return
	}
	task, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrTaskNotFound) {
			writeAPIError(w, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found", requestID)
			return
		}
		slog.Error("failed to get task",
			"request_id", requestID,
			"error", err)
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		slog.Error(
			"failed to encode tasks response",
			"request_id", requestID,
			"error", err)
	}
}

func jsonHelper(w http.ResponseWriter, r *http.Request) (model.CreateTaskRequest, error) {

	var body model.CreateTaskRequest

	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)) //лимит на длинну json (http.MaxBytesReader(w, r.Body, 4096))
	d.DisallowUnknownFields()                                  //очень классная тема - жесткое декодирование и запрет доп данных
	err := d.Decode(&body)
	if err != nil {
		return model.CreateTaskRequest{}, err
	}
	if body.UserID == nil {
		return model.CreateTaskRequest{}, errors.New("user_id is required")
	}
	var extra any
	err = d.Decode(&extra)
	if !errors.Is(err, io.EOF) {
		return model.CreateTaskRequest{}, errors.New("need 1 json, dont more")
	}
	return body, nil
}

func headerCheck(r *http.Request) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")) // парсирование Header на составные части
	if err != nil {
		return errors.New("expected application/json, got another")
	}
	if mediaType != "application/json" {
		return errors.New("expected application/json, got another")
	}
	return nil
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromContext(r.Context())
	err := headerCheck(r)
	if err != nil {
		writeAPIError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json", requestID)
		return
	}
	body, err := jsonHelper(w, r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body", requestID)
		return
	}
	bodyTask := model.Task{
		Name:   body.Name,
		UserID: *body.UserID,
	}
	task, err := h.service.Create(r.Context(), bodyTask)
	if err != nil {
		if errors.Is(err, model.ErrInvalidTaskName) {
			writeAPIError(w, http.StatusBadRequest, "INVALID_TASK_NAME", "Invalid task name", requestID)
			return
		}

		slog.Error(
			"failed to create task",
			"request_id", requestID,
			"error", err)
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		slog.Error(
			"failed to encode tasks response",
			"request_id", requestID,
			"error", err)
	}
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromContext(r.Context())

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeAPIError(w, http.StatusBadRequest, "INVALID_TASK_ID", "Invalid task ID", requestID)
		return
	}

	if err = h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, model.ErrTaskNotFound) {
			writeAPIError(w, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found", requestID)
			return
		}
		slog.Error(
			"failed to delete tasks response",
			"request_id", requestID,
			"error", err)
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromContext(r.Context())

	err := headerCheck(r)
	if err != nil {
		writeAPIError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json", requestID)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeAPIError(w, http.StatusBadRequest, "INVALID_TASK_ID", "Invalid task ID", requestID)
		return
	}

	body, err := jsonHelper(w, r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body", requestID)
		return
	}
	newTask := model.Task{
		Name:   body.Name,
		UserID: *body.UserID,
	}
	task, err := h.service.Update(r.Context(), id, newTask)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidTaskName):
			writeAPIError(w, http.StatusBadRequest, "INVALID_TASK_NAME", "Invalid task name", requestID)
		case errors.Is(err, model.ErrTaskNotFound):
			writeAPIError(w, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found", requestID)
		default:
			slog.Error(
				"failed to update task",
				"request_id", requestID,
				"error", err)
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestID)
		}
		return
	}
	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(task); err != nil {
		slog.Error(
			"failed to encode tasks response",
			"request_id", requestID,
			"error", err)
	}

}
