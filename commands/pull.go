package commands

import (
	"github.com/codegangsta/cli"
)

var Pull = cli.Command{
	Name:  "pull",
	Usage: "retrieve content of a boxed directory from its remote(s) using its hash",
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) {

	},
}
