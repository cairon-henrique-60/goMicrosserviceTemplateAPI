package messaging

import amqp "github.com/rabbitmq/amqp091-go"

func DeclareRetryTopology(url, exchange, queue, bindingKey string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(exchange+".dlx", "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(queue+".dlq", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(queue+".dlq", "#", exchange+".dlx", false, nil); err != nil {
		return err
	}
	args := amqp.Table{"x-dead-letter-exchange": exchange + ".dlx"}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, args); err != nil {
		return err
	}
	return ch.QueueBind(queue, bindingKey, exchange, false, nil)
}
