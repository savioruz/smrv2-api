package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/savioruz/smrv2-api/internal/service"
	"github.com/savioruz/smrv2-api/pkg/helper"
	"github.com/sirupsen/logrus"
)

type ScrapedScheduleHandler interface {
	GetSchedules(ctx *fiber.Ctx) error
	SyncSchedules(ctx *fiber.Ctx) error
}

type ScrapedScheduleHandlerImpl struct {
	Log                    *logrus.Logger
	ScrapedScheduleService service.ScrapedScheduleService
}

func NewScrapedScheduleHandler(log *logrus.Logger, scrapedScheduleService service.ScrapedScheduleService) ScrapedScheduleHandler {
	return &ScrapedScheduleHandlerImpl{
		Log:                    log,
		ScrapedScheduleService: scrapedScheduleService,
	}
}

// @Summary Get schedules
// @Description Get schedules
// @Tags Schedule
// @Accept json
// @Produce json
// @Param study_program query string false "Study Program"
// @Param course_code query string false "Course Code"
// @Param class_code query string false "Class Code"
// @Param course_name query string false "Course Name"
// @Param day_of_week query string false "Day of Week"
// @Param room_number query string false "Room Number"
// @Param semester query string false "Semester"
// @Param lecturer_name query string false "Lecturer Name"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param sort_by query string false "Sort by field" enums(course_code, class_code, course_name, credits, day_of_week, room_number, semester, start_time, end_time, lecturer_name)
// @Param sort_order query string false "Sort order" enums(asc, desc)
// @Success 200 {object} model.Response[[]model.UserSchedulesResponse]
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /schedules [get]
func (h *ScrapedScheduleHandlerImpl) GetSchedules(ctx *fiber.Ctx) error {
	request := new(model.ScrapedScheduleRequest)
	if err := ctx.QueryParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helper.SingleError("query", "INVALID_FORMAT"))
	}

	// Map camelCase query params to snake_case struct fields
	if studyProgram := ctx.Query("studyProgram"); studyProgram != "" {
		request.StudyProgram = studyProgram
	}
	if courseCode := ctx.Query("courseCode"); courseCode != "" {
		request.CourseCode = courseCode
	}
	if classCode := ctx.Query("classCode"); classCode != "" {
		request.ClassCode = classCode
	}
	if courseName := ctx.Query("courseName"); courseName != "" {
		request.CourseName = courseName
	}
	if dayOfWeek := ctx.Query("dayOfWeek"); dayOfWeek != "" {
		request.DayOfWeek = dayOfWeek
	}
	if roomNumber := ctx.Query("roomNumber"); roomNumber != "" {
		request.RoomNumber = roomNumber
	}
	if lecturerName := ctx.Query("lecturerName"); lecturerName != "" {
		request.LecturerName = lecturerName
	}

	response, err := h.ScrapedScheduleService.GetSchedules(ctx.Context(), request)
	if err != nil {
		errResp, ok := err.(*helper.ErrorResponse)
		if !ok {
			errResp = helper.ServerError(h.Log, err.Error())
		}

		status := fiber.StatusInternalServerError
		if errResp.RequestID == "" {
			status = fiber.StatusBadRequest
		}
		return ctx.Status(status).JSON(errResp)
	}

	return ctx.JSON(response)
}

// @Summary Sync schedules @admin
// @Description Sync schedules
// @Tags Schedule
// @Accept json
// @Produce json
// @Param body body model.UserSchedulesSyncRequest true "Sync Schedules Request"
// @Success 200 {object} model.Response[string]
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /schedules/sync/all [post]
// @Security BearerAuth
func (h *ScrapedScheduleHandlerImpl) SyncSchedules(ctx *fiber.Ctx) error {
	request := new(model.UserSchedulesSyncRequest)
	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helper.SingleError("body", "INVALID_FORMAT"))
	}

	response, err := h.ScrapedScheduleService.SyncSchedules(ctx.Context(), request)
	if err != nil {
		errResp, ok := err.(*helper.ErrorResponse)
		if !ok {
			errResp = helper.ServerError(h.Log, err.Error())
		}

		status := fiber.StatusInternalServerError
		if errResp.RequestID == "" {
			status = fiber.StatusBadRequest
		}
		return ctx.Status(status).JSON(errResp)
	}

	return ctx.JSON(response)
}
