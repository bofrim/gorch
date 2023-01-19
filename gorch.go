package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bofrim/gorch/node"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                   "gorch",
		UseShortOptionHandling: true,
		Version:                "0.0.1",
		Usage:                  "A utility for orchestrating multiple remote nodes.",
		Commands: []*cli.Command{
			{
				Name:  "node",
				Usage: "Run a gorch node",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "port",
						Aliases:     []string{"p"},
						Usage:       "Specify a port for the node to serve on",
						Value:       8321,
						DefaultText: "8321",
						Action: func(ctx *cli.Context, v int) error {
							if v >= 65536 {
								return fmt.Errorf("flag port value %v out of range [0-65535]", v)
							}
							return nil
						},
					},
					&cli.StringFlag{
						Name:     "data",
						Aliases:  []string{"d"},
						Usage:    "Specify a directory to use as the node's data directory",
						Required: true,
						Action: func(ctx *cli.Context, v string) error {
							if _, err := os.Stat(v); os.IsNotExist(err) {
								return fmt.Errorf("flag data value %v does not exist", v)
							}
							return nil
						},
					},
				},
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Gorch node running on port: ", cCtx.Int("port"))
					absPath, _ := filepath.Abs(cCtx.String("data"))
					fmt.Println("Gorch node data directory: ", absPath)
					node := node.Node{
						Port:    cCtx.Int("port"),
						DataDir: absPath,
					}
					return node.Run()
				},
			},
		},
	}

	app.Run(os.Args)
}
