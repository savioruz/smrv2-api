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
			if err := handler(msg.Body); err != nil {
				c.log.Errorf("Failed to process message: %v", err)
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
		}
	}()

	<-ctx.Done()
	return nil
}
