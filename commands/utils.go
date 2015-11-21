package commands

import (
	"log"
	"os"
)

var Clog = log.New(os.Stderr, "box: ", 0)
