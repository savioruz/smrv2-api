package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/savioruz/smrv2-api/internal/gateway/messaging"
	"github.com/savioruz/smrv2-api/internal/service"
	"github.com/savioruz/smrv2-api/pkg/mail"
	"github.com/savioruz/smrv2-api/pkg/scrape"
	"github.com/sirupsen/logrus"
)

type UserConsumer struct {
	consumer     *Consumer
	log          *logrus.Logger
	studyService service.StudyPlansService
	mail         *mail.ImplGomail
	producer     *messaging.MailProducer
}

func NewUserConsumer(
	consumer *Consumer,
	log *logrus.Logger,
	studyService service.StudyPlansService,
	mail *mail.ImplGomail,
	producer *messaging.MailProducer,
) *UserConsumer {
	return &UserConsumer{
		consumer:     consumer,
		log:          log,
		studyService: studyService,
		mail:         mail,
		producer:     producer,
	}
}

func (c *UserConsumer) ConsumeStudyData(ctx context.Context) error {
	return c.consumer.Consume(ctx, "study_data", func(body []byte) error {
		var message messaging.StudyDataMessage
		if err := json.Unmarshal(body, &message); err != nil {
			c.log.Errorf("Failed to unmarshal message: %v", err)
			return err
		}

		c.log.Infof("Processing study data for NIM: %s", message.NIM)

		scraper := scrape.NewScrape(60)
		if err := scraper.Initialize(); err != nil {
			c.log.Errorf("Failed to initialize scraper when processing study data for NIM %s: %v", message.NIM, err)
			return err
		}
		defer scraper.Cleanup()

		// Login with the received credentials
		err := scraper.Login(ctx, scrape.Identity{
			NIM:      message.NIM,
			Password: message.Password,
		})
		if err != nil {
			c.log.Errorf("Failed to login for NIM %s: %v", message.NIM, err)
			return err
		}

		// Add small delay after login to ensure session is established
		time.Sleep(2 * time.Second)

		// Get student data
		studentData, err := scraper.GetStudentData(ctx)
		if err != nil {
			if err == context.DeadlineExceeded {
				c.log.Errorf("Timeout while getting student data for NIM %s", message.NIM)
				return fmt.Errorf("timeout while fetching student data")
			}
			c.log.Errorf("Failed to get student data: %v", err)
			return err
		}

		// Get study plans
		studyPlans, err := scraper.GetStudyPlans(ctx)
		if err != nil {
			c.log.Errorf("Failed to get study plans for NIM %s: %v", message.NIM, err)
			return err
		}

		// Process the data
		if err := c.studyService.ProcessStudyData(ctx, studentData, studyPlans); err != nil {
			c.log.Errorf("Failed to process study data: %v", err)
			return err
		}

		if message.SendEmail {
			if err := c.producer.PublishEmailSending(ctx, &messaging.EmailMessage{
				To:      message.UserEmail,
				From:    message.FromEmail,
				Subject: "[Smrv2] Schedule Synced Successfully",
				Body:    message.EmailBody,
			}); err != nil {
				c.log.Errorf("failed to publish email sending: %v", err)
			}
		}

		return nil
	})
}

func (c *UserConsumer) ConsumeEmailSending(ctx context.Context) error {
	return c.consumer.Consume(ctx, "email_sending", func(body []byte) error {
		var message messaging.EmailMessage
		if err := json.Unmarshal(body, &message); err != nil {
			c.log.Errorf("Failed to unmarshal email message: %v", err)
			return err
		}

		mailRequest := mail.SendEmail{
			EmailTo:   message.To,
			EmailFrom: message.From,
			Subject:   message.Subject,
			Body:      *bytes.NewBufferString(message.Body),
		}

		if err := c.mail.SendEmail(&mailRequest); err != nil {
			c.log.Errorf("Failed to send email: %v", err)
			return err
		}

		c.log.Infof("Email sent successfully to: %s", message.To)
		return nil
	})
}
