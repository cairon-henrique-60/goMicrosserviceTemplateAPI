package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/apps/catalog-service/internal/outbox"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/apps/catalog-service/internal/product"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/config"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/database"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/messaging"
	platformmw "github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log, _ := zap.NewProduction()
	defer log.Sync()
	db, err := database.Open(ctx, config.String("DATABASE_URL", ""))
	if err != nil {
		log.Fatal("db", zap.Error(err))
	}
	rabbit, err := messaging.Dial(config.String("RABBITMQ_URL", ""))
	if err != nil {
		log.Fatal("rabbit", zap.Error(err))
	}
	defer rabbit.Close()
	_ = rabbit.DeclareTopic("catalog.events")
	go outbox.New(db, rabbit, log).Run(ctx)
	h := product.NewHandler(product.NewService(db))
	r := chi.NewRouter()
	r.Use(platformmw.RequestID, platformmw.Recovery(log), platformmw.Logging(log))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); _, _ = w.Write([]byte("ok")) })
	r.Route("/api/v1/products", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.Get)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
	run(log, r, config.String("APP_PORT", "8082"), cancel)
}
func run(log *zap.Logger, h http.Handler, port string, cancel context.CancelFunc) {
	s := &http.Server{Addr: ":" + port, Handler: h, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server", zap.Error(err))
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	_ = s.Shutdown(ctx)
}
