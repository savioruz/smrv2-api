package service

import (
	"context"
	"fmt"

	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/savioruz/smrv2-api/internal/gateway/messaging"
	"github.com/savioruz/smrv2-api/pkg/scrape"
	"github.com/sirupsen/logrus"
)

type ScheduleOrchestratorService interface {
	RunScheduleScraping(ctx context.Context, notification *model.ScrapeNotification) error
}

type ScheduleOrchestratorServiceImpl struct {
	log                    *logrus.Logger
	scrapedScheduleService ScrapedScheduleService
	scheduleProducer       *messaging.ScrapedScheduleProducer
}

func NewScheduleOrchestratorService(
	log *logrus.Logger,
	scrapedScheduleService ScrapedScheduleService,
	scheduleProducer *messaging.ScrapedScheduleProducer,
) ScheduleOrchestratorService {
	return &ScheduleOrchestratorServiceImpl{
		log:                    log,
		scrapedScheduleService: scrapedScheduleService,
		scheduleProducer:       scheduleProducer,
	}
}

func (s *ScheduleOrchestratorServiceImpl) RunScheduleScraping(ctx context.Context, notification *model.ScrapeNotification) error {
	s.log.Info("Starting scheduled scraping job")

	// Delete all existing schedules first
	if err := s.scrapedScheduleService.DeleteAllSchedules(ctx); err != nil {
		return fmt.Errorf("failed to delete existing schedules: %w", err)
	}

	// Get study programs
	scraper := scrape.NewScrape(1000)
	if err := scraper.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize scraper: %w", err)
	}
	defer scraper.Cleanup()

	studyPrograms, err := scraper.GetStudyPrograms(ctx)
	if err != nil {
		return fmt.Errorf("failed to get study programs: %w", err)
	}

	totalPrograms := 0
	for _, programs := range studyPrograms {
		totalPrograms += len(programs)
	}

	// Publish to queue with notification details
	if err := s.scheduleProducer.PublishStudyPrograms(ctx, studyPrograms, &messaging.ScrapeMetadata{
		TotalPrograms: totalPrograms,
		Notification:  notification,
	}); err != nil {
		return fmt.Errorf("failed to publish study programs: %w", err)
	}

	return nil
}
