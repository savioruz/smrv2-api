package messaging

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

type MailProducer struct {
	producer *Producer
	log      *logrus.Logger
}

type EmailMessage struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func NewMailProducer(producer *Producer, log *logrus.Logger) *MailProducer {
	return &MailProducer{
		producer: producer,
		log:      log,
	}
}

func (p *MailProducer) PublishEmailSending(ctx context.Context, email *EmailMessage) error {
	messageBytes, err := json.Marshal(email)
	if err != nil {
		p.log.Error("Failed to marshal email message: ", err)
		return err
	}

	return p.producer.PublishMessage(ctx, "email_sending", messageBytes)
}
