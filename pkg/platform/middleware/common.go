package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	platformauth "github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/auth"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/httpx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type contextKey string

const userContextKey contextKey = "user"

type UserContext struct {
	UserID string
	Email  string
	Roles  []string
}

func UserFromContext(ctx context.Context) (UserContext, bool) {
	value, ok := ctx.Value(userContextKey).(UserContext)
	return value, ok
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-ID")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Correlation-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey("correlation_id"), id)))
	})
}

func Recovery(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					log.Error("panic recovered", zap.Any("panic", v), zap.ByteString("stack", debug.Stack()))
					httpx.Error(w, 500, "internal_error", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func Logging(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Info("http request", zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.Duration("duration", time.Since(start)))
		})
	}
}

func Authenticate(manager *platformauth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
			if raw == "" {
				httpx.Error(w, 401, "missing_token", "missing access token")
				return
			}
			claims, err := manager.ParseAccess(raw)
			if err != nil {
				httpx.Error(w, 401, "invalid_token", "invalid or expired access token")
				return
			}
			user := UserContext{claims.UserID, claims.Email, claims.Roles}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
		})
	}
}
