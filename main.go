package main

import (
	"os"

	"github.com/codegangsta/cli"

	"github.com/cellstate/box/command"
)

func main() {
	app := cli.NewApp()
	app.Name = "box"
	app.Usage = "drop it like its hot"
	app.Flags = []cli.Flag{}

	app.Commands = []cli.Command{
		command.Init,
		command.Push,
		command.Pull,
		command.Rm,
	}

	app.Run(os.Args)
}
