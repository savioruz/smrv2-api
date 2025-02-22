package middleware

import (
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/savioruz/smrv2-api/pkg/helper"
)

var (
	uri = []*regexp.Regexp{}
	ua  = []string{
		"Mozilla",
		"Chrome",
		"Safari",
		"Swagger",
	}
)

func LimiterMiddleware(a *fiber.App) {
	a.Use(limiter.New(
		limiter.Config{
			Next: func(c *fiber.Ctx) bool {
				return whitelistPath(c) || whitelistRequest(c)
			},
			Max:               20,
			Expiration:        30 * time.Second,
			LimiterMiddleware: limiter.SlidingWindow{},
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).JSON(helper.SingleError(
					"rate_limit",
					"TOO_MANY_REQUESTS",
				))
			},
		}))
}

// whitelistPath func for checking the next middleware uri
func whitelistPath(c *fiber.Ctx) bool {
	originalURL := strings.ToLower(c.OriginalURL())

	for _, pattern := range uri {
		if pattern.MatchString(originalURL) {
			return true
		}
	}
	return false
}

// whitelistRequest func for checking the next middleware request
func whitelistRequest(c *fiber.Ctx) bool {
	userAgent := c.Get("User-Agent")
	origin := c.Get("Origin")

	for _, pattern := range ua {
		if strings.Contains(userAgent, pattern) {
			if origin == "" {
				return true
			}
		}
	}

	return false
}
