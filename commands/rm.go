package commands

import (
	"github.com/codegangsta/cli"
)

var Rm = cli.Command{
	Name:  "rm",
	Usage: "permanently remove remote data pointed to by the content hash",
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) {

	},
}
