package hook

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const HookListenIdleTimeout = 10 * time.Second
const HookListenShutdownTimeout = 500 * time.Millisecond

type HookListener struct{ ticker *time.Ticker }

func NewHookListener() HookListener {
	return HookListener{
		ticker: time.NewTicker(HookListenIdleTimeout),
	}
}

func (h *HookListener) Listen(port int) error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	done := func() {
		wg.Done()
		cancel()
	}

	wg.Add(1)
	go HookServerThread(h, port, ctx, done)

	wg.Add(1)
	go h.WatchDog(ctx, done)

	wg.Wait()
	cancel()
	return nil
}

func (h *HookListener) WatchDog(ctx context.Context, done func()) {
	defer done()
	for {
		select {
		case <-h.ticker.C:
			log.Println("Hook Watch Dog tripped!")
			return
		case <-ctx.Done():
			return
		}
	}
}

func (h *HookListener) Feed() {
	h.ticker.Reset(HookListenIdleTimeout)
}

func HookServerThread(h *HookListener, port int, ctx context.Context, done func()) {
	defer done()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	// Ensure a graceful shutdown of the server
	go func() {
		<-ctx.Done()
		app.Shutdown()
	}()

	app.Post("/update", func(c *fiber.Ctx) error {
		body := c.Body()
		fmt.Println(string(body))
		return c.SendString("ack")
	})
	app.Post("/keepalive", func(c *fiber.Ctx) error {
		log.Println("Got Keepalive request.")
		h.Feed()
		return c.SendString("ack")
	})

	app.Post("/finish", func(c *fiber.Ctx) error {
		go func() {
			app.ShutdownWithTimeout(3 * time.Second)
		}()
		return c.SendString("ack")
	})
	app.Listen(fmt.Sprintf(":%d", port))
}
