package builder

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/savioruz/smrv2-api/internal/delivery/http"
	"github.com/savioruz/smrv2-api/internal/delivery/http/middleware"
	"github.com/savioruz/smrv2-api/pkg/helper"
	"github.com/savioruz/smrv2-api/pkg/jwt"
)

type Config struct {
	App                    *fiber.App
	Jwt                    *jwt.JWTServiceImpl
	UserHandler            http.UserHandler
	UserScheduleHandler    http.UserScheduleHandler
	ScrapedScheduleHandler http.ScrapedScheduleHandler
}

type Route struct {
	Method  string
	Path    string
	Handler fiber.Handler
}

func (c *Config) Build() {
	authMiddleware := middleware.NewAuthMiddleware(c.Jwt)
	c.App.Use(recover.New())

	g := c.App.Group("/api/v1")

	for _, route := range c.PublicRoute() {
		g.Add(route.Method, route.Path, route.Handler)
	}

	for _, route := range c.PrivateRoute() {
		g.Add(route.Method, route.Path, authMiddleware.ValidateJWT(), route.Handler)
	}

	c.SwaggerRoute()
	c.NotFoundRoute()
}

func (c *Config) PublicRoute() []Route {
	return []Route{
		// auth
		{Method: "POST", Path: "/auth", Handler: c.UserHandler.Login},
		{Method: "POST", Path: "/auth/register", Handler: c.UserHandler.Register},
		{Method: "GET", Path: "/auth/verify/:token", Handler: c.UserHandler.VerifyEmail},
		{Method: "POST", Path: "/auth/reset", Handler: c.UserHandler.ResetPassword},
		{Method: "POST", Path: "/auth/reset/request", Handler: c.UserHandler.ResetPasswordRequest},
		// schedules
		{Method: "GET", Path: "/schedules", Handler: c.ScrapedScheduleHandler.GetSchedules},
		// study
		{Method: "GET", Path: "/study/programs", Handler: c.ScrapedScheduleHandler.GetStudyPrograms},
	}
}

func (c *Config) PrivateRoute() []Route {
	return []Route{
		// user
		{Method: "GET", Path: "/user/schedules", Handler: c.UserScheduleHandler.GetSchedules},
		{Method: "POST", Path: "/user/schedules/sync", Handler: c.UserScheduleHandler.SyncSchedule},
		// schedules
		{Method: "POST", Path: "/schedules/sync/all", Handler: c.ScrapedScheduleHandler.SyncSchedules},
	}
}

func (c *Config) SwaggerRoute() {
	r := c.App.Group("/api/v1/docs")
	r.Get("*", swagger.HandlerDefault)
}

func (c *Config) NotFoundRoute() {
	c.App.Use(func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusNotFound).JSON(helper.SingleError("route", "NOT_FOUND"))
	})
}
