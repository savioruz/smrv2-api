package repository

import (
	"context"

	"github.com/savioruz/smrv2-api/internal/dao/entity"
	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserScheduleRepository interface {
	Repository[entity.UserSchedule]
	GetByStudyPlanID(ctx context.Context, studyPlanID string) (*entity.UserSchedule, error)
	GetSchedules(ctx context.Context, userID string, request *model.UserSchedulesRequest) (*[]model.UserSchedulesResponse, error)
	DeleteAllByUserID(db *gorm.DB, userID string) error
}

type UserScheduleRepositoryImpl struct {
	RepositoryImpl[entity.UserSchedule]
	Log *logrus.Logger
}

func NewUserScheduleRepository(db *gorm.DB, log *logrus.Logger) *UserScheduleRepositoryImpl {
	return &UserScheduleRepositoryImpl{
		RepositoryImpl: RepositoryImpl[entity.UserSchedule]{
			DB: db,
		},
		Log: log,
	}
}

func (r *UserScheduleRepositoryImpl) GetByStudyPlanID(ctx context.Context, studyPlanID string) (*entity.UserSchedule, error) {
	var userSchedule entity.UserSchedule
	if err := r.DB.Where("study_plan_id = ?", studyPlanID).First(&userSchedule).Error; err != nil {
		return nil, err
	}
	return &userSchedule, nil
}

func (r *UserScheduleRepositoryImpl) GetSchedules(ctx context.Context, userID string, request *model.UserSchedulesRequest) (*[]model.UserSchedulesResponse, error) {
	var schedules []model.UserSchedulesResponse
	err := r.DB.Raw(`
		SELECT 
			us.course_code,
			sp.course_name,
			us.class_code,
			ss.day_of_week AS day,
			ss.start_time,
			ss.end_time,
			ss.room_number,
			ss.lecturer_name AS lecturer,
			ss.study_program,
			ss.semester,
			ss.credits
		FROM user_schedules us
		JOIN study_plans sp ON us.study_plan_id = sp.id
		JOIN scraped_schedules ss ON us.course_code = ss.course_code 
			AND us.class_code = ss.class_code
		WHERE sp.user_id = ?
			AND us.deleted_at IS NULL`,
		userID,
	).Limit(request.Limit).Offset(request.Page).Order(request.Sort).Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return &schedules, nil
}

func (r *UserScheduleRepositoryImpl) DeleteAllByUserID(db *gorm.DB, userID string) error {
	q := db.Exec(`
		UPDATE user_schedules us
		SET deleted_at = NOW()
		FROM study_plans sp
		WHERE us.study_plan_id = sp.id
			AND sp.user_id = ?
			AND us.deleted_at IS NULL
	`, userID)
	if err := q.Error; err != nil {
		return err
	}
	return nil
}
