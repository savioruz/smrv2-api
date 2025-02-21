package messaging

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

type UserProducer struct {
	producer *Producer
	log      *logrus.Logger
}

func NewUserProducer(producer *Producer, log *logrus.Logger) *UserProducer {
	return &UserProducer{
		producer: producer,
		log:      log,
	}
}

type StudyDataMessage struct {
	NIM       string `json:"nim"`
	Password  string `json:"password"`
	SendEmail bool   `json:"send_email"`
	UserEmail string `json:"user_email"`
	FromEmail string `json:"from_email"`
	EmailBody string `json:"email_body"`
}

func (p *UserProducer) PublishStudyDataRequest(ctx context.Context, message *StudyDataMessage) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		p.log.Error("Failed to marshal study data message: ", err)
		return err
	}

	return p.producer.PublishMessage(ctx, "study_data", messageBytes)
}
