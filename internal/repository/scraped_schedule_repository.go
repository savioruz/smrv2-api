package repository

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/savioruz/smrv2-api/internal/dao/entity"
	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ScrapedScheduleRepository interface {
	Repository[entity.ScrapedSchedule]
	Exists(ctx context.Context) (bool, error)
	GetByCode(ctx context.Context, courseCode, classCode string) ([]*entity.ScrapedSchedule, error)
	GetStudyPrograms(ctx context.Context) ([]*model.StudyProgram, error)
	GetSchedules(ctx context.Context, request *model.ScrapedScheduleRequest) ([]*entity.ScrapedSchedule, int64, error)
}

type ScrapedScheduleRepositoryImpl struct {
	RepositoryImpl[entity.ScrapedSchedule]
	Log *logrus.Logger
}

func NewScrapedScheduleRepository(db *gorm.DB, log *logrus.Logger) *ScrapedScheduleRepositoryImpl {
	return &ScrapedScheduleRepositoryImpl{
		RepositoryImpl: RepositoryImpl[entity.ScrapedSchedule]{
			DB: db,
		},
		Log: log,
	}
}

func (r *ScrapedScheduleRepositoryImpl) GetStudyPrograms(ctx context.Context) ([]*model.StudyProgram, error) {
	var studyPrograms []*model.StudyProgram
	if err := r.DB.WithContext(ctx).
		Raw("SELECT DISTINCT study_program as name FROM scraped_schedules ORDER BY study_program").
		Scan(&studyPrograms).Error; err != nil {
		return nil, err
	}
	return studyPrograms, nil
}

func (r *ScrapedScheduleRepositoryImpl) Exists(ctx context.Context) (bool, error) {
	var exists bool
	err := r.DB.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM scraped_schedules)").
		Scan(&exists).Error

	if err != nil {
		r.Log.Errorf("Failed to check if schedules exist: %v", err)
		return false, err
	}

	return exists, nil
}

func (r *ScrapedScheduleRepositoryImpl) GetByCode(ctx context.Context, courseCode, classCode string) ([]*entity.ScrapedSchedule, error) {
	var schedules []*entity.ScrapedSchedule
	if err := r.DB.Where("course_code = ? AND class_code = ?", courseCode, classCode).Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *ScrapedScheduleRepositoryImpl) GetSchedules(ctx context.Context, request *model.ScrapedScheduleRequest) ([]*entity.ScrapedSchedule, int64, error) {
	var schedules []*entity.ScrapedSchedule
	var totalCount int64

	query := r.DB.WithContext(ctx).Model(&entity.ScrapedSchedule{})

	if request.StudyProgram != "" {
		query = query.Where("study_program ILIKE ?", "%"+request.StudyProgram+"%")
	}
	if request.CourseCode != "" {
		query = query.Where("course_code ILIKE ?", "%"+request.CourseCode+"%")
	}
	if request.ClassCode != "" {
		query = query.Where("class_code ILIKE ?", "%"+request.ClassCode+"%")
	}
	if request.CourseName != "" {
		query = query.Where("course_name ILIKE ?", "%"+request.CourseName+"%")
	}
	if request.DayOfWeek != "" {
		query = query.Where("day_of_week = ?", request.DayOfWeek)
	}
	if request.StartTime != "" {
		query = query.Where("start_time = ?", request.StartTime)
	}
	if request.EndTime != "" {
		query = query.Where("end_time = ?", request.EndTime)
	}
	if request.RoomNumber != "" {
		query = query.Where("room_number ILIKE ?", "%"+request.RoomNumber+"%")
	}
	if request.Semester != "" {
		query = query.Where("semester = ?", request.Semester)
	}
	if request.LecturerName != "" {
		query = query.Where("lecturer_name ILIKE ?", "%"+request.LecturerName+"%")
	}

	countQuery := query
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(request.Limit)))
	if request.Page > totalPages && totalPages > 0 {
		return nil, totalCount, gorm.ErrRecordNotFound
	}

	// Apply sorting
	if request.Sort != "" {
		direction := "ASC"
		if strings.ToLower(request.Order) == "desc" {
			direction = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", request.Sort, direction))
	}

	// Apply pagination
	offset := (request.Page - 1) * request.Limit
	if err := query.Offset(offset).Limit(request.Limit).Find(&schedules).Error; err != nil {
		return nil, 0, err
	}

	return schedules, totalCount, nil
}
