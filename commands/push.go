package commands

import (
	"github.com/codegangsta/cli"
)

var Push = cli.Command{
	Name:  "push",
	Usage: "publish the content of a boxed directory to its remote(s)",
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) {

	},
}
