package command

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
)

//All command user this logger, it is made public to
//facilitate assertions in unit tests
var Clog = log.New(os.Stderr, "box: ", 0)

type errFunc func(ctx *cli.Context) error

func errAction(f errFunc) func(ctx *cli.Context) {
	return func(ctx *cli.Context) {
		err := f(ctx)
		if err != nil {
			Clog.Fatal(err)
		}
	}
}
