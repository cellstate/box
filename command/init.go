package command

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"

	"github.com/cellstate/box/bucket"
	"github.com/cellstate/box/config"
)

var InitAction = func(ctx *cli.Context) error {
	var err error

	dir := ctx.Args().First()
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to get working directory: %s", err)
		}
	}

	bucketlist := ctx.StringSlice("bucket")
	Clog.Printf("Initializing boxed directory in '%s' with buckets: %v", dir, bucketlist)
	var b bucket.Bucket
	for _, bl := range bucketlist {
		b, err = bucket.Create(bl)
		if err != nil {
			Clog.Printf("Failed to create bucket for uri '%s': %s", bl, err)
			continue
		}

		//@todo support multiple buckets, and select one intelligently instead of the first
		break
	}

	if b == nil {
		return fmt.Errorf("Failed to setup a bucket with any of the given uris: %v", bucketlist)
	}

	conf := config.DefaultConfig()
	conf.Buckets = append(conf.Buckets, b.Config())

	return config.WriteConfig(dir, conf)
}

var Init = cli.Command{
	Name:  "init",
	Usage: "bootstrap a boxed project for the given directory",
	Flags: []cli.Flag{
		cli.StringSliceFlag{Name: "bucket,b", Value: &cli.StringSlice{}, Usage: "the uri to a bucket endpoint"},
	},
	Action: errAction(InitAction),
}
