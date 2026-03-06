package main

import (
	"fmt"
	"os"

	"github.com/pixincreate/gito/cmd"
)

var version = "dev"

func init() {
	cmd.SetVersion(version)
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
