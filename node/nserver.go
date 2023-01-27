package node

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ServerThread(node *Node, ctx context.Context, done func()) {
	defer done()

	// Create a new app
	app := fiber.New()
	go func() {
		<-ctx.Done()
		app.Shutdown()
	}()

	// Status endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		log.Println("Status Checked")
		return c.SendString("Gorch node is up and running!")
	})

	// Endpoint for interacting with the node's data
	dataEp := app.Group("/data")
	dataEp.Get("/", func(c *fiber.Ctx) error {
		log.Println("Get all data")
		return c.JSON(node.Data)
	})
	dataEp.Get("/list", func(c *fiber.Ctx) error {
		log.Println("List data")
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
		log.Printf("List data file: %s\n", file)
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
		file := c.Params("file")
		log.Printf("Get data file: %s\n", file)
		return c.JSON(node.Data[file])
	})

	// Endpoint for running actions on the node
	actionEp := app.Group("/action")
	actionEp.Get("/", func(c *fiber.Ctx) error {
		log.Println("Get available actions.")
		// Manually marshal to pickup tag names
		s, err := json.Marshal(node.Actions)
		if err != nil {
			return err
		}
		return c.SendString(string(s))
	})
	actionEp.Post("/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		log.Printf("Run action %s\n", name)
		var body map[string]string
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(http.StatusBadRequest).Send([]byte(err.Error()))
		}
		action := node.Actions[name]
		outputs, err := action.Run(body)
		if err != nil {
			return err
		}
		out := strings.Join(outputs, "\n")
		return c.SendString(out)
	})
	app.Post(("/reload/"), func(c *fiber.Ctx) error {
		log.Printf("Reload actions\n")
		if _, err := os.Stat(node.ActionsPath); os.IsNotExist(err) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		node.ReloadActions(node.ActionsPath)
		return c.SendStatus(fiber.StatusOK)
	})

	// Run the App
	if node.ServerPort == 0 {
		node.ServerPort = 3000
	}
	app.Listen(fmt.Sprintf(":%d", node.ServerPort))
}

type Empty struct{}
