package entity

import (
	"time"

	"gorm.io/gorm"
)

const TableNameStudyPlan = "study_plans"

// StudyPlan mapped from table <study_plans>
type StudyPlan struct {
	ID         string         `gorm:"column:id;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string         `gorm:"column:user_id;not null" json:"user_id"`
	CourseCode string         `gorm:"column:course_code;not null" json:"course_code"`
	ClassCode  string         `gorm:"column:class_code;not null" json:"class_code"`
	CourseName string         `gorm:"column:course_name;not null" json:"course_name"`
	Credits    int32          `gorm:"column:credits;not null" json:"credits"`
	CreatedAt  time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName StudyPlan's table name
func (*StudyPlan) TableName() string {
	return TableNameStudyPlan
}
