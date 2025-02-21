package repository

import (
	"context"

	"github.com/savioruz/smrv2-api/internal/dao/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	Repository[entity.Subscription]
	GetByUserID(ctx context.Context, userID string) (*entity.Subscription, error)
}

type SubscriptionRepositoryImpl struct {
	RepositoryImpl[entity.Subscription]
	Log *logrus.Logger
}

func NewSubscriptionRepository(db *gorm.DB, log *logrus.Logger) *SubscriptionRepositoryImpl {
	return &SubscriptionRepositoryImpl{
		RepositoryImpl: RepositoryImpl[entity.Subscription]{
			DB: db,
		},
		Log: log,
	}
}

func (r *SubscriptionRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*entity.Subscription, error) {
	var subscription entity.Subscription
	if err := r.DB.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}
