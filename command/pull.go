package command

import (
	"github.com/codegangsta/cli"
)

var PullAction = func(ctx *cli.Context) error {
	return nil
}

var Pull = cli.Command{
	Name:   "pull",
	Usage:  "retrieve content of a boxed directory from its remote(s) using its hash",
	Flags:  []cli.Flag{},
	Action: errAction(PullAction),
}
