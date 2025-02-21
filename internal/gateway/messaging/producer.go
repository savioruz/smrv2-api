package messaging

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Producer struct {
	conn *amqp.Connection
	log  *logrus.Logger
}

func NewProducer(conn *amqp.Connection, log *logrus.Logger) *Producer {
	return &Producer{
		conn: conn,
		log:  log,
	}
}

func (p *Producer) PublishMessage(ctx context.Context, queueName string, body []byte) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
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
		return err
	}

	return ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
}
