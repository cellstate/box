package command

import (
	"os"

	"github.com/codegangsta/cli"

	"github.com/cellstate/box/config"
	"github.com/cellstate/box/graph"
	"github.com/cellstate/box/graph/fs"
	"github.com/cellstate/errwrap"
)

var PushAction = func(ctx *cli.Context) error {
	var err error
	dir := ctx.Args().First()
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			return errwrap.Wrapf("Failed to get working directory: {{err}}", err)
		}
	}

	config, err := config.ReadConfig(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to read configuration for directory '%s': {{err}}", err, dir)
	}

	//compute the local graph
	var lgraph graph.Graph
	lgraph, err = fs.FromFS(dir, Clog)
	if err != nil {
		return errwrap.Wrapf("Failed to construct local graph from directory '%s': {{err}}", err, dir)
	}

	//1. compute local graph and compare with remote graph to find missing nodes

	//2. large files are

	_ = lgraph
	_ = config

	return nil
}

var Push = cli.Command{
	Name:   "push",
	Usage:  "publish the content of a boxed directory to its remote(s)",
	Flags:  []cli.Flag{},
	Action: errAction(PullAction),
}
