package node

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

func ServerThread(node *Node, ctx context.Context, done func()) {
	defer done()

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Gorch is up and running!")
	})
	app.Get("/data", func(c *fiber.Ctx) error {
		return c.JSON(node.Data)
	})

	app.Listen(":3000")
}
