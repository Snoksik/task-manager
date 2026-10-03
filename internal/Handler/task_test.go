package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Snoksik/task-manager/internal/model"
)

type taskServiceStub struct {
	TaskService
	tasks []model.Task

	getByIDInput  int
	getByIDResult model.Task
	getByIDErr    error
	getByIDCalls  int

	createCalls  int
	createInput  model.Task
	createResult model.Task
	createErr    error

	updateCalls     int
	updateInputID   int
	updateInputTask model.Task
	updateResult    model.Task
	updateErr       error

	Filter       model.TaskFilter
	resultFilter model.TaskFilter
	filterError  error

	getAllCalls int
}

func (t *taskServiceStub) GetByID(id int) (model.Task, error) {
	t.getByIDInput = id
	t.getByIDCalls += 1
	return t.getByIDResult, t.getByIDErr
}
func (s *taskServiceStub) GetAll(filter model.TaskFilter) ([]model.Task, error) {
	s.resultFilter = filter
	s.getAllCalls++
	return s.tasks, s.filterError
}

func (s *taskServiceStub) Create(m model.Task) (model.Task, error) {
	s.createCalls++
	s.createInput = m

	return s.createResult, s.createErr
}

func (s *taskServiceStub) Update(id int, task model.Task) (model.Task, error) {
	s.updateCalls++
	s.updateInputID = id
	s.updateInputTask = task
	return s.updateResult, s.updateErr
}

func TestUpdateSuccess(t *testing.T) {
	svc := &taskServiceStub{
		updateResult: model.Task{
			Name: "Новая задача",
			ID:   7,
		},
	}

	h := NewTaskHandler(svc)

	reader := strings.NewReader(`{"name":"Новая задача"}`)
	req := httptest.NewRequest(http.MethodPut, "/tasks/7", reader)
	req.SetPathValue("id", "7")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if svc.updateCalls != 1 {
		t.Fatalf("expected 1 calls update, got %d", svc.updateCalls)
	}

	if svc.updateInputID != 7 {
		t.Errorf("expected 7 updateInputID , got %d", svc.updateInputID)
	}
	if svc.updateInputTask.ID != 0 {
		t.Errorf("expected no task ID in input, got %d", svc.updateInputTask.ID)
	}
	var body model.Task
	err := json.NewDecoder(rec.Body).Decode(&body)

	if err != nil {
		t.Fatalf("expected zero errors, got %v", err)
	}

	if body.ID != svc.updateResult.ID {
		t.Errorf("expected id = 7, got %d", body.ID)
	}
	if body.Name != svc.updateResult.Name {
		t.Errorf("expected %s, got %s", svc.updateResult.Name, body.Name)
	}

}
func TestUpdateRejectsUnsupportedContentType(t *testing.T) {
	svc := &taskServiceStub{}

	h := NewTaskHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/tasks/7", nil)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.SetPathValue("id", "7")

	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != 415 {
		t.Errorf("expected 415, got %d", rec.Code)
	}
	if svc.updateCalls != 0 {
		t.Errorf("expected zero calls update, got %d", svc.updateCalls)
	}

}

func TestCreateAcceptsJSONContentTypeWithCharset(t *testing.T) {
	svc := &taskServiceStub{}

	h := NewTaskHandler(svc)

	reader := strings.NewReader(`{"name":"Новая задача"}`)

	req := httptest.NewRequest(http.MethodPost, "/tasks", reader)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != 201 {
		t.Errorf("expected status code 201, got %d", rec.Code)
	}
	if svc.createCalls != 1 {
		t.Errorf("expected 1 calls service, got %d", svc.createCalls)
	}

}

