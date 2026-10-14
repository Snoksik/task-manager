package handler

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"time"
)

type requestIDKey struct{}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := rand.Text()
		w.Header().Set("X-Request-ID", id)

		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		newRequest := r.WithContext(ctx)

		start := time.Now()
		next.ServeHTTP(w, newRequest)
		elapsed := time.Since(start)
		slog.Info(
			"request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", id,
			"duration_ms", elapsed.Milliseconds(),
		)
	})
}

func requestIDFromContext(ctx context.Context) string {
	id := ctx.Value(requestIDKey{})
	res, ok := id.(string)
	if !ok {
		return ""
	}
	return res

}
