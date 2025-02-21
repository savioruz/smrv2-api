package entity

import (
	"time"

	"gorm.io/gorm"
)

const TableNameScrapedSchedule = "scraped_schedules"

type ScrapedSchedule struct {
	ID           string         `gorm:"column:id;primaryKey;default:gen_random_uuid()" json:"id"`
	CourseCode   string         `gorm:"column:course_code;not null" json:"course_code"`
	ClassCode    string         `gorm:"column:class_code;not null" json:"class_code"`
	CourseName   string         `gorm:"column:course_name;not null" json:"course_name"`
	Credits      int32          `gorm:"column:credits;not null" json:"credits"`
	DayOfWeek    string         `gorm:"column:day_of_week;not null" json:"day_of_week"`
	RoomNumber   string         `gorm:"column:room_number;not null" json:"room_number"`
	Semester     string         `gorm:"column:semester;not null" json:"semester"`
	StartTime    string         `gorm:"column:start_time;not null" json:"start_time"`
	EndTime      string         `gorm:"column:end_time;not null" json:"end_time"`
	LecturerName string         `gorm:"column:lecturer_name" json:"lecturer_name"`
	StudyProgram string         `gorm:"column:study_program;not null" json:"study_program"`
	CreatedAt    time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

func (*ScrapedSchedule) TableName() string {
	return TableNameScrapedSchedule
}
