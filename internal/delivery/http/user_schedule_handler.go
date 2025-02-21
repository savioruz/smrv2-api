package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/savioruz/smrv2-api/internal/service"
	"github.com/savioruz/smrv2-api/pkg/helper"
	"github.com/sirupsen/logrus"
)

type UserScheduleHandler interface {
	GetSchedules(ctx *fiber.Ctx) error
	SyncSchedule(ctx *fiber.Ctx) error
}

type UserScheduleHandlerImpl struct {
	Log                 *logrus.Logger
	UserScheduleService service.UserScheduleService
}

func NewUserScheduleHandler(log *logrus.Logger, userScheduleService service.UserScheduleService) *UserScheduleHandlerImpl {
	return &UserScheduleHandlerImpl{
		Log:                 log,
		UserScheduleService: userScheduleService,
	}
}

// @Summary Get User Schedules
// @Description Get user schedules
// @Tags User
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param sort_by query string false "Sort by field" enums(course_code, class_code, course_name, credits, day_of_week, room_number, semester, start_time, end_time, lecturer_name)
// @Param sort_order query string false "Sort order" enums(asc, desc)
// @Success 200 {object} model.Response[[]model.UserSchedulesResponse]
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /user/schedules [get]
// @Security BearerAuth
func (h *UserScheduleHandlerImpl) GetSchedules(ctx *fiber.Ctx) error {
	request := model.UserSchedulesRequest{}
	if err := ctx.QueryParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helper.SingleError("query", "INVALID_FORMAT"))
	}

	response, err := h.UserScheduleService.GetSchedules(ctx.Context(), &request)
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

// @Summary Sync Schedule
// @Description Sync schedule
// @Tags User
// @Accept json
// @Produce json
// @Param body body model.UserSchedulesSyncRequest true "Sync Schedule Request"
// @Success 200 {object} model.Response[string]
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /user/schedules/sync [post]
// @Security BearerAuth
func (h *UserScheduleHandlerImpl) SyncSchedule(ctx *fiber.Ctx) error {
	request := model.UserSchedulesSyncRequest{}
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helper.SingleError("body", "INVALID_FORMAT"))
	}

	response, err := h.UserScheduleService.SyncSchedules(ctx.Context(), &request)
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
