package main

import (
	"os"

	"github.com/authstar/spry/cli/cmds"
)

func main() {
	code := 0
	var root = cmds.Init()
	var err = root.Execute()
	if err != nil {
		// don't panic, cobra will print the error
		code = 1
	}
	os.Exit(code)
}
