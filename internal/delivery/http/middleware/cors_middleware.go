package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsMiddleware(a *fiber.App) {
	a.Use(cors.New(cors.Config{
		AllowOrigins: "https://simeru.vercel.app, https://*.svrz.xyz",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET,POST,OPTIONS",
	}))
}
