package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type NodeRegistration struct {
	NodeName string `json:"name"`
	NodeAddr string `json:"addr"`
	NodePort int    `json:"port"`
}

func OServerThread(orchestrator *Orchestrator, ctx context.Context, done func()) {
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
			// Was already register
			log.Println("Was already registered.")
		} else {
			conn := NodeConnection{
				Name:            r.NodeName,
				Address:         r.NodeAddr,
				Port:            r.NodePort,
				LastInteraction: time.Now(),
			}
			orchestrator.Nodes[r.NodeName] = &conn
		}
		log.Printf("Orchestrator now has %d nodes registered.\n", len(orchestrator.Nodes))
		return nil
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

	// Forward a get request to a node
	app.Get("/:node/*", func(c *fiber.Ctx) error {
		node := c.Params("node")
		nodeConn, ok := orchestrator.Nodes[node]
		if !ok {
			c.Response().SetStatusCode(404)
			return c.SendString(fmt.Sprintf("Node %s not registered.", node))
		}

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

		nodeUrl := fmt.Sprintf("http://%s:%d/%s", nodeConn.Address, nodeConn.Port, c.Params("*"))
		return c.Redirect(nodeUrl, fiber.StatusTemporaryRedirect)
	})

	err := app.Listen(fmt.Sprintf(":%d", orchestrator.Port))
	log.Println(err)
}
