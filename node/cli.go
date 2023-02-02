package node

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/urfave/cli/v2"
)

func GetCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "node",
		Usage: "Run a gorch node",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
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
				Name:  "host",
				Usage: "Specify the address the node will be accessible at",
				Value: "127.0.0.1",
			},
			&cli.StringFlag{
				Name:     "data",
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
				Name:     "orchestrator",
				Usage:    "Specify a main server to connect this node to",
				Required: false,
			},
			&cli.StringFlag{
				Name:        "name",
				Usage:       "Specify a name of the node.",
				Value:       "Anon",
				DefaultText: "Anon",
				Action: func(ctx *cli.Context, v string) error {
					pattern := "^[a-zA-Z][a-zA-Z0-9_.-]*$"
					match, _ := regexp.MatchString(pattern, v)
					if !match {
						log.Printf("the string %s cannot be used as a node name. Ensure it matches this pattern: %s", v, pattern)
						return fmt.Errorf("the string %s cannot be used as a node name. Ensure it matches this pattern: %s", v, pattern)
					}
					return nil
				},
			},
		},
		Action: func(cCtx *cli.Context) error {
			absDataPath, _ := filepath.Abs(cCtx.String("data"))
			absActionPath := ""
			if cCtx.String("actions") != "" {
				absActionPath, _ = filepath.Abs(cCtx.String("actions"))
			}
			node := Node{
				Name:        cCtx.String("name"),
				ServerPort:  cCtx.Int("port"),
				ServerAddr:  cCtx.String("host"),
				DataDir:     absDataPath,
				ActionsPath: absActionPath,
				OrchAddr:    cCtx.String("orchestrator"),
			}
			return node.Run()
		},
	}
}
