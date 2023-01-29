package user

import (
	"encoding/json"
	"fmt"

	"github.com/bofrim/gorch/orchestrator"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var infoCommand = cli.Command{
	Name:  "info",
	Usage: "Get info about the orchestrator and its nodes",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "host",
			Usage:       "Specify the address of the gorch orchestrator",
			Value:       "127.0.0.1",
			DefaultText: "localhost",
		},
		&cli.IntFlag{
			Name:        "port",
			Usage:       "Specify the port of the gorch system",
			Value:       8322,
			DefaultText: "8322",
			Action: func(ctx *cli.Context, v int) error {
				if v >= 65536 {
					return fmt.Errorf("flag port value %v out of range [0-65535]", v)
				}
				return nil
			},
		},
		&cli.BoolFlag{
			Name:    "json",
			Aliases: []string{"j"},
			Usage:   "Specify if the output should be in JSON format",
			Value:   false,
		},
	},
	Action: func(c *cli.Context) error {
		raw, err := GetNodes(c.String("host"), c.Int("port"))
		if err != nil {
			fmt.Printf("Request Error: %s", err)
			return err
		}
		var o map[string]orchestrator.NodeConnection
		err = json.Unmarshal(raw, &o)
		if err != nil {
			fmt.Printf("Unmarshal Error: %s", err)
			return err
		}

		if c.Bool("json") {
			out, err := json.MarshalIndent(o, "", "  ")
			if err != nil {
				fmt.Printf("Marshal Error: %s", err.Error())
				return err
			}
			fmt.Print(string(out))
		} else {
			fmt.Printf("Gorch System Info:\n")
			y, err := yaml.Marshal(&o)
			if err != nil {
				fmt.Printf("Marshal Error: %s", err.Error())
				return err
			}
			fmt.Println(string(y))
		}
		return err
	},
}
