package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error     APIError `json:"error"`
	RequestID string   `json:"request_id"`
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string, requestID string) {
	response := &ErrorResponse{
		RequestID: requestID,
		Error: APIError{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("failed to encode API error response: %v", err)
	}
}
