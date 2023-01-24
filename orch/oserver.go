package orch

import (
	"context"
	"fmt"
	"log"
	"time"

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

	app.Post("/register/:name/:addr?", func(c *fiber.Ctx) error {
		name := c.Params("name")
		_, ok := orch.Nodes[name]
		if ok {
			// Was already register
			log.Println("Was already registered.")
		} else {
			conn := NodeConnection{
				Name:            name,
				Address:         c.Params("addr", c.IP()),
				LastInteraction: time.Now(),
			}
			orch.Nodes[name] = conn
		}
		log.Printf("Orch now has %d nodes registered.\n", len(orch.Nodes))
		return nil
	})
	app.Get("/nodes", func(c *fiber.Ctx) error {
		nodes := make([]string, len(orch.Nodes))
		i := 0
		for k := range orch.Nodes {
			nodes[i] = k
			i++
		}
		return c.JSON(orch.Nodes)
	})

	err := app.Listen(fmt.Sprintf(":%d", orch.Port))
	log.Println(err)

}
