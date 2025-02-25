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

		scraper := scrape.NewScrape(180) // 3 minutes timeout
		if err := scraper.Initialize(); err != nil {
			c.log.Errorf("Failed to initialize scraper: %v", err)
			return err
		}
		defer scraper.Cleanup()

		loginCtx, loginCancel := context.WithTimeout(ctx, 60*time.Second)
		defer loginCancel()

		var err error
		for attempts := 1; attempts <= 3; attempts++ {
			err = scraper.Login(loginCtx, scrape.Identity{
				NIM:      message.NIM,
				Password: message.Password,
			})
			if err == nil {
				break
			}
			if attempts < 3 {
				time.Sleep(time.Duration(attempts) * 5 * time.Second)
				c.log.Warnf("Retry %d: Login failed for NIM %s: %v", attempts, message.NIM, err)
			}
		}
		if err != nil {
			return fmt.Errorf("all login attempts failed: %v", err)
		}

		time.Sleep(5 * time.Second)

		dataCtx, dataCancel := context.WithTimeout(ctx, 90*time.Second)
		defer dataCancel()

		var studentData *scrape.Student
		for attempts := 1; attempts <= 3; attempts++ {
			studentData, err = scraper.GetStudentData(dataCtx)
			if err == nil {
				break
			}
			if attempts < 3 {
				time.Sleep(time.Duration(attempts) * 5 * time.Second)
				c.log.Warnf("Retry %d: Failed to get student data: %v", attempts, err)
			}
		}
		if err != nil {
			return fmt.Errorf("failed to get student data after retries: %v", err)
		}

		var studyPlans []scrape.StudyPlan
		for attempts := 1; attempts <= 3; attempts++ {
			studyPlans, err = scraper.GetStudyPlans(dataCtx)
			if err == nil {
				break
			}
			if attempts < 3 {
				time.Sleep(time.Duration(attempts) * 5 * time.Second)
				c.log.Warnf("Retry %d: Failed to get study plans: %v", attempts, err)
			}
		}
		if err != nil {
			return fmt.Errorf("failed to get study plans after retries: %v", err)
		}

		// Process the data
		if err := c.studyService.ProcessStudyData(ctx, studentData, studyPlans); err != nil {
			return fmt.Errorf("failed to process study data: %v", err)
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
