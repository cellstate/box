package main

import (
	"os"

	"github.com/codegangsta/cli"

	"github.com/cellstate/box/commands"
)

func main() {
	app := cli.NewApp()
	app.Name = "box"
	app.Usage = "drop it like its hot"
	app.Flags = []cli.Flag{}

	app.Commands = []cli.Command{
		commands.Init,
		commands.Push,
		commands.Pull,
		commands.Rm,
	}

	app.Run(os.Args)
}
