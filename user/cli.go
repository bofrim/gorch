package user

import (
	"github.com/urfave/cli/v2"
)

func GetCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "user",
		Usage: "A utility for commanding a gorch system.",
		Subcommands: []*cli.Command{
			&infoCommand,
			&actionCommand,
			&dataRequestCommand,
			&dataListCommand,
		},
	}
}
