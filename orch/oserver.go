package orch

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

func ServerThread(orch *Orch, ctx context.Context, done func()) {
	defer done()

	app := fiber.New()
	go func() {
		<-ctx.Done()
		app.Shutdown()
	}()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Gorch node is up and running!")
	})

}
