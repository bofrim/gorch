package node

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func NServerThread(node *Node, ctx context.Context, done func()) {
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
	dataEp.Get("/:file", func(c *fiber.Ctx) error {
		file := c.Params("file")
		log.Printf("Get data file: %s\n", file)
		return c.JSON(node.Data[file])
	})

	listEp := app.Group("/list")
	listEp.Get("/", func(c *fiber.Ctx) error {
		log.Println("List data")
		keys := make([]string, len(node.Data))
		i := 0
		for k := range node.Data {
			keys[i] = k
			i++
		}
		return c.JSON(keys)
	})
	listEp.Get("/:file", func(c *fiber.Ctx) error {
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
		var body map[string]string
		if c.Body() != nil {
			err := json.Unmarshal(c.Body(), &body)
			if err != nil {
				log.Printf("Error parsing body: %s\n", err.Error())
				return c.Status(http.StatusBadRequest).Send([]byte(err.Error()))
			}
		} else {
			body = map[string]string{}
		}
		action := node.Actions[name]

		var out string
		if body["stream_addr"] != "" && body["stream_port"] != "" {
			sAddr := body["stream_addr"]
			if sAddr == "loopback" {
				sAddr = c.IP()
			}
			sPort, convErr := strconv.Atoi(body["stream_port"])
			if convErr != nil {
				return c.Status(http.StatusBadRequest).Send([]byte("Invalid stream port"))
			}
			go action.RunStreamed(sAddr, sPort, body)
			out = fmt.Sprintf("[%s] Streaming to %s:%d", node.Name, sAddr, sPort)
		} else {
			outputs, err := action.Run(body)
			if err != nil {
				return err
			}
			out = strings.Join(outputs, "\n")
		}

		return c.SendString(out)
	})
	app.Post(("/reload/"), func(c *fiber.Ctx) error {
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
