package entity

import (
	"time"

	"gorm.io/gorm"
)

const TableNameSubscription = "subscriptions"

// Subscription mapped from table <subscriptions>
type Subscription struct {
	ID        string         `gorm:"column:id;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string         `gorm:"column:user_id;not null" json:"user_id"`
	PlanType  string         `gorm:"column:plan_type;not null" json:"plan_type"`
	Status    string         `gorm:"column:status;not null" json:"status"`
	StartDate time.Time      `gorm:"column:start_date;not null" json:"start_date"`
	EndDate   time.Time      `gorm:"column:end_date;not null" json:"end_date"`
	CreatedAt time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName Subscription's table name
func (*Subscription) TableName() string {
	return TableNameSubscription
}
