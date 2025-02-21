package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/savioruz/smrv2-api/pkg/scrape"
	"github.com/sirupsen/logrus"
)

type ScrapedScheduleProducer struct {
	producer *Producer
	log      *logrus.Logger
}

type ScrapeMetadata struct {
	TotalPrograms int
	Notification  *model.ScrapeNotification
}

func NewScrapedScheduleProducer(producer *Producer, log *logrus.Logger) *ScrapedScheduleProducer {
	return &ScrapedScheduleProducer{
		producer: producer,
		log:      log,
	}
}

// Step 1: Publish study programs to be scraped
func (p *ScrapedScheduleProducer) PublishStudyPrograms(ctx context.Context, studyPrograms map[string][]scrape.StudyProgram, metadata *ScrapeMetadata) error {
	// Publish metadata first
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := p.producer.PublishMessage(ctx, model.QueueScrapeMetadata, metadataBytes); err != nil {
		return fmt.Errorf("failed to publish metadata: %w", err)
	}

	// Then publish study programs
	for facultyID, programs := range studyPrograms {
		for _, program := range programs {
			message := model.StudyProgramMessage{
				FacultyID: facultyID,
				ProgramID: program.Value,
				Faculty:   program.Faculty,
				Program:   program.Name,
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				p.log.Errorf("Failed to marshal study program message: %v", err)
				continue
			}

			if err := p.producer.PublishMessage(ctx, model.QueueScrapeSchedule, messageBytes); err != nil {
				p.log.Errorf("Failed to publish message for program %s: %v", program.Name, err)
				continue
			}

			p.log.Infof("Published scrape request for program: %s", program.Name)
		}
	}
	return nil
}

// Step 2: Publish scraped schedules to be saved
func (p *ScrapedScheduleProducer) PublishScheduleData(ctx context.Context, scheduleMsg model.ScheduleMessage) error {
	messageBytes, err := json.Marshal(scheduleMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal schedule message: %w", err)
	}

	err = p.producer.PublishMessage(ctx, model.QueueSaveSchedule, messageBytes)
	if err != nil {
		return fmt.Errorf("failed to publish schedule data: %w", err)
	}

	p.log.Infof("Published save request for program: %s", scheduleMsg.Program)
	return nil
}
