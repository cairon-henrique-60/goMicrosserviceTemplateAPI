package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	platformauth "github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/auth"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/config"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/database"
	platformmw "github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/middleware"

	authmodule "github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/apps/auth-service/internal/auth"
	httpapi "github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/apps/auth-service/internal/http"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/apps/auth-service/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	log, _ := zap.NewProduction()
	defer log.Sync()
	db, err := database.Open(ctx, config.String("DATABASE_URL", ""))
	if err != nil {
		log.Fatal("database", zap.Error(err))
	}
	redisClient := redis.NewClient(&redis.Options{Addr: config.String("REDIS_ADDR", "localhost:6379"), Password: config.String("REDIS_PASSWORD", "")})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("redis", zap.Error(err))
	}
	refreshTTL := config.Duration("REFRESH_TOKEN_TTL", 7*24*time.Hour)
	jwt := platformauth.NewManager(config.String("JWT_ACCESS_SECRET", "dev-access-secret-change-me-123456789"), config.String("JWT_REFRESH_SECRET", "dev-refresh-secret-change-me-123456789"), config.String("JWT_ISSUER", "go-platform"), config.Duration("ACCESS_TOKEN_TTL", 15*time.Minute), refreshTTL)
	handler := httpapi.NewHandler(authmodule.NewService(user.NewRepository(db), authmodule.NewSessionStore(redisClient), jwt, refreshTTL))
	r := chi.NewRouter()
	r.Use(platformmw.RequestID, platformmw.Recovery(log), platformmw.Logging(log))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); _, _ = w.Write([]byte("ok")) })
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)
		r.Post("/refresh", handler.Refresh)
		r.Post("/logout", handler.Logout)
	})
	run(log, r, config.String("APP_PORT", "8081"))
}
func run(log *zap.Logger, h http.Handler, port string) {
	s := &http.Server{Addr: ":" + port, Handler: h, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server", zap.Error(err))
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = s.Shutdown(ctx)
}
