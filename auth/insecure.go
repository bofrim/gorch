package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slog"
)

var store map[string]struct{} = make(map[string]struct{})

func InsecureAuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("X-Authorization")

	if authHeader == "" {
		slog.Default().Debug("Rejecting request because of missing token", slog.String("Host", c.Get("Host")))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Authorization header missing",
		})
	}

	token := strings.Replace(authHeader, "Bearer ", "", 1)
	if _, ok := store[token]; !ok {
		slog.Default().Debug("Rejecting request because of invalid token", slog.String("Host", c.Get("Host")))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid authorization token",
		})
	}

	return c.Next()
}

func AddToken(token string) {
	store[token] = struct{}{}
}
