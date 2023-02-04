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
	"golang.org/x/exp/slog"
)

func NServerThread(node *Node, ctx context.Context, logger *slog.Logger, done func()) {
	defer done()

	// Create a new app
	app := fiber.New()
	go func() {
		<-ctx.Done()
		logger.Info("Server done. Shutting down.")
		app.Shutdown()
	}()

	// Status endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		logger.Debug("Status check")
		return c.SendString("Gorch node is up and running!")
	})

	// Endpoint for interacting with the node's data
	dataEp := app.Group("/data")
	dataEp.Get("/", func(c *fiber.Ctx) error {
		logger.Debug("Get all data")
		return c.JSON(node.Data)
	})
	dataEp.Get("/:file", func(c *fiber.Ctx) error {
		logger.Debug("Get file data", slog.String("file", c.Params("file")))
		return c.JSON(node.Data[c.Params("file")])
	})

	listEp := app.Group("/list")
	listEp.Get("/", func(c *fiber.Ctx) error {
		logger.Debug("List data.")
		keys := make([]string, len(node.Data))
		i := 0
		for k := range node.Data {
			keys[i] = k
			i++
		}
		return c.JSON(keys)
	})
	listEp.Get("/:file", func(c *fiber.Ctx) error {
		logger.Debug("List data file.", slog.String("file", c.Params("file")))

		fileData := node.Data[c.Params("file")]
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
		logger.Debug("Get Actions",
			slog.String("actions", fmt.Sprintf("%+v", node.Actions)),
		)
		s, err := json.Marshal(node.Actions)
		if err != nil {
			return err
		}
		return c.SendString(string(s))
	})

	// Experimental allow the user to send an arbitrary action to the node
	// The body is the same as a normal action, but an "action" tag is required
	// with the info for the action that would have been included in the actions definition file
	// TODO: Test, simplify, and establish a better structure
	actionEp.Post("/", func(c *fiber.Ctx) error {
		logger.Debug("Run Adhoc action")
		if !node.ArbitraryActions {
			return c.Status(http.StatusForbidden).SendString("Remote actions are disabled.")
		}
		// Parse out an action from the body
		var body map[string]interface{}
		if c.Body() != nil {
			err := json.Unmarshal(c.Body(), &body)
			if err != nil {
				log.Printf("Error parsing body: %s\n", err.Error())
				return c.Status(http.StatusBadRequest).Send([]byte(err.Error()))
			}
		} else {
			body = map[string]interface{}{}
		}

		// Expect an action definition to be specified in the body
		var adhocAction AdHocAction
		actionDefErr := json.Unmarshal(c.Body(), &adhocAction)
		if actionDefErr != nil {
			log.Printf("Error parsing body: %s\n", actionDefErr.Error())
			return c.Status(http.StatusBadRequest).Send([]byte(actionDefErr.Error()))
		}

		// Run the action
		action := adhocAction.ActionDef
		var out string
		if body["stream_addr"] != "" && body["stream_port"] != "" {
			sAddr := body["stream_addr"].(string)
			if sAddr == "loopback" {
				sAddr = c.IP()
			}
			sPort, convErr := strconv.Atoi(body["stream_port"].(string))
			if convErr != nil {
				return c.Status(http.StatusBadRequest).Send([]byte("Invalid stream port"))
			}
			go action.RunStreamed(sAddr, sPort, body, logger)
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

	actionEp.Post("/:name", func(c *fiber.Ctx) error {
		logger.Debug("Run action", slog.String("action", c.Params("name")))
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
		action, ok := node.Actions[name]
		if !ok {
			err := fmt.Errorf("unable to find action '%s'", name)
			logger.Error("No action found", err, slog.String("action", name))
			return err
		}

		var out string
		if body["stream_addr"] != "" && body["stream_port"] != "" {
			logger.Debug("Action is streamed", slog.String("action", c.Params("name")))
			sAddr := body["stream_addr"]
			if sAddr == "loopback" {
				sAddr = c.IP()
			}
			sPort, convErr := strconv.Atoi(body["stream_port"])
			if convErr != nil {
				return c.Status(http.StatusBadRequest).Send([]byte("Invalid stream port"))
			}
			go action.RunStreamed(sAddr, sPort, body, logger)
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
		logger.Debug("Reload actions.")
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
	logger.Debug("Starting server.")
	app.Listen(fmt.Sprintf(":%d", node.ServerPort))
}
