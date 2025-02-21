package entity

import (
	"time"

	"gorm.io/gorm"
)

const TableNameUser = "users"

// User mapped from table <users>
type User struct {
	ID                 string         `gorm:"column:id;primaryKey;default:gen_random_uuid()" json:"id"`
	Nim                string         `gorm:"column:nim;not null" json:"nim"`
	Password           string         `gorm:"column:password;not null" json:"password"`
	Name               string         `gorm:"column:name;default:NULL" json:"name"`
	Email              string         `gorm:"column:email;not null" json:"email"`
	Major              string         `gorm:"column:major;default:NULL" json:"major"`
	Level              string         `gorm:"column:level;not null;default:'1'" json:"level"`
	LastLogin          time.Time      `gorm:"column:last_login" json:"last_login"`
	ResetPasswordToken string         `gorm:"column:reset_password_token;default:NULL" json:"reset_password_token"`
	VerificationToken  string         `gorm:"column:verification_token;default:NULL" json:"verification_token"`
	IsVerified         bool           `gorm:"column:is_verified;not null" json:"is_verified"`
	CreatedAt          time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName User's table name
func (*User) TableName() string {
	return TableNameUser
}
