package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"

	"github.com/cellstate/box/command"
)

func generateProject(t *testing.T, dir string, size int, seed int64) {
	rnd := rand.New(rand.NewSource(seed))

	//write a large file to /
	fpath := filepath.Join(dir, "large_file")
	f, err := os.Open(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(fpath)
			if err != nil {
				t.Fatal(err)
			}

			total := 0
			for {
				n, err := f.Write([]byte{byte(rnd.Intn(256))})
				if err != nil {
					t.Fatal(err)
				}

				total += n
				if total >= size {
					f.Seek(0, 0)
					break
				}
			}

		} else {
			t.Fatal(err)
		}
	}

	defer f.Close()

	//write non-random file to subdir
	err = os.MkdirAll(filepath.Join(dir, "a", "b"), 0777)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(dir, "a", "b", "small_file"), []byte("i'm small"), 0666)
	if err != nil {
		t.Fatal(err)
	}
}

func writeFileAt(t *testing.T, p string, data []byte, pos int64) {
	f, err := os.OpenFile(p, os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	_, err = f.WriteAt(data, pos)
	if err != nil {
		t.Fatal(err)
	}
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
	dir, err := ioutil.TempDir("", "box_test_")
	assert.NoError(t, err, "Creating temporary directory should not fail")
	output := bytes.NewBuffer(nil)
	mw := io.MultiWriter(output, os.Stderr)
	command.Clog.SetOutput(mw)
	app := cli.NewApp()

	//init the boxed project
	log.Printf("$> box init")
	set := apply(command.Init)
	err = set.Parse([]string{"-b=abc", dir})
	assert.NoError(t, err, "Parsing flags should not return err")
	ctx := cli.NewContext(app, set, nil)
	err = command.InitAction(ctx)
	assert.NoError(t, err, "Command should not error")
	assert.Contains(t, output.String(), "abc", "Output should contain bucket uri")
	data, err := ioutil.ReadFile(filepath.Join(dir, ".box", "config"))
	assert.NoError(t, err, "Should be able to read config file")
	assert.Contains(t, string(data), "abc", "Config file should contain bucket endpoint")

	//push first content
	generateProject(t, dir, 10000, 4)
	log.Printf("$> box push")
	set = apply(command.Push)
	err = set.Parse([]string{dir})
	assert.NoError(t, err, "Parsing flags should not return err")
	ctx = cli.NewContext(app, set, nil)
	err = command.PushAction(ctx)
	assert.NoError(t, err, "Command should not error")
	assert.NotContains(t, output.String(), ".box", "Should not travese .box directory")
	assert.Contains(t, output.String(), "0c2be78762d41f24231c83067d19d8f505d0c3d4", "Should contain part hash")
	assert.Contains(t, output.String(), "50dd3df9c5fa56785373b85e4121adccc9b9a849", "Should contain part hash")

	//edit large file partially and create some new files, then push difference
	writeFileAt(t, filepath.Join(dir, "large_file"), []byte("l"), 0)
	log.Printf("$> box push (again)")
	set = apply(command.Push)
	err = set.Parse([]string{dir})
	assert.NoError(t, err, "Parsing flags should not return err")
	ctx = cli.NewContext(app, set, nil)
	err = command.PushAction(ctx)
	assert.NoError(t, err, "Command should not error")
	assert.NotContains(t, output.String(), ".box", "Should not travese .box directory")
	assert.Contains(t, output.String(), "4c0ae6030090de96f22081c36299f893dd83385a", "Should contain part hash")
	assert.Contains(t, output.String(), "50dd3df9c5fa56785373b85e4121adccc9b9a849", "Should contain part hash")

	log.Printf("$> box pull")

	log.Printf("$> box pull (again)")

	log.Printf("$> box rm")
}
