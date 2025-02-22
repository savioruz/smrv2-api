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
	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/savioruz/smrv2-api/internal/gateway/messaging"
	"github.com/savioruz/smrv2-api/internal/repository"
	"github.com/savioruz/smrv2-api/pkg/helper"
	"github.com/savioruz/smrv2-api/pkg/mail"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type UserScheduleService interface {
	GetSchedules(ctx context.Context, request *model.UserSchedulesRequest) (*model.Response[[]model.UserSchedulesResponse], error)
	SyncSchedules(ctx context.Context, request *model.UserSchedulesSyncRequest) (*model.Response[string], error)
}

type UserScheduleServiceImpl struct {
	DB                     *gorm.DB
	Log                    *logrus.Logger
	Validator              *validator.Validate
	Viper                  *viper.Viper
	Cache                  *cache.ImplCache
	UserScheduleRepository *repository.UserScheduleRepositoryImpl
	UserRepository         *repository.UserRepositoryImpl
	Producer               *messaging.UserProducer
	Mail                   *mail.ImplGomail
}

func NewUserScheduleService(
	db *gorm.DB,
	log *logrus.Logger,
	validator *validator.Validate,
	viper *viper.Viper,
	cache *cache.ImplCache,
	userScheduleRepository *repository.UserScheduleRepositoryImpl,
	userRepository *repository.UserRepositoryImpl,
	producer *messaging.UserProducer,
	mail *mail.ImplGomail,
) *UserScheduleServiceImpl {
	return &UserScheduleServiceImpl{
		DB:                     db,
		Log:                    log,
		Validator:              validator,
		Viper:                  viper,
		Cache:                  cache,
		UserScheduleRepository: userScheduleRepository,
		UserRepository:         userRepository,
		Producer:               producer,
		Mail:                   mail,
	}
}

func (s *UserScheduleServiceImpl) GetSchedules(ctx context.Context, request *model.UserSchedulesRequest) (*model.Response[[]model.UserSchedulesResponse], error) {
	if err := s.Validator.Struct(request); err != nil {
		return nil, helper.ValidationError(err)
	}

	if request.Page <= 0 {
		request.Page = 1
	}

	if request.Limit <= 0 || request.Limit > 100 {
		request.Limit = 10
	}

	if request.Sort == "" {
		request.Sort = "day"
	}

	if request.Order == "" {
		request.Order = "asc"
	}

	userID := ctx.Value("userID").(string)
	user, err := s.UserRepository.GetByID(s.DB, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("user", "NOT_FOUND")
		}
		return nil, helper.ServerError(s.Log, "failed to get user")
	}

	if !user.IsPortalVerified {
		return nil, helper.SingleError("portal", "NOT_VERIFIED")
	}

	cacheKey := fmt.Sprintf("schedules:user=%s:page=%d:limit=%d:sort=%s:order=%s", userID, request.Page, request.Limit, request.Sort, request.Order)
	var cacheResponse model.Response[[]model.UserSchedulesResponse]
	if err := s.Cache.Get(cacheKey, &cacheResponse); err == nil {
		return &cacheResponse, nil
	}

	schedules, err := s.UserScheduleRepository.GetSchedules(ctx, userID, request)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("schedule", "NOT_FOUND")
		}
		return nil, helper.ServerError(s.Log, err.Error())
	}

	totalPage := int64(math.Ceil(float64(len(*schedules)) / float64(request.Limit)))
	totalCount := int64(len(*schedules))
	response := model.Response[[]model.UserSchedulesResponse]{
		Data: schedules,
		Paging: &model.Paging{
			Page:       request.Page,
			Limit:      request.Limit,
			TotalPage:  totalPage,
			TotalCount: totalCount,
		},
	}

	if err := s.Cache.Set(cacheKey, response, 3*time.Minute); err != nil {
		s.Log.Errorf("failed to set cache: %v", err)
	}

	return &response, nil
}

func (s *UserScheduleServiceImpl) SyncSchedules(ctx context.Context, request *model.UserSchedulesSyncRequest) (*model.Response[string], error) {
	if err := s.Validator.Struct(request); err != nil {
		return nil, helper.ValidationError(err)
	}

	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Get user
	userID := ctx.Value("userID").(string)
	user, err := s.UserRepository.GetByID(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("user", "NOT_FOUND")
		}
		return nil, helper.ServerError(s.Log, "failed to get user")
	}

	if !user.IsPortalVerified {
		return nil, helper.SingleError("portal", "NOT_VERIFIED")
	}

	// decrypt password
	password, err := helper.DecryptPassword(s.Viper, user.Password)
	if err != nil {
		return nil, helper.ServerError(s.Log, "failed on sync schedules")
	}

	var emailBody string
	response := "Schedule being started. Please wait for the process to complete."
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
		response = "Schedule being started. You will receive an email notification when complete."
	}

	// Publish message to sync schedule with the email content
	if err := s.Producer.PublishStudyDataRequest(ctx, &messaging.StudyDataMessage{
		NIM:       user.Nim,
		Password:  password,
		SendEmail: request.Message,
		UserEmail: user.Email,
		FromEmail: s.Mail.GetFromEmail(),
		EmailBody: emailBody,
	}); err != nil {
		return nil, helper.ServerError(s.Log, "failed to publish study data request")
	}

	return &model.Response[string]{
		Data: &response,
	}, nil
}
