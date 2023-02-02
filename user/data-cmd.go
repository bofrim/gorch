package user

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var dataRequestCommand = cli.Command{
	Name:  "data",
	Usage: "Request data from a node.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "orchestrator",
			Usage:       "Specify the address of the gorch orchestrator",
			Value:       "127.0.0.1:8322",
			DefaultText: "localhost:8322",
		},
		&cli.StringFlag{
			Name:     "node",
			Usage:    "Specify the node to request data from.",
			Required: true,
		},
		&cli.StringFlag{
			Name:        "path",
			Usage:       "Specify the path to the data to request.",
			Value:       "",
			DefaultText: "all data",
		},
		&cli.BoolFlag{
			Name:  "json",
			Usage: "Specify if the output should be in JSON format.",
			Value: false,
		},
	},
	Action: func(ctx *cli.Context) error {
		// Send the request
		raw, err := RequestData(ctx.String("orchestrator"), ctx.String("node"), ctx.String("path"))
		if err != nil {
			log.Printf("error requesting data: %v", err)
			return err
		}

		// Process the response
		var out string

		// Unmarshal the raw data into a map
		o := make(map[string]interface{})
		err = json.Unmarshal(raw, &o)
		if err != nil {
			log.Printf("error unmarshalling data: %v", err)
			return err
		}

		// Marshal the map into either JSON or YAML
		if ctx.Bool("json") {
			j, err := json.MarshalIndent(o, "", "  ")
			if err != nil {
				fmt.Printf("Marshal Error: %s", err.Error())
				return err
			}
			out = string(j)
		} else {
			y, err := yaml.Marshal(&o)
			if err != nil {
				fmt.Printf("Marshal Error: %s", err.Error())
				return err
			}
			out = string(y)
		}

		// Print the output
		fmt.Print(out)
		return nil
	},
}

var dataListCommand = cli.Command{
	Name:  "list",
	Usage: "List the data available from a node.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "orchestrator",
			Usage:       "Specify the address of the gorch orchestrator",
			Value:       "127.0.0.1:8322",
			DefaultText: "localhost:8322",
		},
		&cli.StringFlag{
			Name:     "node",
			Usage:    "Specify the node to request data from.",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "path",
			Usage: "Specify the path to the data to request.",
			Value: "",
		},
		&cli.BoolFlag{
			Name:  "json",
			Usage: "Specify if the output should be in JSON format.",
			Value: false,
		},
	},
	Action: func(ctx *cli.Context) error {
		// Send the request
		raw, err := RequestDataList(ctx.String("orchestrator"), ctx.String("node"), ctx.String("path"))
		if err != nil {
			return err
		}

		// Process the response
		out := string(raw)
		fmt.Println(out)

		return nil
	},
}
