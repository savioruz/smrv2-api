package repository

import (
	"context"

	"github.com/savioruz/smrv2-api/internal/dao/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StudyPlansRepository interface {
	Repository[entity.StudyPlan]
	GetByID(ctx context.Context, id string) (*entity.StudyPlan, error)
}

type StudyPlansRepositoryImpl struct {
	RepositoryImpl[entity.StudyPlan]
	Log *logrus.Logger
}

func NewStudyRepository(db *gorm.DB, log *logrus.Logger) *StudyPlansRepositoryImpl {
	return &StudyPlansRepositoryImpl{
		RepositoryImpl: RepositoryImpl[entity.StudyPlan]{
			DB: db,
		},
		Log: log,
	}
}

func (r *StudyPlansRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.StudyPlan, error) {
	var studyPlan entity.StudyPlan
	if err := r.DB.Where("id = ?", id).First(&studyPlan).Error; err != nil {
		return nil, err
	}
	return &studyPlan, nil
}
