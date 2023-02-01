package user

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
)

const dataRegexPattern = `^[a-zA-Z][a-zA-Z0-9]*=[a-zA-Z0-9"'!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]*$`

var actionCommand = cli.Command{
	Name:      "action",
	Usage:     "Perform an action on a node.",
	ArgsUsage: "node action [options]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "node",
			Usage:    "Specify the node to perform the action on.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "action",
			Usage:    "Specify the action to perform on the node.",
			Required: true,
		},
		&cli.StringFlag{
			Name:        "host",
			Usage:       "Specify the address of the orchestrator.",
			Value:       "127.0.0.1",
			DefaultText: "localhost",
		},
		&cli.IntFlag{
			Name:        "port",
			Usage:       "Specify the port of the orchestrator.",
			Value:       8322,
			DefaultText: "8322",
			Action: func(ctx *cli.Context, v int) error {
				if v >= 65536 {
					return fmt.Errorf("flag port value %v out of range [0-65535]", v)
				}
				return nil
			},
		},
		&cli.StringSliceFlag{
			Name:  "data",
			Usage: "Pass data in key=value format",
			Action: func(ctx *cli.Context, v []string) error {
				// check if the data is in key=value format
				dataRegex := regexp.MustCompile(dataRegexPattern)
				data := make(map[string]string)
				for _, d := range v {
					if !dataRegex.MatchString(d) {
						err := fmt.Errorf("data %v is not in key=value format", d)
						fmt.Println(err.Error())
						return err
					}
					data[strings.Split(d, "=")[0]] = strings.Split(d, "=")[1]
				}

				return nil
			},
		},
		&cli.IntFlag{
			Name:  "stream-port",
			Usage: "A port to use to stream the response from the action.",
			Value: 0,
			Action: func(ctx *cli.Context, v int) error {
				if v >= 65536 {
					return fmt.Errorf("flag stream-port value %v out of range [0-65535]", v)
				}
				return nil
			},
		},
	},
	Action: func(ctx *cli.Context) error {
		addr := ctx.String("host")
		port := ctx.Int("port")
		action := ctx.String("action")
		node := ctx.String("node")
		streamPort := ctx.Int("stream-port")

		// Parse data
		data := make(map[string]string)
		dataRegex := regexp.MustCompile(dataRegexPattern)
		for _, d := range ctx.StringSlice("data") {
			if !dataRegex.MatchString(d) {
				err := fmt.Errorf("data %v is not in key=value format", d)
				fmt.Println(err.Error())
				return err
			}
			data[strings.Split(d, "=")[0]] = strings.Split(d, "=")[1]
		}

		var runErr error
		if streamPort != 0 {
			runErr = StreamAction(addr, port, node, streamPort, action, data)
		} else {
			runErr = RunAction(addr, port, node, action, data)
		}
		if runErr != nil {
			fmt.Printf("Action Error: %v", runErr)
			return runErr
		}
		return nil
	},
}