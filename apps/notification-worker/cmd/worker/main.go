package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/config"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/messaging"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	url := config.String("RABBITMQ_URL", "")
	if err := messaging.DeclareRetryTopology(url, "order.events", "notification.order-created", "order.created"); err != nil {
		log.Fatal("topology", zap.Error(err))
	}
	client, err := messaging.Dial(url)
	if err != nil {
		log.Fatal("rabbit", zap.Error(err))
	}
	defer client.Close()
	log.Info("notification worker started")
	err = client.Consume(ctx, "order.events", "notification.order-created", "order.created", func(ctx context.Context, body []byte) error {
		var event messaging.Event
		if err := json.Unmarshal(body, &event); err != nil {
			return err
		}
		log.Info("notification simulated", zap.String("event", event.Type), zap.String("event_id", event.ID))
		return nil
	})
	if err != nil && ctx.Err() == nil {
		log.Fatal("consumer", zap.Error(err))
	}
}
