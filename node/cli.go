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
	Name             string             `yaml:"name"`
	Port             int                `yaml:"port"`
	Host             string             `yaml:"host"`
	Orchestrator     string             `yaml:"orchestrator"`
	Data             string             `yaml:"data"`
	Log              string             `yaml:"log"`
	LogLevel         string             `yaml:"log-level"`
	CertPath         string             `yaml:"cert-path"`
	ArbitraryActions bool               `yaml:"arbitrary-actions"`
	Actions          map[string]*Action `yaml:"actions"`
	ActionGroups     map[string]int     `yaml:"action-groups"`
}

func NewNodeConfig() *NodeConfig {
	return &NodeConfig{
		Name:             fmt.Sprintf("node-%04d", rand.Intn(10000)),
		Port:             443,
		Host:             "127.0.0.1",
		ArbitraryActions: false,
		LogLevel:         "INFO",
		ActionGroups: map[string]int{
			"total":   100,
			"default": 0,
		},
	}
}

func (c *NodeConfig) ReadConfig(path string) error {

	data, err := os.ReadFile(path)
	if err != nil {
		slog.Default().Error("Error loading node config file.", err, slog.String("path", path))
		return err
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		slog.Default().Error("Error parsing node config.", err, slog.String("path", path))
		return err
	}

	// If action names aren't explicitly specified, fill them in
	for name, a := range c.Actions {
		if a.Name == "" {
			a.Name = name
		}
	}
	return nil
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
		},
		Action: func(cCtx *cli.Context) error {
			// Get config file location from cli
			absConfigPath, err := filepath.Abs(cCtx.String("config"))
			if err != nil {
				slog.Default().Error("Can't get abs path for config.", err)
				return err
			}

			// Parse the config file
			config := NewNodeConfig()
			if err := config.ReadConfig(absConfigPath); err != nil {
				slog.Default().Error("Can't read config.", err)
				return err
			}

			// Find the absolute path to the data dir
			absDataPath := ""
			if config.Data != "" {
				absDataPath, _ = filepath.Abs(config.Data)
			}

			// Construct the node
			node := Node{
				Name:             config.Name,
				ServerPort:       config.Port,
				DataDir:          absDataPath,
				Actions:          config.Actions,
				OrchAddr:         config.Orchestrator,
				ArbitraryActions: config.ArbitraryActions,
				MaxNumActions:    config.ActionGroups["total"],
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