func TestCreateRejectsUnsupportedContentType(t *testing.T) {
	svc := &taskServiceStub{}

	h := NewTaskHandler(svc)

	reader := strings.NewReader(`{"name":"Новая задача"}`)

	req := httptest.NewRequest(http.MethodPost, "/tasks", reader)
	req.Header.Set("Content-Type", "text/plain")

	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != 415 {
		t.Errorf("expected status code 415, got %d", rec.Code)
	}
	if svc.createCalls != 0 {
		t.Errorf("expected zero calls service, got %d", svc.createCalls)
	}

}
func TestCreateSuccess(t *testing.T) {
	svc := &taskServiceStub{
		createResult: model.Task{
			ID:   42,
			Name: "Новая задача",
		},
	}

	h := NewTaskHandler(svc)

	reader := strings.NewReader(`{"name":"Новая задача"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", reader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if svc.createCalls != 1 {
		t.Errorf("expected one calls Create, got %d", svc.createCalls)
	}
	if svc.createInput.Name != "Новая задача" {
		t.Errorf("expected 'Новая задача', got %s", svc.createInput.Name)
	}
	if svc.createInput.ID != 0 {
		t.Errorf("expected zero, got %d", svc.createInput.ID)
	}
	if rec.Code != 201 {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}
	var d model.Task

	err := json.NewDecoder(rec.Body).Decode(&d)
	if err != nil {
		t.Fatalf("expected zero errors, got %v", err)
	}
	if d.ID != 42 {
		t.Errorf("expected 42, got %d", d.ID)
	}
	if d.Name != "Новая задача" {
		t.Errorf("expected Новая задача, got %s", d.Name)
	}
}

func TestCreateRejectsMultipleJSONValues(t *testing.T) {
	svc := &taskServiceStub{}

	h := NewTaskHandler(svc)

	reader := strings.NewReader(`{"name":"Первая"} {"name":"Вторая"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", reader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected bad request, got %d", rec.Code)
	}

	if svc.createCalls != 0 {
		t.Errorf("expected 0 cals service, got %d", svc.createCalls)
	}

}

func TestCreateRejectsUnknownField(t *testing.T) {
	svc := &taskServiceStub{}

	h := NewTaskHandler(svc)

	reader := strings.NewReader(`{"name":"Первая", "unexpected":"421"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", reader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected bad request, got %d", rec.Code)
	}

	if svc.createCalls != 0 {
		t.Errorf("expected 0 cals service, got %d", svc.createCalls)
	}

}

func TestCreateRejectsLargeBody(t *testing.T) {
	svc := &taskServiceStub{}

	h := NewTaskHandler(svc)

	reader := strings.NewReader(`{"name":"` + strings.Repeat("a", 200) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", reader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected bad request, got %d", rec.Code)
	}

	if svc.createCalls != 0 {
		t.Errorf("expected 0 cals service, got %d", svc.createCalls)
	}

}
func TestGetByIDSuccess(t *testing.T) {
	svc := &taskServiceStub{
		getByIDResult: model.Task{
			ID:   7,
			Name: "Проверить handler",
		},
	}

	h := NewTaskHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/tasks/7", nil)
	req.SetPathValue("id", "7")

	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if svc.getByIDInput != 7 {
		t.Errorf("expected id = 7, got %d", svc.getByIDInput)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}

	var body model.Task

	err := json.NewDecoder(rec.Body).Decode(&body)
	if err != nil {
		t.Fatalf("expected zero error, got %v", err)
	}

	if body.ID != 7 {
		t.Errorf("expected 7, got %d", body.ID)
	}
	if body.Name != "Проверить handler" {
		t.Errorf("expected 'Проверить handler', got %q", body.Name)
	}

}

func TestGetByIDNotFound(t *testing.T) {
	svc := &taskServiceStub{
		getByIDErr: model.ErrTaskNotFound,
	}

	h := NewTaskHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/tasks/7", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if svc.getByIDInput != 7 {
		t.Errorf("expected id = 7, got %d", svc.getByIDInput)
	}
	if rec.Code != 404 {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

}

func TestGetAll(t *testing.T) {
	svc := &taskServiceStub{
		tasks: []model.Task{
			{ID: 7, Name: "Проверить handler"},
		},
	}
	h := NewTaskHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()

	h.GetAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")

	if contentType != "application/json" {
		t.Errorf("expected header application/json, got it %s", contentType)
	}

	var body []model.Task
	err := json.NewDecoder(rec.Body).Decode(&body)
	if err != nil {
		t.Fatalf("expected zero errors, got %v", err)
	}

	if len(body) != 1 {
		t.Fatalf("expected len 1, got %d", len(body))
	}
	if body[0].ID != 7 {
		t.Errorf("expected id = 7, got %d", body[0].ID)
	}

	if body[0].Name != "Проверить handler" {
		t.Errorf("expevted 'Проверить handler', got %s", body[0].Name)
	}

}

func TestGetByIDInvalidID(t *testing.T) {
	svc := &taskServiceStub{
		tasks: []model.Task{
			{ID: 7, Name: "Проверить handler"},
		},
	}
	h := NewTaskHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)

	rec := httptest.NewRecorder()

	req.SetPathValue("id", "abc")

	h.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected code 400, got %d", rec.Code)
	}
	if svc.getByIDCalls != 0 {
		t.Errorf("expected 0 calls, got %d", svc.getByIDCalls)
	}

}

