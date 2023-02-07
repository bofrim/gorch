package orchestrator

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func GetCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "orchestrator",
		Usage: "Run a central orchestration server.",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Usage:       "Specify a port for the orchestrator to serve on.",
				Value:       443,
				DefaultText: "443",
				Action: func(ctx *cli.Context, v int) error {
					if v >= 65536 {
						return fmt.Errorf("flag port value %v out of range [0-65535]", v)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:     "log",
				Usage:    "Specify a path to a file to log to. If not specified, logs will be printed to stdout",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "cert-path",
				Usage:    "Specify a path with ssl.crt and ssl.key files",
				Required: true,
			},
		},
		Action: func(cCtx *cli.Context) error {
			fmt.Println("Gorch orchestrator running on port: ", cCtx.Int("port"))
			orchestrator := Orchestrator{
				Port:     cCtx.Int("port"),
				LogFile:  cCtx.String("log"),
				CertPath: cCtx.String("cert-path"),
			}
			return orchestrator.Run()
		},
	}
}
