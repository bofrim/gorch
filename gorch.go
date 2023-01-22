package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/bofrim/gorch/node"
	"github.com/bofrim/gorch/orch"
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
								return fmt.Errorf("data directory %v does not exist", v)
							}
							return nil
						},
					},
					&cli.StringFlag{
						Name:     "actions",
						Aliases:  []string{"a"},
						Usage:    "Specify a path to a file containing the node's actions",
						Required: false,
						Action: func(ctx *cli.Context, v string) error {
							if _, err := os.Stat(v); os.IsNotExist(err) {
								return fmt.Errorf("actions file %v does not exist", v)
							}
							return nil
						},
					},
					&cli.StringFlag{
						Name:     "orch",
						Aliases:  []string{"o"},
						Usage:    "Specify a main server to connect this node to",
						Required: false,
					},
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "Specify a name of the node.",
						Value:       "Anon",
						DefaultText: "Anon",
						Action: func(ctx *cli.Context, v string) error {
							pattern := "^[a-zA-Z][a-zA-Z0-9_.-]+$"
							match, _ := regexp.MatchString(pattern, v)
							if !match {
								log.Printf("the string %s cannot be used as a node name. Ensure it matches %s", v, pattern)
								return fmt.Errorf("the string %s cannot be used as a node name. Ensure it matches %s", v, pattern)
							}
							return nil
						},
					},
				},
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Gorch node running on port: ", cCtx.Int("port"))
					absDataPath, _ := filepath.Abs(cCtx.String("data"))
					absActionPath := ""
					if cCtx.String("actions") != "" {
						absActionPath, _ = filepath.Abs(cCtx.String("actions"))
					}

					node := node.Node{
						Name:        cCtx.String("name"),
						Port:        cCtx.Int("port"),
						DataDir:     absDataPath,
						ActionsPath: absActionPath,
						OrchAddr:    cCtx.String("server"),
					}
					fmt.Println("Gorch node running with name: ", cCtx.String("name"))
					fmt.Println("Gorch node data directory: ", absDataPath)
					fmt.Println("Gorch node actions path: ", absActionPath)
					return node.Run()
				},
			},
			{
				Name:  "orch",
				Usage: "Run a central orchestration server.",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "port",
						Aliases:     []string{"p"},
						Usage:       "Specify a port for the orchestrator to serve on",
						Value:       8322,
						DefaultText: "8322",
						Action: func(ctx *cli.Context, v int) error {
							if v >= 65536 {
								return fmt.Errorf("flag port value %v out of range [0-65535]", v)
							}
							return nil
						},
					},
				},
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Gorch orchestrator running on port: ", cCtx.Int("port"))
					orch := orch.Orch{
						Port: cCtx.Int("port"),
					}
					return orch.Run()
				},
			},
		},
	}

	app.Run(os.Args)
}
