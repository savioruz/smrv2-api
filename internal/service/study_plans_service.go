package service

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/savioruz/smrv2-api/internal/dao/entity"
	"github.com/savioruz/smrv2-api/internal/gateway/messaging"
	"github.com/savioruz/smrv2-api/internal/repository"
	"github.com/savioruz/smrv2-api/pkg/helper"
	"github.com/savioruz/smrv2-api/pkg/scrape"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StudyPlansService interface {
	ProcessStudyData(ctx context.Context, student *scrape.Student, studyPlans []scrape.StudyPlan) error
}

type StudyPlansServiceImpl struct {
	DB                     *gorm.DB
	Log                    *logrus.Logger
	Validator              *validator.Validate
	UserRepository         *repository.UserRepositoryImpl
	StudyRepository        *repository.StudyPlansRepositoryImpl
	UserScheduleRepository *repository.UserScheduleRepositoryImpl
	Producer               *messaging.UserProducer
}

func NewStudyService(
	db *gorm.DB,
	log *logrus.Logger,
	validator *validator.Validate,
	userRepository *repository.UserRepositoryImpl,
	studyRepository *repository.StudyPlansRepositoryImpl,
	userScheduleRepository *repository.UserScheduleRepositoryImpl,
	producer *messaging.UserProducer,
) *StudyPlansServiceImpl {
	return &StudyPlansServiceImpl{
		DB:                     db,
		Log:                    log,
		Validator:              validator,
		UserRepository:         userRepository,
		StudyRepository:        studyRepository,
		UserScheduleRepository: userScheduleRepository,
		Producer:               producer,
	}
}

func (s *StudyPlansServiceImpl) ProcessStudyData(ctx context.Context, studentData *scrape.Student, studyPlans []scrape.StudyPlan) error {
	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Find user by NIM
	user, err := s.UserRepository.GetByNIM(tx, studentData.NIM)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helper.SingleError("user", "NOT_FOUND")
		}
		return helper.ServerError(s.Log, "failed to get user by NIM")
	}

	// Update user data
	user.Name = studentData.Name
	user.Major = studentData.Major
	if err := s.UserRepository.Update(tx, user); err != nil {
		return helper.ServerError(s.Log, "failed to update user data")
	}

	// Delete all existing user schedules for this user first
	if err := s.UserScheduleRepository.DeleteAllByUserID(tx, user.ID); err != nil {
		if err != gorm.ErrRecordNotFound {
			return helper.ServerError(s.Log, err.Error())
		}
	}

	for _, plan := range studyPlans {
		// Delete existing study plans
		if err := tx.Where("course_code = ? AND user_id = ?", plan.Code, user.ID).Delete(&entity.StudyPlan{}).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return helper.ServerError(s.Log, "failed to delete existing study plans")
			}
		}

		// Create new study plan
		studyPlan := &entity.StudyPlan{
			UserID:     user.ID,
			CourseCode: plan.Code,
			ClassCode:  plan.Class,
			CourseName: plan.CourseName,
			Credits:    int32(helper.StringToInt(plan.Credits)),
		}
		if err := s.StudyRepository.Create(tx, studyPlan); err != nil {
			return helper.ServerError(s.Log, "failed to create study plan")
		}

		// Create new user schedule
		userSchedule := &entity.UserSchedule{
			StudyPlanID: studyPlan.ID,
			CourseCode:  plan.Code,
			ClassCode:   plan.Class,
		}
		if err := s.UserScheduleRepository.Create(tx, userSchedule); err != nil {
			return helper.ServerError(s.Log, "failed to create user schedule")
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return helper.ServerError(s.Log, err.Error())
	}

	s.Log.Infof("Successfully processed study data for NIM: %s", studentData.NIM)
	return nil
}
