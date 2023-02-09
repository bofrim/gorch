package node

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/bofrim/gorch/utils"
	"github.com/urfave/cli/v2"
)

func GetCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "node",
		Usage: "Run a gorch node",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "port",
				Usage: "Specify a port for the node to serve on",
				Value: 443,
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
				Name:  "data",
				Usage: "Specify a directory to use as the node's data directory",
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
			&cli.BoolFlag{
				Name:        "arbitrary-actions",
				Usage:       "Allow the node to run arbitrary actions",
				Value:       false,
				DefaultText: "false",
			},
			&cli.StringFlag{
				Name:     "log",
				Usage:    "Specify a path to a file to log to. If not specified, logs will be printed to stdout",
				Required: false,
			},
			&cli.IntFlag{
				Name:  "max-actions",
				Usage: "Specify the number of concurrent actions that can run.",
				Value: 100,
			},
			&cli.StringFlag{
				Name:     "cert-path",
				Usage:    "Specify a path with ssl.crt and ssl.key files.",
				Required: true,
			},
		},
		Action: func(cCtx *cli.Context) error {
			// Parse args and build node
			absDataPath := ""
			if cCtx.String("data") != "" {
				absDataPath, _ = filepath.Abs(cCtx.String("data"))
			}

			absActionPath := ""
			if cCtx.String("actions") != "" {
				absActionPath, _ = filepath.Abs(cCtx.String("actions"))
			}

			node := Node{
				Name:             cCtx.String("name"),
				ServerPort:       cCtx.Int("port"),
				DataDir:          absDataPath,
				ActionsPath:      absActionPath,
				OrchAddr:         cCtx.String("orchestrator"),
				ArbitraryActions: cCtx.Bool("arbitrary-actions"),
				MaxNumActions:    cCtx.Int("max-actions"),
				CertPath:         cCtx.String("cert-path"),
			}

			// Setup logging
			logger, closeFn, err := utils.SetupLogging(node.LogFile)
			if err != nil {
				fmt.Printf("Error while setting up logging; %s\n", err.Error())
				return err
			}
			defer closeFn()

			// Run the node
			if err := node.Run(logger); err != nil {
				logger.Error("Error while running node.", err)
				return err
			}
			return nil
		},
	}
}