//POST /tasks с пустым name
//- ожидаемый результат: 400 Bad Request

func TestCreateEmptyName(t *testing.T) {
	svc := &taskServiceStub{
		createErr: model.ErrInvalidTaskName,
	}
	h := NewTaskHandler(svc)
	reader := strings.NewReader(`{"name":""}`)

	req := httptest.NewRequest("POST", "/tasks/1", reader)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != 400 {
		t.Fatalf("expected status 400, got %d ", rec.Code)
	}

}

// PUT /tasks/{id} для несуществующей задачи
// - ожидаемый результат: 404 Not Found

func TestUpdateTaskNotFound(t *testing.T) {
	svc := &taskServiceStub{
		updateErr: model.ErrTaskNotFound,
	}

	h := NewTaskHandler(svc)

	reader := strings.NewReader(`{"name":"ууу"}`)

	req := httptest.NewRequest("PUT", "/tasks/27", reader)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetPathValue("id", "27")

	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != 404 {
		t.Fatalf("Expected status 404, got %d", rec.Code)
	}

}

func TestGetAllUsesDefaultPagination(t *testing.T) {
	svc := &taskServiceStub{
		Filter: model.TaskFilter{
			Offset: 0,
			Limit:  10,
		},
	}

	h := NewTaskHandler(svc)

	req := httptest.NewRequest("GET", "/tasks", nil)

	rec := httptest.NewRecorder()

	h.GetAll(rec, req)
	if svc.resultFilter != svc.Filter {
		t.Fatalf("expected %+v filter, got %+v", svc.Filter, svc.resultFilter)
	}

}
func TestGetAllReturnsBadRequestForInvalidOffset(t *testing.T) {
	svc := &taskServiceStub{}

	h := NewTaskHandler(svc)
	req := httptest.NewRequest("GET", "/tasks?offset=abc", nil)
	rec := httptest.NewRecorder()

	h.GetAll(rec, req)

	if svc.getAllCalls != 0 {
		t.Fatalf("expected zero calls service, got %d", svc.getAllCalls)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}
}

func TestGetAllReturnsBadRequestForInvalidLimit(t *testing.T) {
	svc := &taskServiceStub{}

	h := NewTaskHandler(svc)
	req := httptest.NewRequest("GET", "/tasks?limit=abc", nil)
	rec := httptest.NewRecorder()

	h.GetAll(rec, req)

	if svc.getAllCalls != 0 {
		t.Fatalf("expected zero calls service, got %d", svc.getAllCalls)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}
}

func TestGetAllReturnsBadRequestForOutOfRangeLimit(t *testing.T) {
	svc := &taskServiceStub{
		filterError: model.ErrInvalidTaskFilter,
	}

	h := NewTaskHandler(svc)
	req := httptest.NewRequest("GET", "/tasks?limit=101", nil)
	rec := httptest.NewRecorder()

	h.GetAll(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}
}
