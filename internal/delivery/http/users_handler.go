package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/savioruz/smrv2-api/internal/service"
	"github.com/savioruz/smrv2-api/pkg/helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type UserHandler interface {
	Register(ctx *fiber.Ctx) error
	Login(ctx *fiber.Ctx) error
	VerifyEmail(ctx *fiber.Ctx) error
	RefreshToken(ctx *fiber.Ctx) error
	ResetPassword(ctx *fiber.Ctx) error
	ResetPasswordRequest(ctx *fiber.Ctx) error
}

type UserHandlerImpl struct {
	Log         *logrus.Logger
	Viper       *viper.Viper
	UserService service.UserService
}

func NewUserHandler(log *logrus.Logger, viper *viper.Viper, userService service.UserService) *UserHandlerImpl {
	return &UserHandlerImpl{
		Log:         log,
		Viper:       viper,
		UserService: userService,
	}
}

// @Summary Register
// @Description Register a new user but sync data from portal
// @Tags Auth
// @Accept json
// @Produce json
// @Param register body model.UsersRegisterRequest true "Register Request"
// @Success 200 {object} model.UsersRegisterResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /auth/register [post]
func (h *UserHandlerImpl) Register(ctx *fiber.Ctx) error {
	request := new(model.UsersRegisterRequest)
	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helper.SingleError("body", "INVALID_FORMAT"))
	}

	response, err := h.UserService.Register(ctx.Context(), request)
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

	return ctx.Status(fiber.StatusOK).JSON(model.NewSuccessResponse(response, nil))
}

// @Summary Login
// @Description Login to the system
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body model.UsersLoginRequest true "Login Request"
// @Success 200 {object} model.UsersResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /auth [post]
func (h *UserHandlerImpl) Login(ctx *fiber.Ctx) error {
	request := new(model.UsersLoginRequest)
	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helper.SingleError("body", "INVALID_FORMAT"))
	}

	response, err := h.UserService.Login(ctx.Context(), request)
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

	return ctx.Status(fiber.StatusOK).JSON(model.NewSuccessResponse(response, nil))
}

// @Summary Refresh Token
// @Description Refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh_token body model.UsersRefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} model.UserRefreshTokenResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /auth/refresh [post]
func (h *UserHandlerImpl) RefreshToken(ctx *fiber.Ctx) error {
	request := new(model.UsersRefreshTokenRequest)
	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helper.SingleError("body", "INVALID_FORMAT"))
	}

	response, err := h.UserService.RefreshToken(ctx.Context(), request)
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

	return ctx.Status(fiber.StatusOK).JSON(model.NewSuccessResponse(response, nil))
}

// @Summary Verify Email
// @Description Verify email
// @Tags Auth
// @Accept json
// @Produce json
// @Param token path string true "Token"
// @Success 200 {object} model.Response[string]
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /auth/verify/{token} [get]
func (h *UserHandlerImpl) VerifyEmail(ctx *fiber.Ctx) error {
	request := new(model.UsersVerifyEmailRequest)
	if err := ctx.ParamsParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helper.SingleError("params", "INVALID_FORMAT"))
	}

	_, err := h.UserService.VerifyEmail(ctx.Context(), request)
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

	link := fmt.Sprintf("https://simeru.vercel.app/auth/login?ref=https%%3A%%2F%%2Fsimeru-scraper.koyeb.app&id=%s", request.Token)
	return ctx.Redirect(link, fiber.StatusTemporaryRedirect)
}

// @Summary Reset Password
// @Description Reset password
// @Tags Auth
// @Accept json
// @Produce json
// @Param reset_password body model.UserResetPassword true "Reset Password Request"
// @Success 200 {object} model.Response[string]
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /auth/reset [post]
func (h *UserHandlerImpl) ResetPassword(ctx *fiber.Ctx) error {
	request := new(model.UserResetPassword)
	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helper.SingleError("body", "INVALID_FORMAT"))
	}

	response, err := h.UserService.ResetPassword(ctx.Context(), request)
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

	return ctx.Status(fiber.StatusOK).JSON(response)
}

// @Summary Reset Password Request
// @Description Reset password request
// @Tags Auth
// @Accept json
// @Produce json
// @Param reset_password_request body model.UserResetPasswordRequest true "Reset Password Request"
// @Success 200 {object} model.Response[string]
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /auth/reset/request [post]
func (h *UserHandlerImpl) ResetPasswordRequest(ctx *fiber.Ctx) error {
	request := new(model.UserResetPasswordRequest)
	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helper.SingleError("body", "INVALID_FORMAT"))
	}

	response, err := h.UserService.RequestResetPassword(ctx.Context(), request)
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

	return ctx.Status(fiber.StatusOK).JSON(response)
}
