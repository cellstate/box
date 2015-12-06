package command

import (
	"github.com/codegangsta/cli"
)

var RmAction = func(ctx *cli.Context) error {
	return nil
}

var Rm = cli.Command{
	Name:   "rm",
	Usage:  "permanently remove remote data pointed to by the content hash",
	Flags:  []cli.Flag{},
	Action: errAction(RmAction),
}
