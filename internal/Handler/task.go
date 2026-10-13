package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
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
	filter, err := parseQueryParams(r)
	if err != nil {
		http.Error(w, model.ErrInvalidTaskFilter.Error(), 400)
		return
	}

	tasks, err := h.service.GetAll(r.Context(), filter)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidTaskFilter):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			log.Printf("failed to get tasks: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("failed to encode tasks response: %v", err)
	}

}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}
	task, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrTaskNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Printf("failed to get by id tasks response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("failed to encode tasks response: %v", err)
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
func headerCheckInReader(w http.ResponseWriter) error {
	mediaType, _, err := mime.ParseMediaType(w.Header().Get("Content-Type")) // парсирование Header на составные части
	if err != nil {
		return errors.New("expected application/json, got another")
	}
	if mediaType != "application/json" {
		return errors.New("expected application/json, got another")
	}
	return nil
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	err := headerCheck(r)
	if err != nil {
		http.Error(w, "expected application/json, got another", http.StatusUnsupportedMediaType)
		return
	}
	body, err := jsonHelper(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bodyTask := model.Task{
		Name:   body.Name,
		UserID: *body.UserID,
	}
	task, err := h.service.Create(r.Context(), bodyTask)
	if err != nil {
		if errors.Is(err, model.ErrInvalidTaskName) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("failed to create task: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("failed to encode tasks response: %v", err)
	}
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	if err = h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, model.ErrTaskNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Printf("failed to delete tasks response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	err := headerCheck(r)
	if err != nil {
		http.Error(w, "expected application/json, got another", http.StatusUnsupportedMediaType)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	body, err := jsonHelper(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, model.ErrTaskNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			log.Printf("failed to update task: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("failed to encode task response: %v", err)
	}

}
