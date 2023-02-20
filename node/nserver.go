package node

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	actionEp.Post("/", func(c *fiber.Ctx) error {
		logger.Debug("Run Adhoc action")

		// Ensure arbitrary actions are allowed
		if !node.ArbitraryActions {
			logger.Debug("Adhoc action not enabled.")
			return c.Status(http.StatusForbidden).SendString("Remote actions are disabled.")
		}

		// Parse the info from the request
		body, sDest, err := parseActionBody(c)
		if err != nil {
			logger.Error("Failed to parse body for adhoc", err)
			return c.Status(http.StatusBadRequest).Send([]byte(err.Error()))
		}

		// Expect an action definition to be specified in the body
		var adhocAction AdHocAction
		if err := json.Unmarshal(c.Body(), &adhocAction); err != nil {
			logger.Error("Failed to parse adhoc action definition", err)
			return c.Status(http.StatusBadRequest).Send([]byte(err.Error()))
		}
		action := adhocAction.ActionDef

		// Run the action
		out, ok, err := node.RunAction(&action, sDest, body, logger)
		if !ok {
			return c.Status(fiber.StatusServiceUnavailable).SendString(
				fmt.Sprintf("%d actions already running", node.MaxNumActions),
			)
		}
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString(out)
	})

	actionEp.Post("/:name", func(c *fiber.Ctx) error {
		logger.Debug("Run action", slog.String("action", c.Params("name")))

		// Parse info from request
		body, sDest, err := parseActionBody(c)
		if err != nil {
			logger.Error("Failed to parse body", err)
			return c.Status(http.StatusBadRequest).Send([]byte(err.Error()))
		}

		// Find the action
		name := c.Params("name")
		action, ok := node.Actions[name]
		if !ok {
			err := fmt.Errorf("unable to find action '%s'", name)
			logger.Error("No action found", err, slog.String("action", name))
			return err
		}

		// Run the action
		out, ok, err := node.RunAction(action, sDest, body, logger)
		if !ok {
			return c.Status(fiber.StatusServiceUnavailable).SendString(
				fmt.Sprintf("%d actions already running", node.MaxNumActions),
			)
		}
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		return c.SendString(out)
	})

	// Run the App
	if node.ServerPort == 0 {
		node.ServerPort = 3000
	}

	if node.CertPath != "" {
		// Create tls certificate
		cer, err := tls.LoadX509KeyPair(
			fmt.Sprintf("%s/ssl.crt", node.CertPath),
			fmt.Sprintf("%s/ssl.key", node.CertPath),
		)
		if err != nil {
			log.Fatal(err)
		}

		config := &tls.Config{
			Certificates:       []tls.Certificate{cer},
			InsecureSkipVerify: true,
		}
		ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", node.ServerPort), config)
		if err != nil {
			panic(err)
		}

		log.Fatal(app.Listener(ln))
	} else {
		err := app.Listen(fmt.Sprintf(":%d", node.ServerPort))
		log.Println(err)
	}

	logger.Debug("Starting server.")
	app.Listen(fmt.Sprintf(":%d", node.ServerPort))
}

func parseActionBody(c *fiber.Ctx) (body map[string]string, sDest string, err error) {
	body = map[string]string{}
	if c.Body() != nil {
		var m map[string]interface{}
		err := json.Unmarshal(c.Body(), &m)
		if err != nil {
			return nil, sDest, err
		}
		for k, v := range m {
			// Skip the "action"; it will be dealt with elsewhere
			if k != "action" {
				body[k] = v.(string)
			}
		}
	}
	if body["stream_addr"] != "" && body["stream_port"] != "" {
		sAddr := body["stream_addr"]
		if sAddr == "loopback" {
			sAddr = c.IP()
		}
		sPortStr := body["stream_port"]
		sPort, convertErr := strconv.Atoi(sPortStr)
		if convertErr != nil {
			return nil, "", fmt.Errorf("invalid stream port: %s\nbody: %+v", sPortStr, body)
		}
		sDest = fmt.Sprintf("%s:%d", sAddr, sPort)
	}

	return body, sDest, err
}
