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
	platformmw "github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()
	port := config.String("APP_PORT", "8080")
	manager := platformauth.NewManager(config.String("JWT_ACCESS_SECRET", "dev-access-secret-change-me-123456789"), config.String("JWT_REFRESH_SECRET", "dev-refresh-secret-change-me-123456789"), config.String("JWT_ISSUER", "go-platform"), config.Duration("ACCESS_TOKEN_TTL", 15*time.Minute), config.Duration("REFRESH_TOKEN_TTL", 7*24*time.Hour))
	r := chi.NewRouter()
	r.Use(platformmw.RequestID, platformmw.Recovery(log), platformmw.Logging(log))
	proxy := NewProxy(map[string]string{"auth": config.String("AUTH_SERVICE_URL", "http://localhost:8081"), "catalog": config.String("CATALOG_SERVICE_URL", "http://localhost:8082"), "orders": config.String("ORDER_SERVICE_URL", "http://localhost:8083")})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Mount("/api/v1/auth", proxy.Protected("auth", "/api/v1/auth"))
	r.Group(func(pr chi.Router) {
		pr.Use(platformmw.Authenticate(manager))
		pr.Mount("/api/v1/products", proxy.Protected("catalog", "/api/v1/products"))
		pr.Mount("/api/v1/orders", proxy.Protected("orders", "/api/v1/orders"))
	})
	server := &http.Server{Addr: ":" + port, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		log.Info("gateway started", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error", zap.Error(err))
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
