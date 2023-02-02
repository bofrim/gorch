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
			Name:        "orchestrator",
			Usage:       "Specify the address of the gorch orchestrator",
			Value:       "127.0.0.1:8322",
			DefaultText: "localhost:8322",
		},
		&cli.BoolFlag{
			Name:    "json",
			Aliases: []string{"j"},
			Usage:   "Specify if the output should be in JSON format",
			Value:   false,
		},
	},
	Action: func(c *cli.Context) error {
		raw, err := GetNodes(c.String("orchestrator"))
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
