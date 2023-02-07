package orchestrator

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slog"
)

type NodeRegistration struct {
	NodeName string `json:"name"`
	NodeAddr string `json:"addr"`
	NodePort int    `json:"port"`
}

func OServerThread(orchestrator *Orchestrator, ctx context.Context, logger *slog.Logger, done func()) {
	defer done()

	app := fiber.New()
	go func() {
		<-ctx.Done()
		app.Shutdown()
	}()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Gorch node is up and running!")
	})

	app.Post("/register", func(c *fiber.Ctx) error {
		log.Println("Registration request.")
		r := new(NodeRegistration)
		if err := c.BodyParser(r); err != nil {
			log.Printf("Unable to parse: %s", err)
			log.Printf("Body: %s", c.Body())
			return err
		}
		_, ok := orchestrator.Nodes[r.NodeName]
		if ok {
			logger.Info("Node already registered.",
				slog.String("node", r.NodeName),
				slog.Int("num_nodes", len(orchestrator.Nodes)),
			)
			return nil
		} else {
			conn := NodeConnection{
				Name:            r.NodeName,
				Address:         r.NodeAddr,
				Port:            r.NodePort,
				LastInteraction: time.Now(),
			}
			orchestrator.Nodes[r.NodeName] = &conn
			log.Printf("Orchestrator now has %d nodes registered.\n", len(orchestrator.Nodes))
			logger.Info("Registered node.",
				slog.String("node", r.NodeName),
				slog.Int("num_nodes", len(orchestrator.Nodes)),
			)
			return nil
		}
	})
	app.Post("/ping/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		node, ok := orchestrator.Nodes[name]
		if ok {
			node.LastInteraction = time.Now()
		} else {
			log.Printf("Orchestrator got pinged by %s, but it was not registered.\n", name)
			c.Response().SetStatusCode(404)
			return c.SendString("Node not registered.")
		}
		return nil

	})
	app.Get("/nodes", func(c *fiber.Ctx) error {
		nodes := make([]string, len(orchestrator.Nodes))
		i := 0
		for k := range orchestrator.Nodes {
			nodes[i] = k
			i++
		}
		return c.JSON(orchestrator.Nodes)
	})

	app.Get("/:node/*", func(c *fiber.Ctx) error {
		node := c.Params("node")
		nodeConn, ok := orchestrator.Nodes[node]
		if !ok {
			c.Response().SetStatusCode(404)
			return c.SendString(fmt.Sprintf("Node %s not registered.", node))
		}

		logger.Info("Redirecting get request.", slog.String("node", nodeConn.Name), slog.String("params", c.Params("*")))
		nodeUrl := fmt.Sprintf("http://%s:%d/%s", nodeConn.Address, nodeConn.Port, c.Params("*"))
		return c.Redirect(nodeUrl, fiber.StatusTemporaryRedirect)
	})

	app.Post("/:node/*", func(c *fiber.Ctx) error {
		node := c.Params("node")
		nodeConn, ok := orchestrator.Nodes[node]
		if !ok {
			c.Response().SetStatusCode(404)
			return c.SendString(fmt.Sprintf("Node %s not registered.", node))
		}

		logger.Info("Redirecting post request.", slog.String("node", nodeConn.Name), slog.String("params", c.Params("*")))
		nodeUrl := fmt.Sprintf("http://%s:%d/%s", nodeConn.Address, nodeConn.Port, c.Params("*"))
		return c.Redirect(nodeUrl, fiber.StatusTemporaryRedirect)
	})

	if orchestrator.CertPath != "" {
		// Create tls certificate
		cer, err := tls.LoadX509KeyPair(
			fmt.Sprintf("%s/ssl.crt", orchestrator.CertPath),
			fmt.Sprintf("%s/ssl.key", orchestrator.CertPath),
		)
		if err != nil {
			log.Fatal(err)
		}

		config := &tls.Config{
			Certificates:       []tls.Certificate{cer},
			InsecureSkipVerify: true,
		}
		ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", orchestrator.Port), config)
		if err != nil {
			panic(err)
		}

		log.Fatal(app.Listener(ln))
	} else {
		err := app.Listen(fmt.Sprintf(":%d", orchestrator.Port))
		log.Println(err)
	}
}
