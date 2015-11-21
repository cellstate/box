package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
	tmpdir, err := ioutil.TempDir("", "box_test_")
	assert.NoError(t, err, "Creating temporary directory should not fail")
	output := bytes.NewBuffer(nil)
	mw := io.MultiWriter(output, os.Stderr)
	command.Clog.SetOutput(mw)
	app := cli.NewApp()

	//init the boxed project
	log.Printf("$> box init")
	set := apply(command.Init)
	err = set.Parse([]string{"-b=abc", tmpdir})
	assert.NoError(t, err, "Parsing flags should not return err")
	ctx := cli.NewContext(app, set, nil)
	err = command.InitAction(ctx)
	assert.NoError(t, err, "Command should not error")
	assert.Contains(t, output.String(), "abc", "Output should contain bucket uri")
	data, err := ioutil.ReadFile(filepath.Join(tmpdir, ".box", "config"))
	assert.NoError(t, err, "Should be able to read config file")
	assert.Contains(t, string(data), "abc", "Config file should contain bucket endpoint")

	//push first content
	log.Printf("$> box push")

	log.Printf("$> box push (again)")

	log.Printf("$> box pull")

	log.Printf("$> box pull (again)")

	log.Printf("$> box rm")
}
