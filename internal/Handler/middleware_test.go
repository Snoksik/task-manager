package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	var receivedID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(requestIDKey{})
		id, ok := value.(string)
		if ok != true {
			t.Fatal("expected request ID to be a string")
		}
		receivedID = id
		w.WriteHeader(http.StatusCreated)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)

	res := RequestIDMiddleware(next)
	res.ServeHTTP(rec, req)

	id := rec.Result().Header.Get("X-Request-ID")
	if id != receivedID {
		t.Fatalf("expected receivedID = id, got not")
	}
	if id == "" {
		t.Fatalf("expected not empty str, got emty")
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/tasks", nil)

	res.ServeHTTP(rec2, req2)
	id2 := rec2.Result().Header.Get("X-Request-ID")
	if id2 == "" {
		t.Fatalf("expected not empty str, got emty")
	}
	if rec2.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec2.Code)

	}
	if id == id2 {
		t.Fatalf("expected different request id, got identical")
	}
}
