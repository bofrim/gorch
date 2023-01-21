package node

import (
	"context"
	"encoding/json"

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
	app.Get("/data/list", func(c *fiber.Ctx) error {
		keys := make([]string, len(node.Data))
		i := 0
		for k := range node.Data {
			keys[i] = k
			i++
		}
		return c.JSON(keys)
	})

	app.Get("/data/list/:file", func(c *fiber.Ctx) error {
		file := c.Params("file")
		fileData := node.Data[file]
		keys := make([]string, len(fileData))
		i := 0
		for k := range fileData {
			keys[i] = k
			i++
		}
		return c.JSON(keys)
	})

	app.Get("/data/:file", func(c *fiber.Ctx) error {
		return c.JSON(node.Data[c.Params("file")])
	})

	app.Get("/action", func(c *fiber.Ctx) error {
		// Manually marshal to pickup tag names
		s, err := json.Marshal(node.Actions)
		if err != nil {
			return err
		}
		return c.SendString(string(s))
	})

	app.Listen(":3000")
}
