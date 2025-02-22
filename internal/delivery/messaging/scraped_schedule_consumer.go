package messaging

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/savioruz/smrv2-api/internal/dao/entity"
	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/savioruz/smrv2-api/internal/gateway/messaging"
	"github.com/savioruz/smrv2-api/internal/service"
	"github.com/savioruz/smrv2-api/pkg/helper"
	"github.com/savioruz/smrv2-api/pkg/scrape"
	"github.com/sirupsen/logrus"
)

type ScrapedScheduleConsumer struct {
	consumer               *Consumer
	log                    *logrus.Logger
	scrapedScheduleService service.ScrapedScheduleService
	scrapeProducer         *messaging.ScrapedScheduleProducer
	mailProducer           *messaging.MailProducer
	processedPrograms      int
	totalPrograms          int
	notification           *model.ScrapeNotification
	mu                     sync.Mutex
}

func NewScrapedScheduleConsumer(
	consumer *Consumer,
	log *logrus.Logger,
	scrapedScheduleService service.ScrapedScheduleService,
	producer *messaging.ScrapedScheduleProducer,
	mailProducer *messaging.MailProducer,
) *ScrapedScheduleConsumer {
	return &ScrapedScheduleConsumer{
		consumer:               consumer,
		log:                    log,
		scrapedScheduleService: scrapedScheduleService,
		scrapeProducer:         producer,
		mailProducer:           mailProducer,
	}
}

// Step 2: Scrape Worker - Consumes study programs and produces schedule data
func (c *ScrapedScheduleConsumer) ConsumeScrapeRequests(ctx context.Context) error {
	return c.consumer.Consume(ctx, model.QueueScrapeSchedule, func(body []byte) error {
		var message model.StudyProgramMessage
		if err := json.Unmarshal(body, &message); err != nil {
			c.log.Errorf("Failed to unmarshal message: %v", err)
			return err
		}

		scraper := scrape.NewScrape(1000)
		if err := scraper.Initialize(); err != nil {
			c.log.Errorf("Failed to initialize scraper when processing study program %s: %v", message.Program, err)
			return err
		}
		defer scraper.Cleanup()

		schedules, err := scraper.GetSchedule(ctx, message.FacultyID, message.ProgramID)
		if err != nil {
			c.log.Errorf("Failed to get schedules for program %s: %v", message.Program, err)
			return err
		}

		// Convert schedules to ScheduleMessage
		scheduleEntries := make([]model.ScheduleEntry, len(schedules))
		for i, s := range schedules {
			startTime, endTime, err := helper.CalculateTimeRange(s.Jam)
			if err != nil {
				startTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
				endTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
				c.log.Warnf("Error calculating time range for schedule: %v, using default time %s-%s", err, startTime, endTime)
				continue
			}

			scheduleEntries[i] = model.ScheduleEntry{
				CourseCode:   s.Kode,
				ClassCode:    s.Kelas,
				CourseName:   s.Matkul,
				Credits:      int32(helper.StringToInt(s.Sks)),
				DayOfWeek:    s.Hari,
				RoomNumber:   s.Ruang,
				Semester:     s.Semester,
				StartTime:    startTime,
				EndTime:      endTime,
				LecturerName: s.Dosen,
			}
		}

		scheduleMsg := model.ScheduleMessage{
			FacultyID:    message.FacultyID,
			ProgramID:    message.ProgramID,
			Faculty:      message.Faculty,
			Program:      message.Program,
			ScheduleData: scheduleEntries,
		}

		// Publish to save queue
		if err := c.scrapeProducer.PublishScheduleData(ctx, scheduleMsg); err != nil {
			c.log.Errorf("Failed to publish schedule data: %v", err)
			return err
		}

		return nil
	})
}

// Step 3: Save Worker - Consumes schedule data and saves to database
func (c *ScrapedScheduleConsumer) ConsumeSaveRequests(ctx context.Context) error {
	return c.consumer.Consume(ctx, model.QueueSaveSchedule, func(body []byte) error {
		var message model.ScheduleMessage
		if err := json.Unmarshal(body, &message); err != nil {
			c.log.Errorf("Failed to unmarshal schedule message: %v", err)
			return err
		}

		totalSchedules := len(message.ScheduleData)
		successCount := 0
		for _, schedule := range message.ScheduleData {
			scheduleEntity := &entity.ScrapedSchedule{
				CourseCode:   schedule.CourseCode,
				ClassCode:    schedule.ClassCode,
				CourseName:   schedule.CourseName,
				Credits:      schedule.Credits,
				DayOfWeek:    schedule.DayOfWeek,
				RoomNumber:   schedule.RoomNumber,
				Semester:     schedule.Semester,
				StartTime:    schedule.StartTime.Format("15:04"),
				EndTime:      schedule.EndTime.Format("15:04"),
				LecturerName: schedule.LecturerName,
				StudyProgram: message.Program,
			}

			if err := c.scrapedScheduleService.SaveScrapedSchedule(ctx, scheduleEntity); err != nil {
				c.log.Errorf("Failed to save schedule: %v", err)
				continue
			}
			successCount++
		}

		c.log.Infof("Action completed for: %s, success: %d, total: %d", message.Program, successCount, totalSchedules)

		// Increment processed count and check if complete
		c.mu.Lock()
		c.processedPrograms++
		isComplete := c.processedPrograms == c.totalPrograms
		notification := c.notification
		c.mu.Unlock()

		// If all programs are processed and email notification was requested
		if isComplete && notification != nil && notification.SendEmail {
			if err := c.mailProducer.PublishEmailSending(ctx, &messaging.EmailMessage{
				To:      notification.UserEmail,
				From:    notification.FromEmail,
				Subject: "[Smrv2] All Scraped Schedule Sync Completed",
				Body:    notification.EmailBody,
			}); err != nil {
				c.log.Errorf("Failed to send completion email: %v", err)
			}
		}

		return nil
	})
}

func (c *ScrapedScheduleConsumer) ConsumeMetadata(ctx context.Context) error {
	return c.consumer.Consume(ctx, model.QueueScrapeMetadata, func(body []byte) error {
		var metadata messaging.ScrapeMetadata
		if err := json.Unmarshal(body, &metadata); err != nil {
			return err
		}

		c.mu.Lock()
		c.totalPrograms = metadata.TotalPrograms
		c.notification = metadata.Notification
		c.processedPrograms = 0
		c.mu.Unlock()

		return nil
	})
}
