package main

import (
	"os"

	"github.com/bofrim/gorch/node"
	"github.com/bofrim/gorch/orchestrator"
	"github.com/bofrim/gorch/user"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                   "gorch",
		UseShortOptionHandling: true,
		Version:                "0.0.1",
		Usage:                  "A utility for orchestrating multiple remote nodes.",
		Commands: []*cli.Command{
			node.GetCliCommand(),
			orchestrator.GetCliCommand(),
			user.GetCliCommand(),
		},
	}

	app.Run(os.Args)
}
