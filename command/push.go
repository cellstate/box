package command

import (
	"github.com/codegangsta/cli"
)

var PushAction = func(ctx *cli.Context) error {
	return nil
}

var Push = cli.Command{
	Name:   "push",
	Usage:  "publish the content of a boxed directory to its remote(s)",
	Flags:  []cli.Flag{},
	Action: errAction(PullAction),
}
