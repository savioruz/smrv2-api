package model

import "time"

type StudyProgramMessage struct {
	FacultyID string `json:"faculty_id"`
	ProgramID string `json:"program_id"`
	Faculty   string `json:"faculty"`
	Program   string `json:"program"`
}

type StudyProgram struct {
	ID   uint8  `json:"id"`
	Name string `json:"name"`
}

type ScheduleMessage struct {
	FacultyID    string          `json:"faculty_id"`
	ProgramID    string          `json:"program_id"`
	Faculty      string          `json:"faculty"`
	Program      string          `json:"program"`
	ScheduleData []ScheduleEntry `json:"schedule_data"`
}

type ScheduleEntry struct {
	CourseCode   string    `json:"course_code"`
	ClassCode    string    `json:"class_code"`
	CourseName   string    `json:"course_name"`
	Credits      int32     `json:"credits"`
	DayOfWeek    string    `json:"day_of_week"`
	RoomNumber   string    `json:"room_number"`
	Semester     string    `json:"semester"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	LecturerName string    `json:"lecturer_name"`
}

type ScrapedScheduleRequest struct {
	CourseCode   string `query:"course_code" validate:"omitempty"`
	ClassCode    string `query:"class_code" validate:"omitempty"`
	CourseName   string `query:"course_name" validate:"omitempty"`
	Credits      int    `query:"credits" validate:"omitempty"`
	DayOfWeek    string `query:"day_of_week" validate:"omitempty,oneof=Senin Selasa Rabu Kamis Jumat Sabtu"`
	RoomNumber   string `query:"room_number" validate:"omitempty"`
	StartTime    string `query:"start_time" validate:"omitempty"`
	EndTime      string `query:"end_time" validate:"omitempty"`
	Semester     string `query:"semester" validate:"omitempty"`
	LecturerName string `query:"lecturer_name" validate:"omitempty"`
	StudyProgram string `query:"study_program" validate:"omitempty"`
	Page         int    `query:"page" validate:"omitempty"`
	Limit        int    `query:"limit" validate:"omitempty,max=100"`
	Sort         string `query:"sort" validate:"omitempty,oneof=course_code class_code course_name credits day_of_week room_number semester start_time end_time lecturer_name"`
	Order        string `query:"order" validate:"omitempty,oneof=asc desc"`
}

type ScrapeNotification struct {
	SendEmail bool   `json:"send_email"`
	UserEmail string `json:"user_email"`
	FromEmail string `json:"from_email"`
	EmailBody string `json:"email_body"`
}

const (
	QueueScrapeSchedule = "scrape_schedule_queue"
	QueueSaveSchedule   = "save_schedule_queue"
	QueueScrapeMetadata = "scrape_metadata_queue"
)
