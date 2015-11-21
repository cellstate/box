package commands

import (
	"os"

	"github.com/codegangsta/cli"
)

var Init = cli.Command{
	Name:  "init",
	Usage: "bootstrap a boxed project for the given directory",
	Flags: []cli.Flag{
		cli.StringSliceFlag{Name: "bucket,b", Value: &cli.StringSlice{}, Usage: "the uri to a bucket endpoint"},
	},
	Action: func(ctx *cli.Context) {
		var err error

		dir := ctx.Args().First()
		if dir == "" {
			dir, err = os.Getwd()
			if err != nil {
				Clog.Fatalf("Failed to get working directory: %s", err)
			}
		}

		buckets := ctx.StringSlice("bucket")

		Clog.Printf("Initiziling boxed directory in '%s' with buckets: %v", dir, buckets)
	},
}
