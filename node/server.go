package node

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

func ServerThread(node *Node, ctx context.Context, done func()) {
	defer done()

	// Create a new app
	app := fiber.New()

	// Status endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Gorch is up and running!")
	})

	// Endpoint for interacting with the node's data
	dataEp := app.Group("/data")
	dataEp.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(node.Data)
	})
	dataEp.Get("/list", func(c *fiber.Ctx) error {
		keys := make([]string, len(node.Data))
		i := 0
		for k := range node.Data {
			keys[i] = k
			i++
		}
		return c.JSON(keys)
	})
	dataEp.Get("/list/:file", func(c *fiber.Ctx) error {
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
	dataEp.Get("/:file", func(c *fiber.Ctx) error {
		return c.JSON(node.Data[c.Params("file")])
	})

	// Endpoint for running actions on the node
	actionEp := app.Group("/action")
	actionEp.Get("/", func(c *fiber.Ctx) error {
		// Manually marshal to pickup tag names
		s, err := json.Marshal(node.Actions)
		if err != nil {
			return err
		}
		return c.SendString(string(s))
	})
	actionEp.Get("/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		params := make(map[string]string)
		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			params[string(key)] = string(value)
		})
		action := node.Actions[name]
		s, err := action.Run(params)
		if err != nil {
			return err
		}
		return c.SendString(s)
	})

	// Run the App
	app.Listen(":3000")
}

type Empty struct{}
