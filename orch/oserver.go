package orch

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

	app.Post("/register", func(c *fiber.Ctx) error {
		log.Println("Registration request.")
		r := new(NodeRegistration)
		if err := c.BodyParser(r); err != nil {
			log.Printf("Unable to parse: %s", err)
			log.Printf("Body: %s", c.Body())
			return err
		}
		_, ok := orch.Nodes[r.NodeName]
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
			orch.Nodes[r.NodeName] = &conn
		}
		log.Printf("Orch now has %d nodes registered.\n", len(orch.Nodes))
		return nil
	})
	app.Post("/ping/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		node, ok := orch.Nodes[name]
		if ok {
			log.Printf("Orch got pinged by %s. [%s]\n", name, orch.Nodes[name].LastInteraction.String())
			node.LastInteraction = time.Now()
		} else {
			log.Printf("Orch got pinged by %s, but it was not registered.\n", name)
			c.Response().SetStatusCode(404)
			return c.SendString("Node not registered.")
		}
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

	app.Post("/:node/action/:action", func(c *fiber.Ctx) error {
		// We will need to send a request to the node to run the action
		// and then wait for the response.
		node := c.Params("node")
		action := c.Params("action")
		body := c.Body()
		nodeConn, ok := orch.Nodes[node]
		if !ok {
			c.Response().SetStatusCode(404)
			return c.SendString("Node not registered.")
		}

		// Send the request
		out, err := nodeConn.RequestAction(action, body)

		if err != nil {
			c.Response().SetStatusCode(500)
			return c.SendString("Error sending request to node.")
		}
		return c.Send(out)
	})

	err := app.Listen(fmt.Sprintf(":%d", orch.Port))
	log.Println(err)

}
