package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/apps/catalog-service/internal/product"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/messaging"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Worker struct {
	db     *gorm.DB
	rabbit *messaging.Client
	log    *zap.Logger
}

func New(db *gorm.DB, r *messaging.Client, l *zap.Logger) *Worker { return &Worker{db, r, l} }
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.flush(ctx)
		}
	}
}
func (w *Worker) flush(ctx context.Context) {
	var events []product.OutboxEvent
	if err := w.db.WithContext(ctx).Where("published_at IS NULL").Order("created_at").Limit(100).Find(&events).Error; err != nil {
		return
	}
	for _, e := range events {
		envelope := messaging.Event{ID: uuid.NewString(), Type: e.EventType, Version: 1, Source: "catalog-service", OccurredAt: e.CreatedAt, Data: e.Payload}
		body, _ := json.Marshal(envelope)
		pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := w.rabbit.Publish(pubCtx, "catalog.events", e.EventType, body)
		cancel()
		if err != nil {
			w.log.Warn("outbox publish failed", zap.Error(err))
			continue
		}
		now := time.Now().UTC()
		_ = w.db.WithContext(ctx).Model(&product.OutboxEvent{}).Where("id = ?", e.ID).Update("published_at", now).Error
	}
}
