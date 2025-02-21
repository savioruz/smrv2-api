package repository

import (
	"github.com/savioruz/smrv2-api/internal/dao/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UsersRepository interface {
	Repository[entity.User]
	GetByID(db *gorm.DB, id string) (*entity.User, error)
	GetByNIM(db *gorm.DB, nim string) (*entity.User, error)
	GetByEmail(db *gorm.DB, email string) (*entity.User, error)
	GetByVerificationToken(db *gorm.DB, token string) (*entity.User, error)
}

type UserRepositoryImpl struct {
	RepositoryImpl[entity.User]
	Log *logrus.Logger
}

func NewUsersRepository(db *gorm.DB, log *logrus.Logger) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		RepositoryImpl: RepositoryImpl[entity.User]{
			DB: db,
		},
		Log: log,
	}
}

func (r *UserRepositoryImpl) GetByID(db *gorm.DB, id string) (*entity.User, error) {
	var user entity.User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) GetByNIM(db *gorm.DB, nim string) (*entity.User, error) {
	var user entity.User
	if err := db.Where("nim = ?", nim).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) GetByEmail(db *gorm.DB, email string) (*entity.User, error) {
	var user entity.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) GetByVerificationToken(db *gorm.DB, token string) (*entity.User, error) {
	var user entity.User
	if err := db.Where("verification_token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
