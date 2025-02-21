package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"text/template"
	"time"

	"github.com/TrinityKnights/Backend/pkg/cache"
	"github.com/go-playground/validator/v10"
	"github.com/savioruz/smrv2-api/internal/dao/entity"
	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/savioruz/smrv2-api/internal/repository"
	"github.com/savioruz/smrv2-api/pkg/helper"
	"github.com/savioruz/smrv2-api/pkg/mail"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ScrapedScheduleService interface {
	SaveScrapedSchedule(ctx context.Context, schedule *entity.ScrapedSchedule) error
	GetSchedules(ctx context.Context, request *model.ScrapedScheduleRequest) (*model.Response[[]model.UserSchedulesResponse], error)
	SyncSchedules(ctx context.Context, request *model.UserSchedulesSyncRequest) (*model.Response[string], error)
	DeleteAllSchedules(ctx context.Context) error
	SetOrchestrator(orchestrator ScheduleOrchestratorService)
}

type ScrapedScheduleServiceImpl struct {
	DB                        *gorm.DB
	Log                       *logrus.Logger
	Validator                 *validator.Validate
	Cache                     *cache.ImplCache
	ScrapedScheduleRepository *repository.ScrapedScheduleRepositoryImpl
	UserRepository            *repository.UserRepositoryImpl
	ScheduleOrchestrator      ScheduleOrchestratorService
	Mail                      *mail.ImplGomail
}

func NewScrapedScheduleService(
	db *gorm.DB,
	log *logrus.Logger,
	validator *validator.Validate,
	cache *cache.ImplCache,
	scrapedScheduleRepository *repository.ScrapedScheduleRepositoryImpl,
	userRepository *repository.UserRepositoryImpl,
	mail *mail.ImplGomail,
) ScrapedScheduleService {
	return &ScrapedScheduleServiceImpl{
		DB:                        db,
		Log:                       log,
		Validator:                 validator,
		Cache:                     cache,
		ScrapedScheduleRepository: scrapedScheduleRepository,
		UserRepository:            userRepository,
		Mail:                      mail,
	}
}

func (s *ScrapedScheduleServiceImpl) SetOrchestrator(orchestrator ScheduleOrchestratorService) {
	s.ScheduleOrchestrator = orchestrator
}

func (s *ScrapedScheduleServiceImpl) SaveScrapedSchedule(ctx context.Context, schedule *entity.ScrapedSchedule) error {
	if !s.isValidSchedule(schedule) {
		s.Log.Infof("Skipping invalid schedule: %s-%s", schedule.CourseCode, schedule.ClassCode)
		return nil
	}

	tx := s.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Try to create the schedule
	err := s.ScrapedScheduleRepository.Create(tx, schedule)
	if err != nil {
		// Check for unique constraint violation
		if tx.Error != nil && tx.Error.Error() == "ERROR: duplicate key value violates unique constraint \"unique_scraped_schedule\"" {
			s.Log.Infof("Schedule already exists: %s-%s on %s at %s",
				schedule.CourseCode,
				schedule.ClassCode,
				schedule.DayOfWeek,
				schedule.StartTime,
			)
			tx.Rollback()
			return nil
		}

		tx.Rollback()
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ScrapedScheduleServiceImpl) GetSchedules(ctx context.Context, request *model.ScrapedScheduleRequest) (*model.Response[[]model.UserSchedulesResponse], error) {
	if err := s.Validator.Struct(request); err != nil {
		return nil, helper.ValidationError(err)
	}

	if request.Page <= 0 {
		request.Page = 1
	}
	if request.Limit <= 0 || request.Limit > 100 {
		request.Limit = 10
	}

	cacheKey := fmt.Sprintf("schedules:program=%s:course=%s:class=%s:name=%s:day=%s:start=%s:end=%s:room=%s:semester=%s:lecturer=%s:page=%d:limit=%d:sort=%s:order=%s",
		request.StudyProgram,
		request.CourseCode,
		request.ClassCode,
		request.CourseName,
		request.DayOfWeek,
		request.StartTime,
		request.EndTime,
		request.RoomNumber,
		request.Semester,
		request.LecturerName,
		request.Page,
		request.Limit,
		request.Sort,
		request.Order,
	)
	var cacheResponse model.Response[[]model.UserSchedulesResponse]
	if err := s.Cache.Get(cacheKey, &cacheResponse); err == nil {
		return &cacheResponse, nil
	}

	schedules, count, err := s.ScrapedScheduleRepository.GetSchedules(ctx, request)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("schedules", "NOT_FOUND")
		}

		return nil, helper.ServerError(s.Log, "failed to get schedules")
	}

	schedulesResponse := make([]model.UserSchedulesResponse, len(schedules))
	for i := range schedules {
		schedulesResponse[i] = model.UserSchedulesResponse{
			CourseCode:   schedules[i].CourseCode,
			CourseName:   schedules[i].CourseName,
			ClassCode:    schedules[i].ClassCode,
			Day:          schedules[i].DayOfWeek,
			StartTime:    schedules[i].StartTime,
			EndTime:      schedules[i].EndTime,
			RoomNumber:   schedules[i].RoomNumber,
			Lecturer:     schedules[i].LecturerName,
			StudyProgram: schedules[i].StudyProgram,
			Semester:     schedules[i].Semester,
		}
	}

	response := model.Response[[]model.UserSchedulesResponse]{
		Data: &schedulesResponse,
		Paging: &model.Paging{
			Page:       request.Page,
			Limit:      request.Limit,
			TotalPage:  int64(math.Ceil(float64(count) / float64(request.Limit))),
			TotalCount: count,
		},
	}

	if err := s.Cache.Set(cacheKey, response, 3*time.Minute); err != nil {
		s.Log.Errorf("failed to set cache: %v", err)
	}

	return &response, nil
}

