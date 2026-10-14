package handler

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
func TestWriteAPIError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeAPIError(rec, 400, "INVALID_TASK_ID", "Invalid task id", "req-123")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected bad request, got %d", rec.Code)
	}
	err := headerCheckInReader(rec)
	if err != nil {
		t.Fatalf("exptected : content-type/application/json, got %v", headerCheckInReader(rec))
	}
	var errorResp ErrorResponse
	err = json.NewDecoder(rec.Body).Decode(&errorResp)
	if err != nil {
		t.Fatalf("expected zero errors in decode, got %v", err)
	}
	if errorResp.RequestID != "req-123" {
		t.Fatalf("expected request id: req-123, got %v", errorResp.RequestID)
	}
	if errorResp.Error.Code != "INVALID_TASK_ID" {
		t.Fatalf("expected code: INVALID-TASK-ID, got %v", errorResp.Error.Code)
	}
	if errorResp.Error.Message != "Invalid task id" {
		t.Fatalf("expected message: Invalid task id, got %v", errorResp.Error.Message)
	}
}
