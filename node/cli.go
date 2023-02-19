package node

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/bofrim/gorch/utils"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

type NodeConfig struct {
	Name             string `yaml:"name"`
	Port             int    `yaml:"port"`
	Host             string `yaml:"host"`
	Orchestrator     string `yaml:"orchestrator"`
	Data             string `yaml:"data"`
	Actions          string `yaml:"actions"`
	ArbitraryActions bool   `yaml:"arbitrary-actions"`
	Log              string `yaml:"log"`
	LogLevel         string `yaml:"log-level"`
	MaxNumActions    int    `yaml:"max-actions"`
	CertPath         string `yaml:"cert-path"`
}

func NewNodeConfig() *NodeConfig {
	return &NodeConfig{
		Name:             fmt.Sprintf("node-%04d", rand.Intn(10000)),
		Port:             443,
		Host:             "127.0.0.1",
		ArbitraryActions: false,
		MaxNumActions:    1000,
		LogLevel:         "INFO",
	}
}

func GetCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "node",
		Usage: "Run a gorch node",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Usage:    "specify a config file for the node",
				Required: true,
				Action: func(ctx *cli.Context, v string) error {
					if _, err := os.Stat(v); os.IsNotExist(err) {
						return fmt.Errorf("config file %v does not exist", v)
					}
					return nil
				},
			},
			// &cli.IntFlag{
			// 	Name:  "port",
			// 	Usage: "Specify a port for the node to serve on",
			// 	Value: 443,
			// 	Action: func(ctx *cli.Context, v int) error {
			// 		if v >= 65536 {
			// 			return fmt.Errorf("flag port value %v out of range [0-65535]", v)
			// 		}
			// 		return nil
			// 	},
			// },
			// &cli.StringFlag{
			// 	Name:  "host",
			// 	Usage: "Specify the address the node will be accessible at",
			// 	Value: "127.0.0.1",
			// },
			// &cli.StringFlag{
			// 	Name:  "data",
			// 	Usage: "Specify a directory to use as the node's data directory",
			// 	Action: func(ctx *cli.Context, v string) error {
			// 		if _, err := os.Stat(v); os.IsNotExist(err) {
			// 			return fmt.Errorf("data directory %v does not exist", v)
			// 		}
			// 		return nil
			// 	},
			// },
			// &cli.StringFlag{
			// 	Name:     "actions",
			// 	Usage:    "Specify a path to a file containing the node's actions",
			// 	Required: false,
			// 	Action: func(ctx *cli.Context, v string) error {
			// 		if _, err := os.Stat(v); os.IsNotExist(err) {
			// 			return fmt.Errorf("actions file %v does not exist", v)
			// 		}
			// 		return nil
			// 	},
			// },
			// &cli.StringFlag{
			// 	Name:     "orchestrator",
			// 	Usage:    "Specify a main server to connect this node to",
			// 	Required: false,
			// },
			// &cli.StringFlag{
			// 	Name:        "name",
			// 	Usage:       "Specify a name of the node.",
			// 	Value:       "Anon",
			// 	DefaultText: "Anon",
			// 	Action: func(ctx *cli.Context, v string) error {
			// 		pattern := "^[a-zA-Z][a-zA-Z0-9_.-]*$"
			// 		match, _ := regexp.MatchString(pattern, v)
			// 		if !match {
			// 			log.Printf("the string %s cannot be used as a node name. Ensure it matches this pattern: %s", v, pattern)
			// 			return fmt.Errorf("the string %s cannot be used as a node name. Ensure it matches this pattern: %s", v, pattern)
			// 		}
			// 		return nil
			// 	},
			// },
			// &cli.BoolFlag{
			// 	Name:        "arbitrary-actions",
			// 	Usage:       "Allow the node to run arbitrary actions",
			// 	Value:       false,
			// 	DefaultText: "false",
			// },
			// &cli.StringFlag{
			// 	Name:     "log",
			// 	Usage:    "Specify a path to a file to log to. If not specified, logs will be printed to stdout",
			// 	Required: false,
			// },
			// &cli.IntFlag{
			// 	Name:  "max-actions",
			// 	Usage: "Specify the number of concurrent actions that can run.",
			// 	Value: 100,
			// },
			// &cli.StringFlag{
			// 	Name:     "cert-path",
			// 	Usage:    "Specify a path with ssl.crt and ssl.key files.",
			// 	Required: true,
			// },
		},
		Action: func(cCtx *cli.Context) error {
			// Get config file location from cli
			absConfigPath := ""
			if cCtx.String("config") != "" {
				absConfigPath, _ = filepath.Abs(cCtx.String("config"))
			}

			// Read config file from disk
			configFile, err := os.ReadFile(absConfigPath)
			if err != nil {
				slog.Default().Error("Error loading node config file.", err, slog.String("path", absConfigPath))
				return err
			}

			// Parse the config file
			config := NewNodeConfig()
			if err := yaml.Unmarshal(configFile, &config); err != nil {
				slog.Default().Error("Error parsing node config.", err, slog.String("path", absConfigPath))
				return err
			}

			// Find the absolute path to the data dir
			absDataPath := ""
			if config.Data != "" {
				absDataPath, _ = filepath.Abs(config.Data)
			}
			// Find the absolute path to the actions file
			absActionPath := ""
			if config.Actions != "" {
				absActionPath, _ = filepath.Abs(config.Actions)
			}

			// Construct the node
			node := Node{
				Name:             config.Name,
				ServerPort:       config.Port,
				DataDir:          absDataPath,
				ActionsPath:      absActionPath,
				OrchAddr:         config.Orchestrator,
				ArbitraryActions: config.ArbitraryActions,
				MaxNumActions:    config.MaxNumActions,
				CertPath:         config.CertPath,
			}

			// Setup logging
			var logLevel slog.Level
			logLevel.UnmarshalText([]byte(config.LogLevel))
			logger, closeFn, err := utils.SetupLogging(node.LogFile, logLevel)
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