func (s *ScrapedScheduleServiceImpl) SyncSchedules(ctx context.Context, request *model.UserSchedulesSyncRequest) (*model.Response[string], error) {
	if err := s.Validator.Struct(request); err != nil {
		return nil, helper.ValidationError(err)
	}

	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	id := ctx.Value("userID").(string)
	user, err := s.UserRepository.GetByID(tx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("user", "NOT_FOUND")
		}
		return nil, helper.ServerError(s.Log, "failed to get user")
	}

	if user.Level != "9" {
		return nil, helper.SingleError("user", "UNAUTHORIZED")
	}

	// Get user email from context
	userEmail := ctx.Value("email").(string)
	fromEmail := s.Mail.GetFromEmail()

	var emailBody string
	response := "Schedule sync started. Please wait for the process to complete."
	if request.Message {
		// Email template preparation
		tmpl, err := template.ParseFS(templateFS, "template/sync_schedule.html")
		if err != nil {
			s.Log.Errorf("failed to parse template: %v", err)
			return nil, helper.ServerError(s.Log, "Failed to parse template")
		}
		var body bytes.Buffer
		if err := tmpl.Execute(&body, nil); err != nil {
			s.Log.Errorf("failed to execute template: %v", err)
			return nil, helper.ServerError(s.Log, "Failed to execute template")
		}
		emailBody = body.String()
		response = "Schedule sync started. You will receive an email notification when complete."
	}

	// Run schedule scraping with email notification details
	if err := s.ScheduleOrchestrator.RunScheduleScraping(ctx, &model.ScrapeNotification{
		SendEmail: request.Message,
		UserEmail: userEmail,
		FromEmail: fromEmail,
		EmailBody: emailBody,
	}); err != nil {
		return nil, helper.ServerError(s.Log, "Failed to run schedule scraping")
	}

	return &model.Response[string]{
		Data: &response,
	}, nil
}

func (s *ScrapedScheduleServiceImpl) DeleteAllSchedules(ctx context.Context) error {
	tx := s.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Unscoped().Where("1 = 1").Delete(&entity.ScrapedSchedule{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete schedules: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.Log.Info("Successfully deleted all schedules")
	return nil
}

func (s *ScrapedScheduleServiceImpl) isValidSchedule(schedule *entity.ScrapedSchedule) bool {
	if schedule.CourseCode == "" || schedule.ClassCode == "" ||
		schedule.DayOfWeek == "" || schedule.CourseName == "" {
		return false
	}

	return true
}
