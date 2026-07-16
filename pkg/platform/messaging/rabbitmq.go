package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Event struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Version       int             `json:"version"`
	Source        string          `json:"source"`
	CorrelationID string          `json:"correlationId"`
	OccurredAt    time.Time       `json:"occurredAt"`
	Data          json.RawMessage `json:"data"`
}

type Client struct{ conn *amqp.Connection }

func Dial(url string) (*Client, error) {
	conn, err := amqp.DialConfig(url, amqp.Config{Heartbeat: 10 * time.Second})
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}
func (c *Client) Close() error { return c.conn.Close() }
func (c *Client) DeclareTopic(exchange string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	return ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
}
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.Confirm(false); err != nil {
		return err
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	err = ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{ContentType: "application/json", DeliveryMode: amqp.Persistent, MessageId: routingKey, Timestamp: time.Now().UTC(), Body: body})
	if err != nil {
		return err
	}
	select {
	case confirmation := <-confirms:
		if !confirmation.Ack {
			return fmt.Errorf("message not confirmed")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type Handler func(context.Context, []byte) error

func (c *Client) Consume(ctx context.Context, exchange, queue, bindingKey string, handler Handler) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	args := amqp.Table{"x-dead-letter-exchange": exchange + ".dlx"}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, args); err != nil {
		return err
	}
	if err := ch.QueueBind(queue, bindingKey, exchange, false, nil); err != nil {
		return err
	}
	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("delivery channel closed")
			}
			if err := handler(ctx, msg.Body); err != nil {
				_ = msg.Nack(false, false)
			} else {
				_ = msg.Ack(false)
			}
		}
	}
}
