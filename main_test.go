package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"

	"github.com/cellstate/box/command"
)

func init() {

}

func apply(cmd cli.Command) *flag.FlagSet {
	set := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	for _, f := range cmd.Flags {
		f.Apply(set)
	}

	return set
}

// the basic usage scenario looks like
// the following
func TestMainScenario(t *testing.T) {
	output := bytes.NewBuffer(nil)
	mw := io.MultiWriter(output, os.Stderr)
	command.Clog.SetOutput(mw)
	app := cli.NewApp()

	//init the boxed project
	log.Printf("$> box init")
	set := apply(command.Init)
	err := set.Parse([]string{"-b=abc"})
	assert.NoError(t, err, "Parsing flags should not return err")
	ctx := cli.NewContext(app, set, nil)
	err = command.InitAction(ctx)
	assert.NoError(t, err, "Command should not error")
	assert.Contains(t, output.String(), "abc", "Output should contain bucket uri")

	//push first content
	log.Printf("$> box push")

	log.Printf("$> box push (again)")

	log.Printf("$> box pull")

	log.Printf("$> box pull (again)")

	log.Printf("$> box rm")
}
