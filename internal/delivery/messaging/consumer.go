package messaging

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	conn *amqp.Connection
	log  *logrus.Logger
}

func NewConsumer(conn *amqp.Connection, log *logrus.Logger) *Consumer {
	return &Consumer{
		conn: conn,
		log:  log,
	}
}

func (c *Consumer) Consume(ctx context.Context, queueName string, handler func([]byte) error) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	go func() {
		for msg := range msgs {
			headers := msg.Headers
			if headers == nil {
				headers = make(amqp.Table)
			}
			retryCount, _ := headers["retry_count"].(int32)

			if err := handler(msg.Body); err != nil {
				retryCount++

				if retryCount >= 5 {
					c.log.Errorf("Message failed after %d attempts, discarding: %v", retryCount, err)
					msg.Nack(false, false) // Don't requeue
					continue
				}

				c.log.Warnf("Failed to process message (attempt %d/5): %v", retryCount, err)
				headers["retry_count"] = retryCount

				ch.PublishWithContext(ctx,
					"",             // exchange
					msg.RoutingKey, // routing key
					false,          // mandatory
					false,          // immediate
					amqp.Publishing{
						ContentType: msg.ContentType,
						Body:        msg.Body,
						Headers:     headers,
					},
				)

				msg.Ack(false)
				continue
			}

			msg.Ack(false)
		}
	}()

	<-ctx.Done()
	return nil
}
