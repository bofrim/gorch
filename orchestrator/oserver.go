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

func ServerThread(orchestrator *Orchestrator, ctx context.Context, done func()) {
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
			log.Printf("Orchestrator got pinged by %s. [%s]\n", name, orchestrator.Nodes[name].LastInteraction.String())
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

	app.Post("/:node/action/:action", func(c *fiber.Ctx) error {
		node := c.Params("node")
		action := c.Params("action")
		body := c.Body()
		nodeConn, ok := orchestrator.Nodes[node]
		if !ok {
			c.Response().SetStatusCode(404)
			return c.SendString(fmt.Sprintf("Node %s not registered.", node))
		}

		// Send the request
		out, err := nodeConn.RequestAction(action, body)
		if err != nil {
			c.Response().SetStatusCode(500)
			return c.SendString("Error sending request to node.")
		}
		return c.Send(out)
	})

	app.Get("/:node/data", func(c *fiber.Ctx) error {
		node := c.Params("node")
		body := c.Body()
		nodeConn, ok := orchestrator.Nodes[node]
		if !ok {
			c.Response().SetStatusCode(404)
			return c.SendString(fmt.Sprintf("Node %s not registered.", node))
		}

		out, err := nodeConn.RequestData(body)
		if err != nil {
			c.Response().SetStatusCode(500)
			return c.SendString("Error getting data from node.")
		}
		return c.Send(out)
	})

	err := app.Listen(fmt.Sprintf(":%d", orchestrator.Port))
	log.Println(err)

}
