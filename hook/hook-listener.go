package hook

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

const HookListenIdleTimeout = 10 * time.Second
const HookListenShutdownTimeout = 500 * time.Millisecond

type HookListener struct{}

func (h *HookListener) Listen(port int) error {
	app := fiber.New(fiber.Config{
		IdleTimeout:           HookListenIdleTimeout, // TODO: This is not working; find a better way
		DisableStartupMessage: true,
	})
	app.Post("/update", func(c *fiber.Ctx) error {
		body := c.Body()
		fmt.Println(string(body))
		return c.SendString("ack")
	})
	app.Post("/keepalive", func(c *fiber.Ctx) error {
		log.Println("Got Keepalive request.")
		return c.SendString("ack")
	})

	app.Post("/finish", func(c *fiber.Ctx) error {
		go func() {
			app.ShutdownWithTimeout(3 * time.Second)
		}()
		return c.SendString("ack")
	})
	return app.Listen(fmt.Sprintf(":%d", port))
}
