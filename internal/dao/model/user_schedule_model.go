package model

type UserSchedulesRequest struct {
	Page  int    `query:"page" validate:"omitempty"`
	Limit int    `query:"limit" validate:"omitempty,max=100"`
	Sort  string `query:"sort" validate:"omitempty,oneof=course_code course_name class_code day start_time end_time room_number lecturer"`
	Order string `query:"order" validate:"omitempty,oneof=asc desc ASC DESC"`
}

type UserSchedulesResponse struct {
	CourseCode   string `json:"course_code"`
	CourseName   string `json:"course_name"`
	ClassCode    string `json:"class_code"`
	Day          string `json:"day"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	RoomNumber   string `json:"room_number"`
	Lecturer     string `json:"lecturer,omitempty"`
	StudyProgram string `json:"study_program"`
	Semester     string `json:"semester"`
	Credits      int32  `json:"credits"`
}

type UserSchedulesSyncRequest struct {
	Message bool `json:"message" validate:"required,boolean"`
}
