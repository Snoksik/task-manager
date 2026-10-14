package handler

import (
	"crypto/rand"
	"net/http"
)

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := rand.Text()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}
