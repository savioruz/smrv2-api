package entity

import (
	"time"

	"gorm.io/gorm"
)

const TableNameUserSchedule = "user_schedules"

// Schedule mapped from table <schedules>
type UserSchedule struct {
	ID          string         `gorm:"column:id;primaryKey;default:gen_random_uuid()" json:"id"`
	StudyPlanID string         `gorm:"column:study_plan_id;not null" json:"study_plan_id"`
	CourseCode  string         `gorm:"column:course_code;not null" json:"course_code"`
	ClassCode   string         `gorm:"column:class_code;not null" json:"class_code"`
	CreatedAt   time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName Schedule's table name
func (*UserSchedule) TableName() string {
	return TableNameUserSchedule
}
